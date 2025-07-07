package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test default values
	if config.SecretURL != "https://au.mitimes.com/careers/apply/secret" {
		t.Errorf("Expected default SecretURL to be 'https://au.mitimes.com/careers/apply/secret', got '%s'", config.SecretURL)
	}

	if config.ApplicationURL != "https://au.mitimes.com/careers/apply" {
		t.Errorf("Expected default ApplicationURL to be 'https://au.mitimes.com/careers/apply', got '%s'", config.ApplicationURL)
	}

	if config.Timeout != 30 {
		t.Errorf("Expected default Timeout to be 30, got %d", config.Timeout)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	configContent := `{
  "secret_url": "https://test.com/secret",
  "application_url": "https://test.com/apply",
  "timeout_seconds": 60
}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := DefaultConfig()
	err = loadConfigFromFile(configFile, config)
	if err != nil {
		t.Fatalf("loadConfigFromFile failed: %v", err)
	}

	// Verify the config was loaded correctly
	if config.SecretURL != "https://test.com/secret" {
		t.Errorf("Expected SecretURL to be 'https://test.com/secret', got '%s'", config.SecretURL)
	}

	if config.ApplicationURL != "https://test.com/apply" {
		t.Errorf("Expected ApplicationURL to be 'https://test.com/apply', got '%s'", config.ApplicationURL)
	}

	if config.Timeout != 60 {
		t.Errorf("Expected Timeout to be 60, got %d", config.Timeout)
	}
}

func TestLoadConfigFromFileErrors(t *testing.T) {
	config := DefaultConfig()

	// Test non-existent file
	err := loadConfigFromFile("/non/existent/file.json", config)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test invalid JSON
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	err = loadConfigFromFile(invalidFile, config)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "save_test.json")

	config := &Config{
		SecretURL:      "https://save.test.com/secret",
		ApplicationURL: "https://save.test.com/apply",
		Timeout:        45,
	}

	err := SaveConfig(config, configFile)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify content by loading it back
	loadedConfig := DefaultConfig()
	err = loadConfigFromFile(configFile, loadedConfig)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Compare configs
	if loadedConfig.SecretURL != config.SecretURL {
		t.Errorf("Expected SecretURL '%s', got '%s'", config.SecretURL, loadedConfig.SecretURL)
	}
	if loadedConfig.ApplicationURL != config.ApplicationURL {
		t.Errorf("Expected ApplicationURL '%s', got '%s'", config.ApplicationURL, loadedConfig.ApplicationURL)
	}
	if loadedConfig.Timeout != config.Timeout {
		t.Errorf("Expected Timeout %d, got %d", config.Timeout, loadedConfig.Timeout)
	}
}

func TestSaveConfigErrors(t *testing.T) {
	config := DefaultConfig()

	// Test invalid directory
	err := SaveConfig(config, "/invalid/directory/config.json")
	if err == nil {
		t.Error("Expected error for invalid directory, got nil")
	}
}

func TestLoadConfigWithFlags(t *testing.T) {
	// Reset flag variables for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	tests := []struct {
		name            string
		args            []string
		expectedSecret  string
		expectedApp     string
		expectedTimeout int
		shouldError     bool
	}{
		{
			name:            "default values",
			args:            []string{"micv"},
			expectedSecret:  "https://au.mitimes.com/careers/apply/secret",
			expectedApp:     "https://au.mitimes.com/careers/apply",
			expectedTimeout: 30,
			shouldError:     false,
		},
		{
			name:            "custom secret URL",
			args:            []string{"micv", "--secret-url", "https://custom.com/secret"},
			expectedSecret:  "https://custom.com/secret",
			expectedApp:     "https://au.mitimes.com/careers/apply",
			expectedTimeout: 30,
			shouldError:     false,
		},
		{
			name:            "custom app URL",
			args:            []string{"micv", "--app-url", "https://custom.com/apply"},
			expectedSecret:  "https://au.mitimes.com/careers/apply/secret",
			expectedApp:     "https://custom.com/apply",
			expectedTimeout: 30,
			shouldError:     false,
		},
		{
			name:            "custom timeout",
			args:            []string{"micv", "--timeout", "45"},
			expectedSecret:  "https://au.mitimes.com/careers/apply/secret",
			expectedApp:     "https://au.mitimes.com/careers/apply",
			expectedTimeout: 45,
			shouldError:     false,
		},
		{
			name:            "all custom values",
			args:            []string{"micv", "--secret-url", "https://test.com/secret", "--app-url", "https://test.com/apply", "--timeout", "90"},
			expectedSecret:  "https://test.com/secret",
			expectedApp:     "https://test.com/apply",
			expectedTimeout: 90,
			shouldError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag state
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Backup original os.Args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			// Redirect stderr to avoid output during tests
			flag.CommandLine.SetOutput(os.Stderr)

			configResult, err := LoadConfig()

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.shouldError && configResult != nil {
				config := configResult.Config
				if config.SecretURL != tt.expectedSecret {
					t.Errorf("Expected SecretURL '%s', got '%s'", tt.expectedSecret, config.SecretURL)
				}
				if config.ApplicationURL != tt.expectedApp {
					t.Errorf("Expected ApplicationURL '%s', got '%s'", tt.expectedApp, config.ApplicationURL)
				}
				if config.Timeout != tt.expectedTimeout {
					t.Errorf("Expected Timeout %d, got %d", tt.expectedTimeout, config.Timeout)
				}
			}
		})
	}
}

func TestLoadConfigWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	configContent := `{
  "secret_url": "https://file.test.com/secret",
  "application_url": "https://file.test.com/apply",
  "timeout_seconds": 120
}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset flag variables for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args with config file
	os.Args = []string{"micv", "--config", configFile}

	configResult, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	config := configResult.Config

	// Verify the config was loaded from file
	if config.SecretURL != "https://file.test.com/secret" {
		t.Errorf("Expected SecretURL from file, got '%s'", config.SecretURL)
	}
	if config.ApplicationURL != "https://file.test.com/apply" {
		t.Errorf("Expected ApplicationURL from file, got '%s'", config.ApplicationURL)
	}
	if config.Timeout != 120 {
		t.Errorf("Expected Timeout from file, got %d", config.Timeout)
	}
}

func TestLoadConfigFileOverrideWithFlags(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	configContent := `{
  "secret_url": "https://file.test.com/secret",
  "application_url": "https://file.test.com/apply",
  "timeout_seconds": 120
}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Reset flag variables for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args with config file and override flags
	os.Args = []string{"micv", "--config", configFile, "--secret-url", "https://override.com/secret", "--timeout", "60"}

	configResult, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	config := configResult.Config

	// Verify flags override config file values
	if config.SecretURL != "https://override.com/secret" {
		t.Errorf("Expected SecretURL to be overridden by flag, got '%s'", config.SecretURL)
	}
	if config.ApplicationURL != "https://file.test.com/apply" {
		t.Errorf("Expected ApplicationURL from file (not overridden), got '%s'", config.ApplicationURL)
	}
	if config.Timeout != 60 {
		t.Errorf("Expected Timeout to be overridden by flag, got %d", config.Timeout)
	}
}

func TestLoadConfigInvalidConfigFile(t *testing.T) {
	// Reset flag variables for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args with non-existent config file
	os.Args = []string{"micv", "--config", "/non/existent/config.json"}

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error for non-existent config file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load config file") {
		t.Errorf("Expected error message about loading config file, got: %s", err.Error())
	}
}

func TestConfigJSONMarshalUnmarshal(t *testing.T) {
	original := &Config{
		SecretURL:      "https://test.com/secret",
		ApplicationURL: "https://test.com/apply",
		Timeout:        75,
	}

	// Test JSON marshaling
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "marshal_test.json")

	err := SaveConfig(original, configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test JSON unmarshaling
	loaded := DefaultConfig()
	err = loadConfigFromFile(configFile, loaded)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Compare all fields
	if loaded.SecretURL != original.SecretURL {
		t.Errorf("SecretURL mismatch: expected '%s', got '%s'", original.SecretURL, loaded.SecretURL)
	}
	if loaded.ApplicationURL != original.ApplicationURL {
		t.Errorf("ApplicationURL mismatch: expected '%s', got '%s'", original.ApplicationURL, loaded.ApplicationURL)
	}
	if loaded.Timeout != original.Timeout {
		t.Errorf("Timeout mismatch: expected %d, got %d", original.Timeout, loaded.Timeout)
	}
}

func TestGetVersionInfo(t *testing.T) {
	version, buildTime, commitHash := GetVersionInfo()

	// In test mode, these should return the default values
	if version == "" {
		t.Error("Version should not be empty")
	}
	if buildTime == "" {
		t.Error("BuildTime should not be empty")
	}
	if commitHash == "" {
		t.Error("CommitHash should not be empty")
	}

	// Test that we get some reasonable default values
	if version != "dev" && buildTime == "unknown" && commitHash == "unknown" {
		t.Log("Build-time values detected (built with ldflags)")
	} else {
		t.Log("Development mode values detected")
	}
}

func TestLoadApplicationData(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid application data",
			filename:    "test-valid-data.json",
			expectError: false,
		},
		{
			name:        "missing name field",
			filename:    "test-invalid-data.json",
			expectError: true,
			errorMsg:    "missing required fields: name",
		},
		{
			name:        "empty fields",
			filename:    "test-empty-fields.json",
			expectError: true,
			errorMsg:    "missing required fields: name, email, job_title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use testdata files
			dataFile := filepath.Join("testdata", tt.filename)

			// Test LoadApplicationData
			appData, err := LoadApplicationData(dataFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if appData == nil {
					t.Error("Expected application data but got nil")
				} else {
					// Verify required fields are present
					if appData.Name == "" {
						t.Error("Expected name to be present")
					}
					if appData.Email == "" {
						t.Error("Expected email to be present")
					}
					if appData.JobTitle == "" {
						t.Error("Expected job_title to be present")
					}
				}
			}
		})
	}
}

func TestValidateApplicationData(t *testing.T) {
	tests := []struct {
		name        string
		appData     ApplicationData
		expectError bool
		errorMsg    string
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
			name: "missing name",
			appData: ApplicationData{
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			expectError: true,
			errorMsg:    "missing required fields: name",
		},
		{
			name: "missing email",
			appData: ApplicationData{
				Name:     "John Doe",
				JobTitle: "Software Engineer",
			},
			expectError: true,
			errorMsg:    "missing required fields: email",
		},
		{
			name: "missing job_title",
			appData: ApplicationData{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectError: true,
			errorMsg:    "missing required fields: job_title",
		},
		{
			name: "empty name",
			appData: ApplicationData{
				Name:     "",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			expectError: true,
			errorMsg:    "missing required fields: name",
		},
		{
			name: "whitespace only name",
			appData: ApplicationData{
				Name:     "   \t\n   ",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			expectError: true,
			errorMsg:    "missing required fields: name",
		},
		{
			name: "all fields missing",
			appData: ApplicationData{
				Name:     "",
				Email:    "",
				JobTitle: "",
			},
			expectError: true,
			errorMsg:    "missing required fields: name, email, job_title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApplicationData(&tt.appData)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
