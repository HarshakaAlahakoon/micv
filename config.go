package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

// Build-time variables (set via -ldflags)
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

// GetVersionInfo returns detailed version information
func GetVersionInfo() (version, buildTime, commitHash string) {
	version = Version
	buildTime = BuildTime
	commitHash = CommitHash

	// Try to get version from build info if not set at build time
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}
		}
	}

	return version, buildTime, commitHash
}

// Config holds all configuration options
type Config struct {
	SecretURL      string `json:"secret_url"`
	ApplicationURL string `json:"application_url"`
	Timeout        int    `json:"timeout_seconds"`
}

// ConfigResult holds the config and additional flags
type ConfigResult struct {
	Config   *Config
	DataFile string
	Verbose  bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		SecretURL:      "https://au.mitimes.com/careers/apply/secret",
		ApplicationURL: "https://au.mitimes.com/careers/apply",
		Timeout:        30,
	}
}

// LoadConfig loads configuration from file and command line arguments
func LoadConfig() (*ConfigResult, error) {
	config := DefaultConfig()

	// Define command line flags
	var (
		configFile         = flag.String("config", "", "Path to configuration file")
		secretURL          = flag.String("secret-url", "", "URL for the secret endpoint")
		appURL             = flag.String("app-url", "", "URL for the application endpoint")
		timeout            = flag.Int("timeout", 0, "Request timeout in seconds")
		dataFile           = flag.String("data", "", "Path to JSON file containing application data")
		generateDataJSON   = flag.Bool("generate-data-json", false, "Generate sample data.json file")
		generateConfigJSON = flag.Bool("generate-config-json", false, "Generate sample config.json file")
		verbose            = flag.Bool("verbose", false, "Enable verbose logging (debug level)")
		showHelp           = flag.Bool("help", false, "Show help message")
		showVersion        = flag.Bool("version", false, "Show version information")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [<name> <email> <job_title> [final_attempt]]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  --config string\n")
		fmt.Fprintf(os.Stderr, "        Path to configuration file\n")
		fmt.Fprintf(os.Stderr, "  --secret-url string\n")
		fmt.Fprintf(os.Stderr, "        URL for the secret endpoint\n")
		fmt.Fprintf(os.Stderr, "  --app-url string\n")
		fmt.Fprintf(os.Stderr, "        URL for the application endpoint\n")
		fmt.Fprintf(os.Stderr, "  --timeout int\n")
		fmt.Fprintf(os.Stderr, "        Request timeout in seconds\n")
		fmt.Fprintf(os.Stderr, "  --data string\n")
		fmt.Fprintf(os.Stderr, "        Path to JSON file containing application data\n")
		fmt.Fprintf(os.Stderr, "  --generate-data-json\n")
		fmt.Fprintf(os.Stderr, "        Generate sample data.json file\n")
		fmt.Fprintf(os.Stderr, "  --generate-config-json\n")
		fmt.Fprintf(os.Stderr, "        Generate sample config.json file\n")
		fmt.Fprintf(os.Stderr, "  --verbose\n")
		fmt.Fprintf(os.Stderr, "        Enable verbose logging (debug level)\n")
		fmt.Fprintf(os.Stderr, "  --version\n")
		fmt.Fprintf(os.Stderr, "        Show version information\n")
		fmt.Fprintf(os.Stderr, "  --help\n")
		fmt.Fprintf(os.Stderr, "        Show help message\n")
		fmt.Fprintf(os.Stderr, "\nArguments (when --data is not used):\n")
		fmt.Fprintf(os.Stderr, "  name           Full name of the applicant\n")
		fmt.Fprintf(os.Stderr, "  email          Email address of the applicant\n")
		fmt.Fprintf(os.Stderr, "  job_title      Job title to apply for\n")
		fmt.Fprintf(os.Stderr, "  final_attempt  Set to 'true' for final attempt (optional)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s \"John Doe\" \"john@example.com\" \"Software Engineer\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config config.json \"John Doe\" \"john@example.com\" \"Software Engineer\" true\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --secret-url https://custom.com/secret \"John Doe\" \"john@example.com\" \"Software Engineer\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --data application.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config config.json --data application.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --generate-data-json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --generate-config-json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --generate-data-json --generate-config-json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --verbose \"John Doe\" \"john@example.com\" \"Software Engineer\"\n", os.Args[0])
	}

	flag.Parse()

	// Handle generation flags
	if *generateDataJSON || *generateConfigJSON {
		return handleGenerateFiles(*generateDataJSON, *generateConfigJSON)
	}

	if *showVersion {
		version, buildTime, commitHash := GetVersionInfo()
		fmt.Fprintf(os.Stderr, "micv version %s\n", version)
		fmt.Fprintf(os.Stderr, "Built: %s\n", buildTime)
		fmt.Fprintf(os.Stderr, "Commit: %s\n", commitHash)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Load from config file if specified
	if *configFile != "" {
		if err := loadConfigFromFile(*configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with command line arguments if provided
	if *secretURL != "" {
		config.SecretURL = *secretURL
	}
	if *appURL != "" {
		config.ApplicationURL = *appURL
	}
	if *timeout > 0 {
		config.Timeout = *timeout
	}

	loadFromEnvironment(config)

	return &ConfigResult{
		Config:   config,
		DataFile: *dataFile,
		Verbose:  *verbose,
	}, nil
}

// loadConfigFromFile loads configuration from a JSON file
func loadConfigFromFile(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// LoadApplicationData loads application data from a JSON file
func LoadApplicationData(filename string) (*ApplicationData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	var appData ApplicationData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&appData); err != nil {
		return nil, fmt.Errorf("failed to decode data file: %w", err)
	}

	// Validate required fields
	if err := validateApplicationData(&appData); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &appData, nil
}

// ValidationRule represents a validation function
type ValidationRule[T any] func(T) error

// Validator provides functional validation capabilities
type Validator[T any] struct {
	rules []ValidationRule[T]
}

// NewValidator creates a new validator
func NewValidator[T any]() *Validator[T] {
	return &Validator[T]{rules: make([]ValidationRule[T], 0)}
}

// AddRule adds a validation rule
func (v *Validator[T]) AddRule(rule ValidationRule[T]) *Validator[T] {
	v.rules = append(v.rules, rule)
	return v
}

// Validate runs all validation rules
func (v *Validator[T]) Validate(value T) Result[T] {
	for _, rule := range v.rules {
		if err := rule(value); err != nil {
			return NewError[T](err)
		}
	}
	return NewResult(value)
}

// Common validation rules for application data
func RequiredField(fieldName string) ValidationRule[string] {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		return nil
	}
}

func EmailFormat() ValidationRule[string] {
	return func(value string) error {
		if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}
}

func MinLength(min int) ValidationRule[string] {
	return func(value string) error {
		if len(strings.TrimSpace(value)) < min {
			return fmt.Errorf("must be at least %d characters", min)
		}
		return nil
	}
}

// validateApplicationData validates that required fields are present and not empty
func validateApplicationData(appData *ApplicationData) error {
	var missingFields []string

	if strings.TrimSpace(appData.Name) == "" {
		missingFields = append(missingFields, "name")
	}
	if strings.TrimSpace(appData.Email) == "" {
		missingFields = append(missingFields, "email")
	}
	if strings.TrimSpace(appData.JobTitle) == "" {
		missingFields = append(missingFields, "job_title")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// validateApplicationDataFunctional provides functional validation
func validateApplicationDataFunctional(data ApplicationData) Result[ApplicationData] {
	nameValidator := NewValidator[string]().
		AddRule(RequiredField("name")).
		AddRule(MinLength(2))

	emailValidator := NewValidator[string]().
		AddRule(RequiredField("email")).
		AddRule(EmailFormat())

	jobTitleValidator := NewValidator[string]().
		AddRule(RequiredField("job_title")).
		AddRule(MinLength(3))

	// Validate all fields
	if result := nameValidator.Validate(data.Name); result.IsError() {
		return NewError[ApplicationData](result.Error)
	}

	if result := emailValidator.Validate(data.Email); result.IsError() {
		return NewError[ApplicationData](result.Error)
	}

	if result := jobTitleValidator.Validate(data.JobTitle); result.IsError() {
		return NewError[ApplicationData](result.Error)
	}

	return NewResult(data)
}

// loadFromEnvironment loads configuration from environment variables
func loadFromEnvironment(config *Config) {
	if secretURL := os.Getenv("MICV_SECRET_URL"); secretURL != "" {
		config.SecretURL = secretURL
	}

	if appURL := os.Getenv("MICV_APPLICATION_URL"); appURL != "" {
		config.ApplicationURL = appURL
	}

	if timeoutStr := os.Getenv("MICV_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			config.Timeout = timeout
		}
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config.SecretURL == "" {
		return fmt.Errorf("secret URL is required")
	}

	if config.ApplicationURL == "" {
		return fmt.Errorf("application URL is required")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// handleGenerateFiles handles generation of config and/or data files
func handleGenerateFiles(generateData, generateConfig bool) (*ConfigResult, error) {
	var generatedFiles []string

	if generateData {
		fmt.Println("🎯 Generating sample data.json file...")
		sampleData := createSampleApplicationData()
		filename := "data.json"
		if err := SaveApplicationData(sampleData, filename); err != nil {
			fmt.Printf("❌ Error generating sample data file: %v\n", err)
			os.Exit(1)
		}
		generatedFiles = append(generatedFiles, filename)
		fmt.Printf("✅ Sample data.json file generated successfully!\n")
	}

	if generateConfig {
		fmt.Println("🎯 Generating sample config.json file...")
		sampleConfig := DefaultConfig()
		filename := "config.json"
		if err := SaveConfig(sampleConfig, filename); err != nil {
			fmt.Printf("❌ Error generating sample config file: %v\n", err)
			os.Exit(1)
		}
		generatedFiles = append(generatedFiles, filename)
		fmt.Printf("✅ Sample config.json file generated successfully!\n")
	}

	// Display summary
	fmt.Printf("\n📄 Generated files:\n")
	for _, file := range generatedFiles {
		fmt.Printf("   - %s\n", file)
	}

	fmt.Printf("\n� Usage examples:\n")
	if generateData && generateConfig {
		fmt.Printf("   %s --config config.json --data data.json\n", os.Args[0])
	} else if generateData {
		fmt.Printf("   %s --data data.json\n", os.Args[0])
	} else if generateConfig {
		fmt.Printf("   %s --config config.json\n", os.Args[0])
	}

	os.Exit(0)
	return nil, nil // This line will never be reached due to os.Exit above
}

// createSampleApplicationData creates sample application data with realistic values
func createSampleApplicationData() ApplicationData {
	finalAttempt := false
	extraInfo := ExtraInfo{
		PersonalAttributes: []string{
			"Problem-solver",
			"Team player",
			"Fast learner",
			"Detail-oriented",
			"Passionate about technology",
			"Strong communication skills",
		},
		Experience: Experience{
			YearsOfExperience: 5,
			PreviousRoles: []string{
				"Senior Software Engineer",
				"Full Stack Developer",
				"Backend Developer",
			},
			KeyProjects: []string{
				"Built scalable microservices handling 1M+ requests/day",
				"Developed real-time data processing pipeline",
				"Led team of 4 developers in agile environment",
				"Migrated legacy system to cloud infrastructure",
			},
			Languages: []string{
				"Go",
				"Python",
				"JavaScript",
				"TypeScript",
				"Java",
				"SQL",
			},
			Frameworks: []string{
				"React",
				"Node.js",
				"Django",
				"Gin",
				"Echo",
				"Spring Boot",
			},
		},
		WhyHireMe: "I bring a unique combination of technical expertise and leadership skills. My experience in building scalable systems, combined with my passion for clean code and collaborative development, makes me an ideal candidate. I'm committed to continuous learning and always strive to deliver high-quality solutions that exceed expectations.",
		TechnicalSkills: []string{
			"Microservices Architecture",
			"Cloud Computing (AWS, GCP, Azure)",
			"Docker & Kubernetes",
			"CI/CD Pipelines",
			"Database Design (SQL & NoSQL)",
			"API Design & Development",
			"Test-Driven Development",
			"Agile/Scrum Methodologies",
		},
		Education:    "Bachelor's in Computer Science",
		Location:     "Australia",
		Availability: "Immediate",
	}

	return ApplicationData{
		Name:             "John Doe",
		Email:            "john.doe@example.com",
		JobTitle:         "Software Engineer",
		FinalAttempt:     &finalAttempt,
		ExtraInformation: extraInfo,
	}
}

// SaveApplicationData saves application data to a JSON file
func SaveApplicationData(data ApplicationData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create data file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}
