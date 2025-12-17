package rbkrb

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

type KrbCurlHttpClient struct {
	ProxyURL string
}

func (k *KrbCurlHttpClient) Get(url string) (*http.Response, error) {
	args := []string{
		"--proxy-negotiate", "-u", ":",
		"-L", "-s",
		"-H", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	}
	if k.ProxyURL != "" {
		args = append(args, "-x", k.ProxyURL)
	}
	args = append(args, url)

	cmd := exec.Command("curl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("curl failed: %w", err)
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(output)),
	}, nil
}

func (k *KrbCurlHttpClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	args := []string{
		"--proxy-negotiate", "-u", ":",
		"-L", "-s",
		"-X", "POST",
		"-H", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"-H", fmt.Sprintf("Content-Type: %s", contentType),
		"-d", string(bodyBytes),
	}
	if k.ProxyURL != "" {
		args = append(args, "-x", k.ProxyURL)
	}
	args = append(args, url)

	cmd := exec.Command("curl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("curl failed: %w", err)
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(output)),
	}, nil
}

func NewKrbCurlHttpClient(proxyURL string) *KrbCurlHttpClient {
	return &KrbCurlHttpClient{ProxyURL: proxyURL}
}
