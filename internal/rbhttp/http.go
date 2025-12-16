package rbhttp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type HttpBody struct {
	Command string `json:"command"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
}

type CheckInResponse struct {
	Command string `json:"command"`
}

type NewCommandResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type NewCommandRequest struct {
	Command string `json:"command"`
}

func (n *NewCommandRequest) Bind(r *http.Request) error {
	return nil
}

func (h *HttpBody) Bind(r *http.Request) error {
	return nil
}

// EncodeCommand encodes a command string as base64 for transport to the beacon.
func EncodeCommand(cmd string) string {
	return base64.StdEncoding.EncodeToString([]byte(cmd))
}

// ReadJsonBody decodes a JSON HttpBody from the request.
func ReadJsonBody(r *http.Request) (HttpBody, error) {
	defer r.Body.Close()

	var body HttpBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return HttpBody{}, err
	}

	return body, nil
}

type BeaconResponse struct {
	ID      string    `json:"id"`
	Time    time.Time `json:"time"`
	Stdout  string    `json:"stdout"`
	Stderr  string    `json:"stderr"`
	Command string    `json:"command"`
}

func NewBeaconResponse(cmd, stdout, stderr string) *BeaconResponse {
	return &BeaconResponse{
		ID:      uuid.New().String(),
		Time:    time.Now(),
		Stdout:  stdout,
		Stderr:  stderr,
		Command: cmd,
	}
}

type BeaconResponses struct {
	Responses []BeaconResponse
	sync.Mutex
}

func NewBeaconResponses() *BeaconResponses {
	return &BeaconResponses{
		Responses: make([]BeaconResponse, 0),
	}
}

func (b *BeaconResponses) Append(r BeaconResponse) {
	b.Responses = append(b.Responses, r)
}
