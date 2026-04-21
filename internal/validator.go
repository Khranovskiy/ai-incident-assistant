package internal

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// Validate checks the TriageResult against the output contract.
// Returns nil if valid, or the first validation error found.
func Validate(result *TriageResult) error {
	// Category
	if !isValidCategory(result.Category) {
		return ValidationError{Field: "category", Message: fmt.Sprintf("invalid value %q", result.Category)}
	}

	// Summary
	if strings.TrimSpace(result.Summary) == "" {
		return ValidationError{Field: "summary", Message: "must not be empty"}
	}

	// Affected
	if strings.TrimSpace(result.Affected) == "" {
		return ValidationError{Field: "affected", Message: "must not be empty"}
	}

	// Severity
	if !isValidSeverity(result.Severity) {
		return ValidationError{Field: "severity", Message: fmt.Sprintf("invalid value %q", result.Severity)}
	}

	// Hypotheses count
	if len(result.Hypotheses) == 0 {
		return ValidationError{Field: "hypotheses", Message: "must contain at least 1 hypothesis"}
	}
	if len(result.Hypotheses) > 3 {
		return ValidationError{Field: "hypotheses", Message: fmt.Sprintf("must contain at most 3 hypotheses, got %d", len(result.Hypotheses))}
	}

	// Each hypothesis
	for i, h := range result.Hypotheses {
		if strings.TrimSpace(h.Title) == "" {
			return ValidationError{Field: fmt.Sprintf("hypotheses[%d].title", i), Message: "must not be empty"}
		}
		if strings.TrimSpace(h.Reasoning) == "" {
			return ValidationError{Field: fmt.Sprintf("hypotheses[%d].reasoning", i), Message: "must not be empty"}
		}
		if len(h.NextSteps) < 2 || len(h.NextSteps) > 3 {
			return ValidationError{
				Field:   fmt.Sprintf("hypotheses[%d].next_steps", i),
				Message: fmt.Sprintf("must contain 2-3 items, got %d", len(h.NextSteps)),
			}
		}
		for j, step := range h.NextSteps {
			if strings.TrimSpace(step) == "" {
				return ValidationError{
					Field:   fmt.Sprintf("hypotheses[%d].next_steps[%d]", i, j),
					Message: "must not be empty",
				}
			}
		}
	}

	// Language check: verify the summary is predominantly ASCII/Latin (English heuristic)
	if !looksEnglish(result.Summary) {
		return ValidationError{Field: "summary", Message: "does not appear to be in English"}
	}

	return nil
}

// ValidateRawJSON checks that the raw string is valid JSON and can be parsed
// into a TriageResult. Returns the parsed result or an error.
func ValidateRawJSON(raw string) (*TriageResult, error) {
	var result TriageResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("invalid JSON: %v", err)}
	}
	if err := Validate(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func isValidCategory(c Category) bool {
	for _, v := range ValidCategories {
		if c == v {
			return true
		}
	}
	return false
}

func isValidSeverity(s Severity) bool {
	for _, v := range ValidSeverities {
		if s == v {
			return true
		}
	}
	return false
}

// looksEnglish returns true if the text is predominantly Latin characters.
func looksEnglish(text string) bool {
	if text == "" {
		return false
	}
	latinCount := 0
	totalLetters := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			totalLetters++
			if r < 0x0250 { // Basic Latin + Latin Extended-A/B
				latinCount++
			}
		}
	}
	if totalLetters == 0 {
		return false
	}
	return float64(latinCount)/float64(totalLetters) > 0.8
}
