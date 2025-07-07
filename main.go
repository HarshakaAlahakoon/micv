package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	// Load configuration
	configResult, err := LoadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	config := configResult.Config

	// Determine log level based on verbose flag
	logLevel := LogLevelInfo
	if configResult.Verbose {
		logLevel = LogLevelDebug
	}

	// Initialize dependencies
	deps := NewAppDependencies(config, logLevel)
	logger := deps.Logger()

	// Create application instance
	app := NewApplication(deps)

	// Load application data
	appData, err := loadApplicationData(configResult)
	if err != nil {
		logger.Error("Failed to load application data", "error", err)
		fmt.Printf("❌ Error loading application data: %v\n", err)
		os.Exit(1)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Timeout+10)*time.Second)
	defer cancel()

	// Run application
	if err := app.Run(ctx, appData); err != nil {
		logger.Error("Application execution failed", "error", err)

		// Enhanced error reporting for users
		if appErr, ok := err.(*AppError); ok {
			fmt.Printf("❌ %s: %s\n", appErr.Code, appErr.Message)
			if appErr.Cause != nil {
				fmt.Printf("   Cause: %v\n", appErr.Cause)
			}
			for key, value := range appErr.Context {
				fmt.Printf("   %s: %v\n", key, value)
			}
		} else {
			fmt.Printf("❌ Error: %v\n", err)
		}
		os.Exit(1)
	}
}

// getAuthTokenWithClient fetches auth token using the provided HTTP client (testable version)
func getAuthTokenWithClient(client HTTPClient, secretURL string) (string, error) {
	// Make request to secret endpoint
	resp, err := client.Get(secretURL)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Validate response status
	if err := validateSecretResponse(resp); err != nil {
		return "", err
	}

	// Read and parse response
	return parseSecretResponse(resp)
}

// validateSecretResponse validates the HTTP response from secret endpoint
func validateSecretResponse(resp *http.Response) error {
	fmt.Printf("🌐 Secret endpoint HTTP Status: %d %s\n", resp.StatusCode, resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("secret endpoint returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// parseSecretResponse parses the secret endpoint response to extract token
func parseSecretResponse(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("📄 Secret endpoint response body: %s\n", string(body))

	// Parse JSON response
	var secretResp SecretResponse
	if err := json.Unmarshal(body, &secretResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if secretResp.Result == "" {
		return "", fmt.Errorf("empty result in secret response")
	}

	return secretResp.Result, nil
}

// submitApplicationWithClient submits application using the provided HTTP client (testable version)
func submitApplicationWithClient(client HTTPClient, applicationURL string, token string, appData ApplicationData) error {
	// Prepare JSON data
	jsonData, err := prepareApplicationJSON(appData)
	if err != nil {
		return err
	}

	// Create and send request
	req, err := createApplicationRequest(applicationURL, token, jsonData)
	if err != nil {
		return err
	}

	// Execute request and handle response
	return executeApplicationRequest(client, req)
}

// prepareApplicationJSON converts application data to JSON
func prepareApplicationJSON(appData ApplicationData) ([]byte, error) {
	jsonData, err := json.MarshalIndent(appData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Printf("📋 Application data being sent:\n%s\n", string(jsonData))
	return jsonData, nil
}

// createApplicationRequest creates HTTP request for application submission
func createApplicationRequest(applicationURL string, token string, jsonData []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", applicationURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	return req, nil
}

// executeApplicationRequest executes the application request and handles response
func executeApplicationRequest(client HTTPClient, req *http.Request) error {
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read and process response
	return processApplicationResponse(resp)
}

// processApplicationResponse processes the application submission response
func processApplicationResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Print results
	fmt.Printf("🎯 Application submission HTTP Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("📄 Application submission response body: %s\n", string(body))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("✅ Application submitted successfully!")
	} else {
		fmt.Println("⚠️  Application submission completed with non-success status")
	}

	return nil
}

// loadApplicationData loads application data from file or command line arguments
func loadApplicationData(configResult *ConfigResult) (ApplicationData, error) {
	var appData ApplicationData

	// Check remaining command line arguments (after flags)
	args := flag.Args()

	// Validate that both --data flag and command line arguments are not provided together
	if configResult.DataFile != "" && len(args) > 0 {
		fmt.Println("❌ Error: Cannot use both --data flag and command line arguments together")
		fmt.Println("💡 Please use either:")
		fmt.Println("   - The --data flag to specify a JSON file: --data applicant-data.json")
		fmt.Println("   - Command line arguments: \"Name\" \"email@example.com\" \"Job Title\"")
		fmt.Println("   - Use --help for more information")
		os.Exit(1)
	}

	if configResult.DataFile != "" {
		// Load application data from JSON file
		fmt.Printf("📖 Loading application data from: %s\n", configResult.DataFile)
		loadedData, err := LoadApplicationData(configResult.DataFile)
		if err != nil {
			return appData, err
		}
		appData = *loadedData
		fmt.Println("✅ Application data loaded successfully from file")
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

// createDefaultApplicationData creates application data with default extra information
func createDefaultApplicationData(name, email, jobTitle string, finalAttempt *bool) ApplicationData {
	extraInfo := createDefaultExtraInfo()

	return ApplicationData{
		Name:             name,
		Email:            email,
		JobTitle:         jobTitle,
		FinalAttempt:     finalAttempt,
		ExtraInformation: extraInfo,
	}
}

// createDefaultExtraInfo creates default extra information for the application
func createDefaultExtraInfo() ExtraInfo {
	return ExtraInfo{
		PersonalAttributes: getDefaultPersonalAttributes(),
		Experience:         createDefaultExperience(),
		WhyHireMe:          getDefaultWhyHireMe(),
		TechnicalSkills:    getDefaultTechnicalSkills(),
		Education:          "Bachelor's in Computer Science",
		Location:           "Australia",
		Availability:       "Immediate",
	}
}

// getDefaultPersonalAttributes returns default personal attributes
func getDefaultPersonalAttributes() []string {
	return []string{
		"Problem-solver",
		"Team player",
		"Fast learner",
		"Detail-oriented",
		"Passionate about technology",
	}
}

// createDefaultExperience creates default experience information
func createDefaultExperience() Experience {
	return Experience{
		YearsOfExperience: 5,
		PreviousRoles:     getDefaultPreviousRoles(),
		KeyProjects:       getDefaultKeyProjects(),
		Languages:         getDefaultLanguages(),
		Frameworks:        getDefaultFrameworks(),
	}
}

// getDefaultPreviousRoles returns default previous roles
func getDefaultPreviousRoles() []string {
	return []string{
		"Senior Software Engineer",
		"Full Stack Developer",
		"Backend Developer",
	}
}

// getDefaultKeyProjects returns default key projects
func getDefaultKeyProjects() []string {
	return []string{
		"Built scalable microservices handling 1M+ requests/day",
		"Developed real-time data processing pipeline",
		"Led team of 4 developers in agile environment",
	}
}

// getDefaultLanguages returns default programming languages
func getDefaultLanguages() []string {
	return []string{"Go", "Python", "JavaScript", "TypeScript", "Java"}
}

// getDefaultFrameworks returns default frameworks
func getDefaultFrameworks() []string {
	return []string{"React", "Node.js", "Django", "Gin", "Echo"}
}

// getDefaultWhyHireMe returns default "why hire me" text
func getDefaultWhyHireMe() string {
	return "I bring a unique combination of technical expertise and leadership skills. My experience in building scalable systems, combined with my passion for clean code and collaborative development, makes me an ideal candidate. I'm committed to continuous learning and always strive to deliver high-quality solutions that exceed expectations."
}

// getDefaultTechnicalSkills returns default technical skills
func getDefaultTechnicalSkills() []string {
	return []string{
		"Microservices Architecture",
		"Cloud Computing (AWS, GCP)",
		"Docker & Kubernetes",
		"CI/CD Pipelines",
		"Database Design (SQL & NoSQL)",
		"API Design & Development",
		"Test-Driven Development",
	}
}
