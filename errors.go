package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides structured logging capabilities
type Logger struct {
	logger *slog.Logger
	level  LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel) *Logger {
	var slogLevel slog.Level
	switch level {
	case LogLevelDebug:
		slogLevel = slog.LevelDebug
	case LogLevelInfo:
		slogLevel = slog.LevelInfo
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{
		logger: logger,
		level:  level,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Debug(msg, fields...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Info(msg, fields...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Warn(msg, fields...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Error(msg, fields...)
	}
}

// With adds fields to the logger context
func (l *Logger) With(fields ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(fields...),
		level:  l.level,
	}
}

// AppError represents application-specific errors with context
type AppError struct {
	Code      string
	Message   string
	Cause     error
	Context   map[string]interface{}
	Timestamp time.Time
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(code, message string, cause error) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// Error codes for better error categorization
const (
	ErrCodeNetwork     = "NETWORK_ERROR"
	ErrCodeValidation  = "VALIDATION_ERROR"
	ErrCodeConfig      = "CONFIG_ERROR"
	ErrCodeAuth        = "AUTH_ERROR"
	ErrCodeApplication = "APPLICATION_ERROR"
	ErrCodeParsing     = "PARSING_ERROR"
	ErrCodeTimeout     = "TIMEOUT_ERROR"
	ErrCodeUnexpected  = "UNEXPECTED_ERROR"
)

// Enhanced error handling functions
func WrapNetworkError(err error, url string) *AppError {
	return NewAppError(ErrCodeNetwork, "Network request failed", err).
		WithContext("url", url).
		WithContext("retry_suggested", true)
}

func WrapValidationError(err error, field string) *AppError {
	return NewAppError(ErrCodeValidation, "Validation failed", err).
		WithContext("field", field).
		WithContext("user_action_required", true)
}

func WrapConfigError(err error, configPath string) *AppError {
	return NewAppError(ErrCodeConfig, "Configuration error", err).
		WithContext("config_path", configPath).
		WithContext("check_config_file", true)
}

func WrapAuthError(err error, endpoint string) *AppError {
	return NewAppError(ErrCodeAuth, "Authentication failed", err).
		WithContext("endpoint", endpoint).
		WithContext("check_credentials", true)
}

// Circuit breaker pattern for resilient HTTP calls
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailTime time.Time
	state        CircuitState
	logger       *Logger
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger *Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
		logger:       logger,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open state")
		} else {
			cb.logger.Warn("Circuit breaker is open, rejecting call")
			return NewAppError(ErrCodeTimeout, "Circuit breaker is open", nil)
		}
	}

	err := fn()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
		cb.logger.Error("Circuit breaker opened due to failures",
			"failures", cb.failures,
			"max_failures", cb.maxFailures)
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	cb.state = CircuitClosed
	if cb.state == CircuitHalfOpen {
		cb.logger.Info("Circuit breaker closed after successful call")
	}
}

// Retry mechanism with exponential backoff
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, config RetryConfig, logger *Logger, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		logger.Debug("Attempting operation",
			"attempt", attempt,
			"max_attempts", config.MaxAttempts)

		err := fn()
		if err == nil {
			if attempt > 1 {
				logger.Info("Operation succeeded after retry",
					"successful_attempt", attempt)
			}
			return nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			logger.Error("All retry attempts exhausted",
				"attempts", attempt,
				"last_error", err)
			break
		}

		logger.Warn("Operation failed, retrying",
			"attempt", attempt,
			"delay", delay,
			"error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return NewAppError(ErrCodeUnexpected, "Operation failed after all retries", lastErr)
}

// Pipeline represents a functional pipeline of operations
type Pipeline[T any] struct {
	operations []func(T) Result[T]
}

// NewPipeline creates a new pipeline
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{operations: make([]func(T) Result[T], 0)}
}

// Add adds an operation to the pipeline
func (p *Pipeline[T]) Add(op func(T) Result[T]) *Pipeline[T] {
	p.operations = append(p.operations, op)
	return p
}

// Execute runs all operations in the pipeline
func (p *Pipeline[T]) Execute(input T) Result[T] {
	result := NewResult(input)
	for _, op := range p.operations {
		if result.IsError() {
			return result
		}
		result = op(result.Value)
	}
	return result
}
