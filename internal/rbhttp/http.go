package rbhttp

import (
	"encoding/base64"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

func ReadPlaintextBody(r *http.Request) (string, error) {
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(bodyBytes))
	if err != nil {
		return "", err
	}

	return string(decoded), nil
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
