package validator

import (
	"regexp"
)

// Validator holds validation errors
type Validator struct {
	Errors map[string]string
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message for a given key
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message if validation check is not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// In returns true if a value is in a list
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Matches returns true if a value matches a regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique returns true if all values in a slice are unique
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
