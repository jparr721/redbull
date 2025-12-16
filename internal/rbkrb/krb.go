package rbkrb

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

type ProxyAwareHttpClient struct {
	proxyURL   *url.URL
	httpClient *http.Client
	krbClient  *client.Client
	krbConfig  *config.Config
	sync.Mutex
}

func NewProxyAwareHttpClient(proxyURL, password string) (*ProxyAwareHttpClient, error) {
	pURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	c := &ProxyAwareHttpClient{
		proxyURL: pURL,
	}
	if err := c.init(password); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *ProxyAwareHttpClient) init(password string) error {
	cfg, err := loadKrb5Config()
	if err != nil {
		return fmt.Errorf("failed to load krb5 config: %w", err)
	}
	c.krbConfig = cfg

	ccachePath, err := discoverCCache()
	if err != nil {
		// No ccache found, try kinit
		if err := runKinit(password); err != nil {
			return fmt.Errorf("no credentials cache and kinit failed: %w", err)
		}
		ccachePath, err = discoverCCache()
		if err != nil {
			return fmt.Errorf("kinit succeeded but ccache still not found: %w", err)
		}
	}

	ccache, err := credentials.LoadCCache(ccachePath)
	if err != nil {
		// ccache exists but may be expired/corrupt, try kinit
		if err := runKinit(password); err != nil {
			return fmt.Errorf("failed to load ccache and kinit failed: %w", err)
		}
		ccache, err = credentials.LoadCCache(ccachePath)
		if err != nil {
			return fmt.Errorf("failed to load ccache after kinit: %w", err)
		}
	}

	c.krbClient, err = client.NewFromCCache(ccache, cfg)
	if err != nil {
		return fmt.Errorf("failed to create kerberos client: %w", err)
	}

	c.httpClient = &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyURL(c.proxyURL)},
	}
	return nil
}

func (c *ProxyAwareHttpClient) getNegotiateToken() (string, error) {
	spn := "HTTP/" + c.proxyURL.Hostname()
	s := spnego.SPNEGOClient(c.krbClient, spn)

	if err := s.AcquireCred(); err != nil {
		return "", fmt.Errorf("failed to acquire credentials: %w", err)
	}

	token, err := s.InitSecContext()
	if err != nil {
		return "", fmt.Errorf("failed to init security context: %w", err)
	}

	tokenBytes, err := token.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	return base64.StdEncoding.EncodeToString(tokenBytes), nil
}

func (c *ProxyAwareHttpClient) MakeReq(req *http.Request) (*http.Response, error) {
	c.Lock()
	token, err := c.getNegotiateToken()
	c.Unlock()
	if err != nil {
		return nil, err
	}

	req2 := req.Clone(req.Context())
	req2.Header.Set("Proxy-Authorization", "Negotiate "+token)

	resp, err := c.httpClient.Do(req2)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusProxyAuthRequired {
		resp.Body.Close()
		return nil, fmt.Errorf("proxy authentication failed")
	}

	return resp, nil
}

func (c *ProxyAwareHttpClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.MakeReq(req)
}

// Post performs an HTTP POST request.
func (c *ProxyAwareHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.MakeReq(req)
}

func runKinit(password string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("kinit", "--keychain")
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	var output []byte
	done := make(chan error)
	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	fmt.Fprintf(stdin, "%s\n", password)
	stdin.Close()

	if err := <-done; err != nil {
		return fmt.Errorf("kinit failed: %w\noutput: %s", err, string(output))
	}

	return nil
}

func loadKrb5Config() (*config.Config, error) {
	if envPath := os.Getenv("KRB5_CONFIG"); envPath != "" {
		if cfg, err := config.Load(envPath); err == nil {
			return cfg, nil
		}
	}
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		candidates = []string{
			filepath.Join(os.Getenv("PROGRAMDATA"), "MIT", "Kerberos5", "krb5.ini"),
			filepath.Join(os.Getenv("WINDIR"), "krb5.ini"),
		}
	case "darwin":
		candidates = []string{
			"/etc/krb5.conf",
			"/Library/Preferences/edu.mit.Kerberos",
			"/usr/local/etc/krb5.conf",
		}
	}

	for _, path := range candidates {
		if cfg, err := config.Load(path); err == nil {
			return cfg, nil
		}
	}
	return nil, fmt.Errorf("no krb5 configuration found")
}

// discoverCCache finds the Kerberos credentials cache for the current user.
func discoverCCache() (string, error) {
	// Check environment variable first
	if envPath := os.Getenv("KRB5CCNAME"); envPath != "" {
		path := envPath
		if len(path) > 5 && path[:5] == "FILE:" {
			path = path[5:]
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		uid := os.Getuid()
		user := os.Getenv("USER")
		candidates = []string{
			fmt.Sprintf("/tmp/krb5cc_%d", uid),
			fmt.Sprintf("/tmp/krb5cc_%s", user),
		}
		if user != "" {
			candidates = append(candidates, fmt.Sprintf("/var/db/krb5cc/krb5cc_%s", user))
		}

	case "windows":
		temp := os.Getenv("TEMP")
		user := os.Getenv("USERNAME")
		profile := os.Getenv("USERPROFILE")
		if temp != "" && user != "" {
			candidates = append(candidates, filepath.Join(temp, "krb5cc_"+user))
		}
		if profile != "" {
			candidates = append(candidates, filepath.Join(profile, "krb5cc"))
		}
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no credentials cache found")
}
