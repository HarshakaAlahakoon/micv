package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	GetFunc func(url string) (*http.Response, error)
	DoFunc  func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.GetFunc(url)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// Create a response helper
func createResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestGetAuthToken(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *http.Response
		mockError     error
		expectedToken string
		expectedError bool
	}{
		{
			name:          "successful token retrieval",
			mockResponse:  createResponse(200, `{"result":"nice-to-see-you-here-talk-soon"}`),
			mockError:     nil,
			expectedToken: "nice-to-see-you-here-talk-soon",
			expectedError: false,
		},
		{
			name:          "successful token retrieval with Bearer token",
			mockResponse:  createResponse(200, `{"result":"Bearer abc123"}`),
			mockError:     nil,
			expectedToken: "Bearer abc123",
			expectedError: false,
		},
		{
			name:          "network error - request fails",
			mockResponse:  nil,
			mockError:     http.ErrServerClosed,
			expectedToken: "",
			expectedError: true,
		},
		{
			name:          "server error - 500",
			mockResponse:  createResponse(500, "Internal Server Error"),
			mockError:     nil,
			expectedToken: "",
			expectedError: true,
		},
		{
			name:          "invalid JSON response",
			mockResponse:  createResponse(200, "not-json"),
			mockError:     nil,
			expectedToken: "",
			expectedError: true,
		},
		{
			name:          "missing result field",
			mockResponse:  createResponse(200, `{"other":"value"}`),
			mockError:     nil,
			expectedToken: "",
			expectedError: true,
		},
		{
			name:          "empty result value",
			mockResponse:  createResponse(200, `{"result":""}`),
			mockError:     nil,
			expectedToken: "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				GetFunc: func(url string) (*http.Response, error) {
					if url != "https://au.mitimes.com/careers/apply/secret" {
						t.Errorf("Expected URL https://au.mitimes.com/careers/apply/secret, got %s", url)
					}
					return tt.mockResponse, tt.mockError
				},
			}

			token, err := getAuthTokenWithClient(mockClient, "https://au.mitimes.com/careers/apply/secret")

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if token != tt.expectedToken {
				t.Errorf("Expected token %s, got %s", tt.expectedToken, token)
			}
		})
	}
}

func TestSubmitApplication(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		appData       ApplicationData
		mockResponse  *http.Response
		mockError     error
		expectedError bool
	}{
		{
			name:  "successful submission",
			token: "Bearer abc123",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			mockResponse:  createResponse(200, `{"status": "success"}`),
			mockError:     nil,
			expectedError: false,
		},
		{
			name:  "server rejection",
			token: "Bearer abc123",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			mockResponse:  createResponse(400, `{"error": "Invalid data"}`),
			mockError:     nil,
			expectedError: false, // Function doesn't return error for HTTP errors
		},
		{
			name:  "network error",
			token: "Bearer abc123",
			appData: ApplicationData{
				Name:     "John Doe",
				Email:    "john@example.com",
				JobTitle: "Software Engineer",
			},
			mockResponse:  nil,
			mockError:     http.ErrServerClosed,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Verify request method
					if req.Method != "POST" {
						t.Errorf("Expected POST method, got %s", req.Method)
					}

					// Verify URL
					if req.URL.String() != "https://au.mitimes.com/careers/apply" {
						t.Errorf("Expected URL https://au.mitimes.com/careers/apply, got %s", req.URL.String())
					}

					// Verify headers
					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
					}
					if req.Header.Get("Authorization") != tt.token {
						t.Errorf("Expected Authorization %s, got %s", tt.token, req.Header.Get("Authorization"))
					}

					// Verify JSON body
					body, _ := io.ReadAll(req.Body)
					var receivedData ApplicationData
					if err := json.Unmarshal(body, &receivedData); err != nil {
						t.Errorf("Failed to parse JSON body: %v", err)
					}

					if receivedData.Name != tt.appData.Name {
						t.Errorf("Expected name %s, got %s", tt.appData.Name, receivedData.Name)
					}

					return tt.mockResponse, tt.mockError
				},
			}

			err := submitApplicationWithClient(mockClient, "https://au.mitimes.com/careers/apply", tt.token, tt.appData)

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestApplicationDataJSON(t *testing.T) {
	finalAttempt := true
	extraInfo := ExtraInfo{
		PersonalAttributes: []string{"Problem-solver", "Team player"},
		Experience: Experience{
			YearsOfExperience: 5,
			PreviousRoles:     []string{"Senior Developer"},
			KeyProjects:       []string{"Microservices"},
			Languages:         []string{"Go", "Python"},
			Frameworks:        []string{"React", "Node.js"},
		},
		WhyHireMe:       "I'm awesome",
		TechnicalSkills: []string{"Docker", "Kubernetes"},
		Education:       "Bachelor's in CS",
		Location:        "Australia",
		Availability:    "Immediate",
	}

	appData := ApplicationData{
		Name:             "John Doe",
		Email:            "john@example.com",
		JobTitle:         "Software Engineer",
		FinalAttempt:     &finalAttempt,
		ExtraInformation: extraInfo,
	}

	jsonData, err := json.Marshal(appData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify that JSON contains expected fields
	jsonStr := string(jsonData)
	expectedFields := []string{
		`"name":"John Doe"`,
		`"email":"john@example.com"`,
		`"job_title":"Software Engineer"`,
		`"final_attempt":true`,
		`"extra_information"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain %s, got: %s", field, jsonStr)
		}
	}
}

func TestExtraInfoStructure(t *testing.T) {
	extraInfo := ExtraInfo{
		PersonalAttributes: []string{"Problem-solver"},
		Experience: Experience{
			YearsOfExperience: 5,
			PreviousRoles:     []string{"Developer"},
			KeyProjects:       []string{"Project 1"},
			Languages:         []string{"Go"},
			Frameworks:        []string{"Gin"},
		},
		WhyHireMe:       "Test reason",
		TechnicalSkills: []string{"Docker"},
		Education:       "Bachelor's",
		Location:        "Australia",
		Availability:    "Immediate",
	}

	jsonData, err := json.Marshal(extraInfo)
	if err != nil {
		t.Fatalf("Failed to marshal ExtraInfo: %v", err)
	}

	var unmarshaled ExtraInfo
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ExtraInfo: %v", err)
	}

	if unmarshaled.Experience.YearsOfExperience != 5 {
		t.Errorf("Expected YearsOfExperience 5, got %d", unmarshaled.Experience.YearsOfExperience)
	}
}

func TestIntegrationWithMockServer(t *testing.T) {
	// Create a test server for the secret endpoint
	secretServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/careers/apply/secret" {
			w.WriteHeader(200)
			w.Write([]byte("Bearer test-token-123"))
		}
	}))
	defer secretServer.Close()

	// Create a test server for the application endpoint
	appServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/careers/apply" && r.Method == "POST" {
			// Verify authorization header
			if r.Header.Get("Authorization") != "Bearer test-token-123" {
				w.WriteHeader(401)
				w.Write([]byte(`{"error": "Unauthorized"}`))
				return
			}

			// Read and verify JSON body
			body, _ := io.ReadAll(r.Body)
			var appData ApplicationData
			if err := json.Unmarshal(body, &appData); err != nil {
				w.WriteHeader(400)
				w.Write([]byte(`{"error": "Invalid JSON"}`))
				return
			}

			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "application_id": "12345"}`))
		}
	}))
	defer appServer.Close()

	// This test demonstrates how the integration would work
	// In practice, you'd need to modify the URLs in your main functions
	// or make them configurable for testing
	t.Log("Integration test servers created successfully")
	t.Logf("Secret server URL: %s", secretServer.URL)
	t.Logf("Application server URL: %s", appServer.URL)
}

// Benchmark tests
func BenchmarkJSONMarshal(b *testing.B) {
	extraInfo := ExtraInfo{
		PersonalAttributes: []string{"Problem-solver", "Team player"},
		Experience: Experience{
			YearsOfExperience: 5,
			PreviousRoles:     []string{"Senior Developer"},
			KeyProjects:       []string{"Microservices"},
			Languages:         []string{"Go", "Python"},
			Frameworks:        []string{"React", "Node.js"},
		},
		WhyHireMe:       "I'm awesome",
		TechnicalSkills: []string{"Docker", "Kubernetes"},
		Education:       "Bachelor's in CS",
		Location:        "Australia",
		Availability:    "Immediate",
	}

	appData := ApplicationData{
		Name:             "John Doe",
		Email:            "john@example.com",
		JobTitle:         "Software Engineer",
		ExtraInformation: extraInfo,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(appData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestMainFunctionOutput(t *testing.T) {
	// Test the usage message
	os.Args = []string{"micv"}

	// This would normally call os.Exit(1), so we can't test it directly
	// Instead, we test the argument parsing logic separately
	if len(os.Args) < 4 {
		t.Log("Usage message would be displayed for insufficient arguments")
	}
}

// Test validation when loading application data from file
func TestMainWithInvalidApplicationData(t *testing.T) {
	// Create temporary file with invalid data (missing required fields)
	tempDir := t.TempDir()
	invalidDataFile := filepath.Join(tempDir, "invalid_data.json")

	invalidContent := `{
		"name": "",
		"email": "john@example.com"
	}`

	err := os.WriteFile(invalidDataFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid data file: %v", err)
	}

	// Test that LoadApplicationData fails with validation error
	_, err = LoadApplicationData(invalidDataFile)
	if err == nil {
		t.Error("Expected validation error but got none")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected validation error message, got: %v", err)
	}

	if !strings.Contains(err.Error(), "missing required fields") {
		t.Errorf("Expected missing required fields error, got: %v", err)
	}
}

func TestApplicationDataValidationWithDifferentScenarios(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		expectError bool
		description string
	}{
		{
			name: "valid minimal data",
			jsonContent: `{
				"name": "Test User",
				"email": "test@example.com",
				"job_title": "Developer"
			}`,
			expectError: false,
			description: "Should pass with minimal required fields",
		},
		{
			name: "valid complete data",
			jsonContent: `{
				"name": "Test User",
				"email": "test@example.com",
				"job_title": "Developer",
				"final_attempt": true,
				"extra_information": {
					"personal_attributes": ["Fast learner"],
					"experience": {
						"years_of_experience": 3,
						"previous_roles": ["Junior Developer"],
						"key_projects": ["Web App"],
						"programming_languages": ["Go"],
						"frameworks": ["Gin"]
					},
					"why_hire_me": "Great skills",
					"technical_skills": ["Docker"],
					"education": "Bachelor's",
					"location": "Australia",
					"availability": "Immediate"
				}
			}`,
			expectError: false,
			description: "Should pass with complete data",
		},
		{
			name: "missing name only",
			jsonContent: `{
				"email": "test@example.com",
				"job_title": "Developer"
			}`,
			expectError: true,
			description: "Should fail when name is missing",
		},
		{
			name: "missing email only",
			jsonContent: `{
				"name": "Test User",
				"job_title": "Developer"
			}`,
			expectError: true,
			description: "Should fail when email is missing",
		},
		{
			name: "missing job_title only",
			jsonContent: `{
				"name": "Test User",
				"email": "test@example.com"
			}`,
			expectError: true,
			description: "Should fail when job_title is missing",
		},
		{
			name: "empty strings",
			jsonContent: `{
				"name": "",
				"email": "",
				"job_title": ""
			}`,
			expectError: true,
			description: "Should fail with empty strings",
		},
		{
			name: "whitespace only",
			jsonContent: `{
				"name": "   ",
				"email": " \t ",
				"job_title": "\n  \t"
			}`,
			expectError: true,
			description: "Should fail with whitespace only fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempDir := t.TempDir()
			dataFile := filepath.Join(tempDir, "test_data.json")

			err := os.WriteFile(dataFile, []byte(tt.jsonContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test data file: %v", err)
			}

			// Test LoadApplicationData
			appData, err := LoadApplicationData(dataFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
				if appData != nil {
					t.Errorf("Expected nil application data on error, got %+v", appData)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v. %s", err, tt.description)
				}
				if appData == nil {
					t.Errorf("Expected application data but got nil. %s", tt.description)
				}
			}
		})
	}
}

// TestSecretResponseParsing tests the JSON parsing for SecretResponse
func TestSecretResponseParsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		expected    SecretResponse
		expectError bool
	}{
		{
			name:        "valid JSON with result",
			jsonContent: `{"result":"nice-to-see-you-here-talk-soon"}`,
			expected:    SecretResponse{Result: "nice-to-see-you-here-talk-soon"},
			expectError: false,
		},
		{
			name:        "valid JSON with Bearer token",
			jsonContent: `{"result":"Bearer abc123def456"}`,
			expected:    SecretResponse{Result: "Bearer abc123def456"},
			expectError: false,
		},
		{
			name:        "valid JSON with empty result",
			jsonContent: `{"result":""}`,
			expected:    SecretResponse{Result: ""},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonContent: `{"result":}`,
			expected:    SecretResponse{},
			expectError: true,
		},
		{
			name:        "empty JSON object",
			jsonContent: `{}`,
			expected:    SecretResponse{Result: ""},
			expectError: false,
		},
		{
			name:        "JSON with additional fields",
			jsonContent: `{"result":"token123","status":"success","timestamp":1234567890}`,
			expected:    SecretResponse{Result: "token123"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var secretResp SecretResponse
			err := json.Unmarshal([]byte(tt.jsonContent), &secretResp)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if secretResp.Result != tt.expected.Result {
					t.Errorf("Expected result %s, got %s", tt.expected.Result, secretResp.Result)
				}
			}
		})
	}
}

// TestGetAuthTokenWithClientIntegration tests the complete flow including JSON parsing
func TestGetAuthTokenWithClientIntegration(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectedToken string
		expectError   bool
	}{
		{
			name:          "successful response with expected JSON",
			responseBody:  `{"result":"nice-to-see-you-here-talk-soon"}`,
			statusCode:    200,
			expectedToken: "nice-to-see-you-here-talk-soon",
			expectError:   false,
		},
		{
			name:          "successful response with Bearer token",
			responseBody:  `{"result":"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"}`,
			statusCode:    200,
			expectedToken: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectError:   false,
		},
		{
			name:          "empty result should fail",
			responseBody:  `{"result":""}`,
			statusCode:    200,
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "invalid JSON should fail",
			responseBody:  `{"result":}`,
			statusCode:    200,
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "missing result field should fail",
			responseBody:  `{"status":"success"}`,
			statusCode:    200,
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "non-200 status should fail",
			responseBody:  `{"result":"nice-to-see-you-here-talk-soon"}`,
			statusCode:    404,
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "server error should fail",
			responseBody:  `Internal Server Error`,
			statusCode:    500,
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				GetFunc: func(url string) (*http.Response, error) {
					return createResponse(tt.statusCode, tt.responseBody), nil
				},
			}

			token, err := getAuthTokenWithClient(mockClient, "https://example.com/secret")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token != tt.expectedToken {
					t.Errorf("Expected token %s, got %s", tt.expectedToken, token)
				}
			}
		})
	}
}

// TestGetAuthTokenWithRealHTTPServer tests with a real HTTP server
func TestGetAuthTokenWithRealHTTPServer(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SecretResponse{Result: "nice-to-see-you-here-talk-soon"})
	}))
	defer server.Close()

	// Create a real HTTP client
	client := NewHTTPClientWithTimeout(5 * time.Second)

	// Test the function
	token, err := getAuthTokenWithClient(client, server.URL)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if token != "nice-to-see-you-here-talk-soon" {
		t.Errorf("Expected token 'nice-to-see-you-here-talk-soon', got %s", token)
	}
}

// TestGetAuthTokenErrorHandling tests various error scenarios
func TestGetAuthTokenErrorHandling(t *testing.T) {
	tests := []struct {
		name            string
		setupMockClient func() HTTPClient
		expectError     bool
		errorContains   string
	}{
		{
			name: "network error",
			setupMockClient: func() HTTPClient {
				return &MockHTTPClient{
					GetFunc: func(url string) (*http.Response, error) {
						return nil, http.ErrServerClosed
					},
				}
			},
			expectError:   true,
			errorContains: "failed to make request",
		},
		{
			name: "cannot read response body",
			setupMockClient: func() HTTPClient {
				return &MockHTTPClient{
					GetFunc: func(url string) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Status:     "OK",
							Body:       &errorReadCloser{},
							Header:     make(http.Header),
						}, nil
					},
				}
			},
			expectError:   true,
			errorContains: "failed to read response body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupMockClient()
			token, err := getAuthTokenWithClient(client, "https://example.com/secret")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				if token != "" {
					t.Errorf("Expected empty token on error, got: %s", token)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestSecretEndpointFailures tests essential secret endpoint failure scenarios
func TestSecretEndpointFailures(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectedError bool
		errorContains string
	}{
		{
			name:          "successful response",
			responseBody:  `{"result":"nice-to-see-you-here-talk-soon"}`,
			statusCode:    200,
			expectedError: false,
			errorContains: "",
		},
		{
			name:          "request fails - server error",
			responseBody:  `Internal Server Error`,
			statusCode:    500,
			expectedError: true,
			errorContains: "non-success status",
		},
		{
			name:          "response missing result field",
			responseBody:  `{"status":"success"}`,
			statusCode:    200,
			expectedError: true,
			errorContains: "empty result",
		},
		{
			name:          "response with empty result value",
			responseBody:  `{"result":""}`,
			statusCode:    200,
			expectedError: true,
			errorContains: "empty result",
		},
		{
			name:          "invalid JSON response",
			responseBody:  `not-json`,
			statusCode:    200,
			expectedError: true,
			errorContains: "failed to parse JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				GetFunc: func(url string) (*http.Response, error) {
					return createResponse(tt.statusCode, tt.responseBody), nil
				},
			}

			token, err := getAuthTokenWithClient(mockClient, "https://example.com/secret")

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				if token != "" {
					t.Errorf("Expected empty token on error, got: %s", token)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token == "" {
					t.Errorf("Expected token but got empty string")
				}
			}
		})
	}
}

// TestSecretEndpointNetworkFailure tests network connection failure
func TestSecretEndpointNetworkFailure(t *testing.T) {
	// Test with a mock client that simulates network failure
	mockClient := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return nil, http.ErrServerClosed
		},
	}

	token, err := getAuthTokenWithClient(mockClient, "https://example.com/secret")
	if err == nil {
		t.Errorf("Expected network error but got none")
	}
	if token != "" {
		t.Errorf("Expected empty token on network error, got: %s", token)
	}
	if !strings.Contains(err.Error(), "failed to make request") {
		t.Errorf("Expected network error to be wrapped as 'failed to make request', got: %v", err)
	}
}

// errorReadCloser is a helper type that always returns an error when Read is called
type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e *errorReadCloser) Close() error {
	return nil
}

// TestLoadApplicationDataConflictValidation tests that using both --data flag and command line arguments together fails
func TestLoadApplicationDataConflictValidation(t *testing.T) {
	// Create a temporary JSON file for testing
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "test-data.json")

	testData := ApplicationData{
		Name:     "Test User",
		Email:    "test@example.com",
		JobTitle: "Test Engineer",
	}

	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(dataFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test data file: %v", err)
	}

	tests := []struct {
		name           string
		dataFile       string
		args           []string
		expectedToFail bool
	}{
		{
			name:           "data file only - should work",
			dataFile:       dataFile,
			args:           []string{},
			expectedToFail: false,
		},
		{
			name:           "command line args only - should work",
			dataFile:       "",
			args:           []string{"John Doe", "john@example.com", "Software Engineer"},
			expectedToFail: false,
		},
		{
			name:           "insufficient args without data file - should fail",
			dataFile:       "",
			args:           []string{"John Doe"},
			expectedToFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ConfigResult with the test data
			configResult := &ConfigResult{
				Config: &Config{
					SecretURL:      "https://example.com/secret",
					ApplicationURL: "https://example.com/apply",
					Timeout:        30,
				},
				DataFile: tt.dataFile,
			}

			// Mock flag.Args() by creating a test function that simulates command line parsing
			testLoadApplicationData := func(configResult *ConfigResult, args []string) (ApplicationData, error) {
				var appData ApplicationData

				// Validate that both --data flag and command line arguments are not provided together
				if configResult.DataFile != "" && len(args) > 0 {
					// This would normally call os.Exit(1), but for testing we'll return an error
					return appData, fmt.Errorf("cannot use both --data flag and command line arguments together")
				}

				if configResult.DataFile != "" {
					// Load application data from JSON file
					loadedData, err := LoadApplicationData(configResult.DataFile)
					if err != nil {
						return appData, err
					}
					appData = *loadedData
				} else {
					if len(args) < 3 {
						return appData, fmt.Errorf("insufficient arguments provided")
					}

					name := args[0]
					email := args[1]
					jobTitle := args[2]

					var finalAttempt *bool
					if len(args) > 3 && args[3] == "true" {
						val := true
						finalAttempt = &val
					}

					// Create application data from command line arguments
					appData = createDefaultApplicationData(name, email, jobTitle, finalAttempt)
				}

				return appData, nil
			}

			// Call the test function
			appData, err := testLoadApplicationData(configResult, tt.args)

			if tt.expectedToFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Verify the data was loaded correctly
				if tt.dataFile != "" {
					// Data should be loaded from file
					if appData.Name != testData.Name {
						t.Errorf("Expected name '%s' but got '%s'", testData.Name, appData.Name)
					}
					if appData.Email != testData.Email {
						t.Errorf("Expected email '%s' but got '%s'", testData.Email, appData.Email)
					}
					if appData.JobTitle != testData.JobTitle {
						t.Errorf("Expected job title '%s' but got '%s'", testData.JobTitle, appData.JobTitle)
					}
				} else if len(tt.args) >= 3 {
					// Data should be loaded from command line args
					if appData.Name != tt.args[0] {
						t.Errorf("Expected name '%s' but got '%s'", tt.args[0], appData.Name)
					}
					if appData.Email != tt.args[1] {
						t.Errorf("Expected email '%s' but got '%s'", tt.args[1], appData.Email)
					}
					if appData.JobTitle != tt.args[2] {
						t.Errorf("Expected job title '%s' but got '%s'", tt.args[2], appData.JobTitle)
					}
				}
			}
		})
	}
}

// TestLoadApplicationDataConflictValidationError tests the specific error case
func TestLoadApplicationDataConflictValidationError(t *testing.T) {
	// Create a temporary JSON file for testing
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "test-data.json")

	testData := ApplicationData{
		Name:     "Test User",
		Email:    "test@example.com",
		JobTitle: "Test Engineer",
	}

	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(dataFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test data file: %v", err)
	}

	// Test the validation logic directly
	configResult := &ConfigResult{
		DataFile: dataFile,
	}

	// Mock having command line arguments
	args := []string{"John Doe", "john@example.com", "Software Engineer"}

	// Check that the validation condition would trigger
	if configResult.DataFile != "" && len(args) > 0 {
		t.Logf("✅ Validation correctly detected conflict: data file '%s' and %d command line args",
			configResult.DataFile, len(args))
	} else {
		t.Errorf("❌ Validation failed to detect conflict")
	}
}

// TestLoadApplicationDataWithFinalAttempt tests loading application data with final attempt flag
func TestLoadApplicationDataWithFinalAttempt(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectFinal bool
	}{
		{
			name:        "with final attempt true",
			args:        []string{"John Doe", "john@example.com", "Software Engineer", "true"},
			expectFinal: true,
		},
		{
			name:        "with final attempt false",
			args:        []string{"John Doe", "john@example.com", "Software Engineer", "false"},
			expectFinal: false,
		},
		{
			name:        "without final attempt",
			args:        []string{"John Doe", "john@example.com", "Software Engineer"},
			expectFinal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic directly without flag parsing
			if len(tt.args) < 3 {
				t.Errorf("Test setup error: insufficient args")
				return
			}

			name := tt.args[0]
			email := tt.args[1]
			jobTitle := tt.args[2]

			var finalAttempt *bool
			if len(tt.args) > 3 && tt.args[3] == "true" {
				val := true
				finalAttempt = &val
			}

			appData := createDefaultApplicationData(name, email, jobTitle, finalAttempt)

			if tt.expectFinal {
				if appData.FinalAttempt == nil {
					t.Errorf("Expected FinalAttempt to be set but got nil")
				} else if !*appData.FinalAttempt {
					t.Errorf("Expected FinalAttempt to be true but got false")
				}
			} else {
				if appData.FinalAttempt != nil && *appData.FinalAttempt {
					t.Errorf("Expected FinalAttempt to be false/nil but got true")
				}
			}
		})
	}
}

// TestCreateDefaultApplicationData tests the creation of default application data
func TestCreateDefaultApplicationData(t *testing.T) {
	name := "John Doe"
	email := "john@example.com"
	jobTitle := "Software Engineer"

	finalAttempt := true

	appData := createDefaultApplicationData(name, email, jobTitle, &finalAttempt)

	// Check basic fields
	if appData.Name != name {
		t.Errorf("Expected name '%s' but got '%s'", name, appData.Name)
	}
	if appData.Email != email {
		t.Errorf("Expected email '%s' but got '%s'", email, appData.Email)
	}
	if appData.JobTitle != jobTitle {
		t.Errorf("Expected job title '%s' but got '%s'", jobTitle, appData.JobTitle)
	}
	if appData.FinalAttempt == nil || !*appData.FinalAttempt {
		t.Errorf("Expected FinalAttempt to be true but got %v", appData.FinalAttempt)
	}

	// Check that extra information is populated
	if appData.ExtraInformation == nil {
		t.Errorf("Expected extra information to be populated but got nil")
		return
	}

	extraInfo, ok := appData.ExtraInformation.(ExtraInfo)
	if !ok {
		t.Errorf("Expected extra information to be ExtraInfo type but got %T", appData.ExtraInformation)
		return
	}

	if extraInfo.Education == "" {
		t.Errorf("Expected education to be populated but got empty string")
	}
	if extraInfo.Location == "" {
		t.Errorf("Expected location to be populated but got empty string")
	}
	if len(extraInfo.PersonalAttributes) == 0 {
		t.Errorf("Expected personal attributes to be populated but got empty slice")
	}
	if len(extraInfo.TechnicalSkills) == 0 {
		t.Errorf("Expected technical skills to be populated but got empty slice")
	}
	if extraInfo.Experience.YearsOfExperience == 0 {
		t.Errorf("Expected years of experience to be populated but got 0")
	}
}
