package main

import (
	"context"
	"time"
)

// Dependencies interface defines all external dependencies
type Dependencies interface {
	HTTPClient() HTTPClient
	Logger() *Logger
	Config() *Config
	CircuitBreaker() *CircuitBreaker
}

// AppDependencies implements Dependencies interface
type AppDependencies struct {
	httpClient     HTTPClient
	logger         *Logger
	config         *Config
	circuitBreaker *CircuitBreaker
}

// HTTPClient returns the HTTP client
func (d *AppDependencies) HTTPClient() HTTPClient {
	return d.httpClient
}

// Logger returns the logger
func (d *AppDependencies) Logger() *Logger {
	return d.logger
}

// Config returns the configuration
func (d *AppDependencies) Config() *Config {
	return d.config
}

// CircuitBreaker returns the circuit breaker
func (d *AppDependencies) CircuitBreaker() *CircuitBreaker {
	return d.circuitBreaker
}

// NewAppDependencies creates a new dependencies container
func NewAppDependencies(config *Config, logLevel LogLevel) *AppDependencies {
	logger := NewLogger(logLevel)
	httpClient := NewHTTPClientWithTimeout(time.Duration(config.Timeout) * time.Second)
	circuitBreaker := NewCircuitBreaker(3, 30*time.Second, logger)

	return &AppDependencies{
		httpClient:     httpClient,
		logger:         logger,
		config:         config,
		circuitBreaker: circuitBreaker,
	}
}

// ApplicationService provides high-level application operations
type ApplicationService struct {
	deps Dependencies
}

// NewApplicationService creates a new application service
func NewApplicationService(deps Dependencies) *ApplicationService {
	return &ApplicationService{
		deps: deps,
	}
}

// SubmitApplication handles the complete application submission process
func (s *ApplicationService) SubmitApplication(ctx context.Context, appData ApplicationData) error {
	logger := s.deps.Logger().With("operation", "submit_application")

	logger.Debug("Starting application submission",
		"name", appData.Name,
		"email", appData.Email,
		"job_title", appData.JobTitle)

	// Validate application data
	if err := s.validateApplication(appData); err != nil {
		logger.Error("Application validation failed", "error", err)
		return WrapValidationError(err, "application_data")
	}

	// Fetch authorization token with circuit breaker protection
	token, err := s.fetchTokenWithResilience(ctx)
	if err != nil {
		logger.Error("Failed to fetch authorization token", "error", err)
		return err
	}

	// Submit application with retry mechanism
	if err := s.submitWithResilience(ctx, token, appData); err != nil {
		logger.Error("Failed to submit application", "error", err)
		return err
	}

	logger.Debug("Application submitted successfully")
	return nil
}

// validateApplication validates the application data
func (s *ApplicationService) validateApplication(appData ApplicationData) error {
	result := validateApplicationDataFunctional(appData)
	if result.IsError() {
		return result.Error
	}
	return nil
}

// fetchTokenWithResilience fetches auth token with circuit breaker protection
func (s *ApplicationService) fetchTokenWithResilience(ctx context.Context) (string, error) {
	logger := s.deps.Logger().With("operation", "fetch_token")

	var token string
	var err error

	circuitBreakerErr := s.deps.CircuitBreaker().Call(ctx, func() error {
		token, err = s.fetchTokenWithRetry(ctx)
		return err
	})

	if circuitBreakerErr != nil {
		logger.Error("Circuit breaker rejected token fetch", "error", circuitBreakerErr)
		return "", circuitBreakerErr
	}

	return token, nil
}

// fetchTokenWithRetry fetches auth token with retry logic
func (s *ApplicationService) fetchTokenWithRetry(ctx context.Context) (string, error) {
	logger := s.deps.Logger().With("operation", "fetch_token_retry")

	var token string

	err := WithRetry(ctx, DefaultRetryConfig(), logger, func() error {
		var fetchErr error
		token, fetchErr = getAuthTokenWithClient(s.deps.HTTPClient(), s.deps.Config().SecretURL)
		if fetchErr != nil {
			logger.Debug("Token fetch attempt failed", "error", fetchErr)
			return WrapAuthError(fetchErr, s.deps.Config().SecretURL)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return token, nil
}

// submitWithResilience submits application with retry mechanism
func (s *ApplicationService) submitWithResilience(ctx context.Context, token string, appData ApplicationData) error {
	logger := s.deps.Logger().With("operation", "submit_with_resilience")

	return WithRetry(ctx, DefaultRetryConfig(), logger, func() error {
		err := submitApplicationWithClient(
			s.deps.HTTPClient(),
			s.deps.Config().ApplicationURL,
			token,
			appData,
		)

		if err != nil {
			logger.Debug("Application submission attempt failed", "error", err)
			return WrapNetworkError(err, s.deps.Config().ApplicationURL)
		}

		return nil
	})
}

// AuthTokenService handles token-related operations
type AuthTokenService struct {
	deps Dependencies
}

// NewAuthTokenService creates a new auth token service
func NewAuthTokenService(deps Dependencies) *AuthTokenService {
	return &AuthTokenService{
		deps: deps,
	}
}

// GetToken fetches an authentication token
func (s *AuthTokenService) GetToken(ctx context.Context) (string, error) {
	logger := s.deps.Logger().With("service", "auth_token")

	logger.Debug("Fetching authentication token",
		"endpoint", s.deps.Config().SecretURL)

	token, err := getAuthTokenWithClient(s.deps.HTTPClient(), s.deps.Config().SecretURL)
	if err != nil {
		logger.Error("Failed to fetch token", "error", err)
		return "", WrapAuthError(err, s.deps.Config().SecretURL)
	}

	logger.Debug("Authentication token fetched successfully")
	return token, nil
}

// ConfigService handles configuration operations
type ConfigService struct {
	deps Dependencies
}

// NewConfigService creates a new config service
func NewConfigService(deps Dependencies) *ConfigService {
	return &ConfigService{
		deps: deps,
	}
}

// ValidateConfig validates the current configuration
func (s *ConfigService) ValidateConfig() error {
	logger := s.deps.Logger().With("service", "config")
	config := s.deps.Config()

	if config.SecretURL == "" {
		return WrapConfigError(
			NewAppError(ErrCodeConfig, "secret URL is required", nil),
			"secret_url",
		)
	}

	if config.ApplicationURL == "" {
		return WrapConfigError(
			NewAppError(ErrCodeConfig, "application URL is required", nil),
			"application_url",
		)
	}

	if config.Timeout <= 0 {
		return WrapConfigError(
			NewAppError(ErrCodeConfig, "timeout must be positive", nil),
			"timeout",
		)
	}

	logger.Debug("Configuration validation successful")
	return nil
}

// Application represents the main application
type Application struct {
	deps          Dependencies
	appService    *ApplicationService
	authService   *AuthTokenService
	configService *ConfigService
}

// NewApplication creates a new application instance
func NewApplication(deps Dependencies) *Application {
	return &Application{
		deps:          deps,
		appService:    NewApplicationService(deps),
		authService:   NewAuthTokenService(deps),
		configService: NewConfigService(deps),
	}
}

// Run executes the main application logic
func (app *Application) Run(ctx context.Context, appData ApplicationData) error {
	logger := app.deps.Logger().With("component", "application")

	// Validate configuration
	if err := app.configService.ValidateConfig(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		return err
	}

	// Submit application
	if err := app.appService.SubmitApplication(ctx, appData); err != nil {
		logger.Error("Application submission failed", "error", err)
		return err
	}

	logger.Debug("Application execution completed successfully")
	return nil
}
