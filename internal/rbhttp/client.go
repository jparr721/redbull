package rbhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HttpClient interface {
	Get(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
}

type SimpleHttpClient struct {
}

func (d *SimpleHttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func (d *SimpleHttpClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	return http.Post(url, contentType, body)
}

func NewSimpleHttpClient() *SimpleHttpClient {
	return &SimpleHttpClient{}
}

// Generic helper functions that work with any HttpClient

func Get[T any](client HttpClient, url string) (*T, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func Post[T any](client HttpClient, url string, body any) (*T, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle 204 No Content (empty response body)
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
