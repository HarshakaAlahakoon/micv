package main

import "fmt"

// SecretResponse represents the JSON structure returned by the secret endpoint
type SecretResponse struct {
	Result string `json:"result"`
}

// ApplicationData represents the JSON structure to be sent
type ApplicationData struct {
	Name             string      `json:"name"`
	Email            string      `json:"email"`
	JobTitle         string      `json:"job_title"`
	FinalAttempt     *bool       `json:"final_attempt,omitempty"`
	ExtraInformation interface{} `json:"extra_information,omitempty"`
}

// ExtraInfo represents additional information about the candidate
type ExtraInfo struct {
	PersonalAttributes []string   `json:"personal_attributes"`
	Experience         Experience `json:"experience"`
	WhyHireMe          string     `json:"why_hire_me"`
	TechnicalSkills    []string   `json:"technical_skills"`
	Education          string     `json:"education"`
	Location           string     `json:"location"`
	Availability       string     `json:"availability"`
}

// Experience represents professional experience information
type Experience struct {
	YearsOfExperience int      `json:"years_of_experience"`
	PreviousRoles     []string `json:"previous_roles"`
	KeyProjects       []string `json:"key_projects"`
	Languages         []string `json:"programming_languages"`
	Frameworks        []string `json:"frameworks"`
}

// Result represents a functional result type for better error handling
type Result[T any] struct {
	Value T
	Error error
}

// NewResult creates a new Result with a value
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

// NewError creates a new Result with an error
func NewError[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Error: err}
}

// IsSuccess checks if the result is successful
func (r Result[T]) IsSuccess() bool {
	return r.Error == nil
}

// IsError checks if the result contains an error
func (r Result[T]) IsError() bool {
	return r.Error != nil
}

// Map applies a function to the result value if successful
func (r Result[T]) Map(fn func(T) T) Result[T] {
	if r.IsError() {
		return r
	}
	return NewResult(fn(r.Value))
}

// FlatMap applies a function that returns a Result to the result value if successful
func (r Result[T]) FlatMap(fn func(T) Result[T]) Result[T] {
	if r.IsError() {
		return r
	}
	return fn(r.Value)
}

// Filter applies a predicate to the result value
func (r Result[T]) Filter(predicate func(T) bool, errorMsg string) Result[T] {
	if r.IsError() {
		return r
	}
	if !predicate(r.Value) {
		return NewError[T](fmt.Errorf("%s", errorMsg))
	}
	return r
}

// OrElse returns the result if successful, otherwise returns the alternative
func (r Result[T]) OrElse(alternative T) T {
	if r.IsError() {
		return alternative
	}
	return r.Value
}

// Functional helpers for common operations
func Compose[A, B, C any](f func(B) C, g func(A) B) func(A) C {
	return func(a A) C {
		return f(g(a))
	}
}

func Curry[A, B, C any](f func(A, B) C) func(A) func(B) C {
	return func(a A) func(B) C {
		return func(b B) C {
			return f(a, b)
		}
	}
}
