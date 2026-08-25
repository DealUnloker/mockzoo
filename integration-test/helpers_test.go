package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

const (
	// Base settings.
	host     = "app"
	attempts = 20

	// Attempts connection.
	httpURL        = "http://" + host + ":8080"
	healthPath     = httpURL + "/healthz"
	requestTimeout = 5 * time.Second

	// HTTP REST.
	basePathV1 = httpURL + "/v1"
)

var errHealthCheck = fmt.Errorf("url %s is not available", healthPath)

// doWebRequestWithTimeout sends an HTTP request with a Content-Type of
// application/json. The timeout lives on the client, not on a per-call
// context: it covers the whole exchange including reading resp.Body, and a
// context canceled when the helper returns would kill exactly that read.
func doWebRequestWithTimeout(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: requestTimeout}

	return client.Do(req)
}

// parseJSON is a generic JSON parser for HTTP responses.
func parseJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()

	var result T

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("parseJSON: failed to decode response: %v", err)
	}

	return result
}

func doGet(t *testing.T, url string) *http.Response {
	t.Helper()

	resp, err := doWebRequestWithTimeout(t.Context(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}

	return resp
}

func doPost(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()

	resp, err := doWebRequestWithTimeout(t.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}

	return resp
}

func doPatch(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()

	resp, err := doWebRequestWithTimeout(t.Context(), http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("PATCH %s: %v", url, err)
	}

	return resp
}

func doDelete(t *testing.T, url string) *http.Response {
	t.Helper()

	resp, err := doWebRequestWithTimeout(t.Context(), http.MethodDelete, url, http.NoBody)
	if err != nil {
		t.Fatalf("DELETE %s: %v", url, err)
	}

	return resp
}

func getHealthCheck(url string) (int, error) {
	resp, err := doWebRequestWithTimeout(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func healthCheck(attempts int) error {
	for attempts > 0 {
		statusCode, err := getHealthCheck(healthPath)
		if err == nil && statusCode == http.StatusOK {
			return nil
		}

		if err != nil {
			log.Printf("Integration tests: url %s is not available: %v, attempts left: %d", healthPath, err, attempts)
		} else {
			log.Printf("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)
		}

		time.Sleep(time.Second)

		attempts--
	}

	return errHealthCheck
}

func TestMain(m *testing.M) {
	err := healthCheck(attempts)
	if err != nil {
		log.Fatalf("Integration tests: httpURL %s is not available: %s", httpURL, err)
	}

	log.Printf("Integration tests: httpURL %s is available", httpURL)

	code := m.Run()
	os.Exit(code)
}
