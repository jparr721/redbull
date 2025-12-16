package rbhttp

import (
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
