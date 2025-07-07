package main

import (
	"net/http"
	"time"
)

// HTTPClient interface to allow mocking
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// MiClient wraps the standard http.Client to implement HTTPClient interface
type MiClient struct {
	client *http.Client
}

func (m *MiClient) Get(url string) (*http.Response, error) {
	return m.client.Get(url)
}

func (m *MiClient) Do(req *http.Request) (*http.Response, error) {
	return m.client.Do(req)
}

// NewHTTPClient creates a new HTTP client with default timeout
func NewHTTPClient() HTTPClient {
	return NewHTTPClientWithTimeout(30 * time.Second)
}

// NewHTTPClientWithTimeout creates a new HTTP client with specified timeout
func NewHTTPClientWithTimeout(timeout time.Duration) HTTPClient {
	return &MiClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}
