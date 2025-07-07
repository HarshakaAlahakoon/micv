# MICV Documentation

## Table of Contents

- [Command-Line Options](#command-line-options)
  - [Basic Flags](#basic-flags)
  - [Configuration Flags](#configuration-flags)
  - [Data Management Flags](#data-management-flags)
  - [Usage Examples](#usage-examples)
    - [Verbose Mode](#verbose-mode)
    - [Generate Sample Data File](#generate-sample-data-file)
    - [Generate Sample Configuration File](#generate-sample-configuration-file)
    - [Generate Both Configuration and Data Files](#generate-both-configuration-and-data-files)
    - [Combined Flags](#combined-flags)
    - [Configuration File with Data File](#configuration-file-with-data-file)
  - [Understanding Verbose Mode](#understanding-verbose-mode)
  - [File Generation Features](#file-generation-features)
    - [Configuration File Generation](#configuration-file-generation)
    - [Data File Generation](#data-file-generation)
- [Configuration](#configuration)
  - [Configuration Hierarchy](#configuration-hierarchy-highest-to-lowest-priority)
  - [Environment Variables](#environment-variables)
  - [Configuration File Example](#configuration-file-example)

## Command-Line Options

The application supports various command-line flags for flexible configuration and operation:

### Basic Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--verbose` | boolean | Enable verbose logging (debug level) | `--verbose` |
| `--help` | boolean | Show help message and usage information | `--help` |
| `--version` | boolean | Display version, build time, and commit hash | `--version` |

### Configuration Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--config` | string | Path to configuration file | `--config config.json` |
| `--secret-url` | string | URL for the secret endpoint | `--secret-url https://custom.com/secret` |
| `--app-url` | string | URL for the application endpoint | `--app-url https://custom.com/apply` |
| `--timeout` | int | Request timeout in seconds | `--timeout 60` |

### Data Management Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--data` | string | Path to JSON file containing application data | `--data application.json` |
| `--generate-data-json` | boolean | Generate sample data.json file and exit | `--generate-data-json` |
| `--generate-config-json` | boolean | Generate sample config.json file and exit | `--generate-config-json` |

### Usage Examples

#### Verbose Mode
```bash
# Enable detailed debug logging
./micv --verbose "John Doe" "john@example.com" "Software Engineer"
```

#### Generate Sample Data File
```bash
# Generate a sample data.json file with realistic examples
./micv --generate-data-json
# This creates a 'data.json' file you can edit and use with:
./micv --data data.json
```

#### Generate Sample Configuration File
```bash
# Generate a sample config.json file with default settings
./micv --generate-config-json
# This creates a 'config.json' file you can edit and use with:
./micv --config config.json
```

#### Generate Both Configuration and Data Files
```bash
# Generate both config.json and data.json files at once
./micv --generate-config-json --generate-data-json
# This creates both files which you can then use together:
./micv --config config.json --data data.json
```

#### Combined Flags
```bash
# Use multiple flags together
./micv --verbose --timeout 60 --config custom-config.json "John Doe" "john@example.com" "Software Engineer"

# Use custom endpoints with verbose logging
./micv --verbose --secret-url https://staging.com/secret --app-url https://staging.com/apply "John Doe" "john@example.com" "Software Engineer"
```

#### Configuration File with Data File
```bash
# Use both configuration and data files
./micv --config production.json --data application.json --verbose
```

### Understanding Verbose Mode

When `--verbose` is enabled, the application provides detailed debug information including:

- HTTP request/response details
- Configuration loading steps
- Authentication token retrieval process
- Network timeout and retry information
- Detailed error context and stack traces
- Performance metrics and timing information

Example verbose output:
```
🔧 Debug: Loading configuration from file: config.json
🔧 Debug: Overriding secret URL from command line
🔧 Debug: Timeout set to 45 seconds
🌐 Debug: Making request to secret endpoint
🔧 Debug: Auth token received (length: 64 characters)
🌐 Debug: Submitting application with token
✅ Application submitted successfully
```

### File Generation Features

The application provides convenient commands to generate sample configuration and data files:

#### Configuration File Generation
```bash
./micv --generate-config-json
```
This generates a `config.json` file with default settings:
```json
{
  "secret_url": "https://au.mitimes.com/careers/apply/secret",
  "application_url": "https://au.mitimes.com/careers/apply", 
  "timeout_seconds": 30
}
```

#### Data File Generation
```bash
./micv --generate-data-json
```
This generates a `data.json` file with sample application data including:
- Personal information (name, email, job title)
- Professional attributes and skills
- Work experience and previous roles
- Key projects and achievements
- Technical skills and programming languages
- Availability and preferences

Both generated files can be edited with your actual information and used with the `--config` and `--data` flags respectively.

## Configuration

### Configuration Hierarchy (highest to lowest priority)

1. **Command Line Flags** (highest priority)
2. **Environment Variables**
3. **Configuration File**
4. **Default Values** (lowest priority)

### Environment Variables

```bash
export MICV_SECRET_URL="https://au.mitimes.com/careers/apply/secret"
export MICV_APPLICATION_URL="https://au.mitimes.com/careers/apply"
export MICV_TIMEOUT="30"
```

### Configuration File Example

```json
{
  "secret_url": "https://au.mitimes.com/careers/apply/secret",
  "application_url": "https://au.mitimes.com/careers/apply",
  "timeout_seconds": 30
}
```