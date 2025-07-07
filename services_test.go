package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// TestApplicationService tests the application service
func TestApplicationService(t *testing.T) {
	tests := []struct {
		name        string
		appData     ApplicationData
		setupMocks  func(*MockDependencies)
		expectError bool
		errorCode   string
	}{
		{
			name: "successful submission",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			setupMocks: func(deps *MockDependencies) {
				deps.httpClient.GetFunc = func(url string) (*http.Response, error) {
					return createResponse(200, `{"result":"token123"}`), nil
				}
				deps.httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
					return createResponse(200, `{"status":"success"}`), nil
				}
			},
			expectError: false,
		},
		{
			name: "validation error",
			appData: ApplicationData{
				Name:     "",
				Email:    "invalid-email",
				JobTitle: "SE",
			},
			setupMocks: func(deps *MockDependencies) {
				// No setup needed for validation error
			},
			expectError: true,
			errorCode:   ErrCodeValidation,
		},
		{
			name: "network error",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			setupMocks: func(deps *MockDependencies) {
				deps.httpClient.GetFunc = func(url string) (*http.Response, error) {
					return nil, errors.New("network error")
				}
			},
			expectError: true,
			errorCode:   ErrCodeUnexpected, // Will be wrapped in retry mechanism
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := NewMockDependencies()
			tt.setupMocks(deps)

			service := NewApplicationService(deps)
			ctx := context.Background()

			err := service.SubmitApplication(ctx, tt.appData)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if appErr, ok := err.(*AppError); ok {
					if appErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, appErr.Code)
					}
				} else {
					t.Error("Expected AppError but got different type")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestCircuitBreaker tests the circuit breaker functionality
func TestCircuitBreaker(t *testing.T) {
	logger := NewLogger(LogLevelError) // Reduce log noise during tests
	cb := NewCircuitBreaker(2, 100*time.Millisecond, logger)

	ctx := context.Background()

	// Test successful calls
	err := cb.Call(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error for successful call, got: %v", err)
	}

	// Test failure threshold
	for i := 0; i < 2; i++ {
		cb.Call(ctx, func() error {
			return errors.New("test error")
		})
	}

	// Circuit should be open now
	err = cb.Call(ctx, func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected circuit breaker to reject call when open")
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit should allow calls again
	err = cb.Call(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected circuit breaker to allow call after reset, got: %v", err)
	}
}

// TestRetryMechanism tests the retry functionality
func TestRetryMechanism(t *testing.T) {
	logger := NewLogger(LogLevelError) // Reduce log noise during tests
	ctx := context.Background()

	attempts := 0
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	// Test successful retry after failures
	err := WithRetry(ctx, config, logger, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after retries, got: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}

	// Test failure after all retries
	attempts = 0
	err = WithRetry(ctx, config, logger, func() error {
		attempts++
		return errors.New("persistent error")
	})

	if err == nil {
		t.Error("Expected error after all retries exhausted")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
}

// TestFunctionalValidation tests the functional validation
func TestFunctionalValidation(t *testing.T) {
	tests := []struct {
		name        string
		appData     ApplicationData
		expectError bool
	}{
		{
			name: "valid data",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			expectError: false,
		},
		{
			name: "invalid name",
			appData: ApplicationData{
				Name:     "",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			expectError: true,
		},
		{
			name: "invalid email",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "invalid-email",
				JobTitle: "Software Engineer",
			},
			expectError: true,
		},
		{
			name: "invalid job title",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "SE",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateApplicationDataFunctional(tt.appData)

			if tt.expectError && result.IsSuccess() {
				t.Error("Expected validation error but validation succeeded")
			}

			if !tt.expectError && result.IsError() {
				t.Errorf("Expected validation success but got error: %v", result.Error)
			}
		})
	}
}

// MockDependencies implements Dependencies for testing
type MockDependencies struct {
	httpClient     *MockHTTPClient
	logger         *Logger
	config         *Config
	circuitBreaker *CircuitBreaker
}

func NewMockDependencies() *MockDependencies {
	config := DefaultConfig()
	logger := NewLogger(LogLevelError) // Reduce log noise during tests

	return &MockDependencies{
		httpClient:     &MockHTTPClient{},
		logger:         logger,
		config:         config,
		circuitBreaker: NewCircuitBreaker(3, 30*time.Second, logger),
	}
}

func (m *MockDependencies) HTTPClient() HTTPClient {
	return m.httpClient
}

func (m *MockDependencies) Logger() *Logger {
	return m.logger
}

func (m *MockDependencies) Config() *Config {
	return m.config
}

func (m *MockDependencies) CircuitBreaker() *CircuitBreaker {
	return m.circuitBreaker
}

// TestApplication tests the main application flow
func TestApplication(t *testing.T) {
	deps := NewMockDependencies()
	deps.httpClient.GetFunc = func(url string) (*http.Response, error) {
		return createResponse(200, `{"result":"token123"}`), nil
	}
	deps.httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return createResponse(200, `{"status":"success"}`), nil
	}

	app := NewApplication(deps)
	ctx := context.Background()

	appData := ApplicationData{
		Name:     "John Doe",
		Email:    "john@example.com",
		JobTitle: "Software Engineer",
	}

	err := app.Run(ctx, appData)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
}

// Benchmark tests for performance
func BenchmarkApplicationService(b *testing.B) {
	deps := NewMockDependencies()
	deps.httpClient.GetFunc = func(url string) (*http.Response, error) {
		return createResponse(200, `{"result":"token123"}`), nil
	}
	deps.httpClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return createResponse(200, `{"status":"success"}`), nil
	}

	service := NewApplicationService(deps)
	ctx := context.Background()

	appData := ApplicationData{
		Name:     "John Doe",
		Email:    "john@example.com",
		JobTitle: "Software Engineer",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SubmitApplication(ctx, appData)
	}
}
