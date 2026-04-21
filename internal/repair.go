package internal

import (
	"regexp"
	"strings"
)

var jsonFenceRe = regexp.MustCompile(`(?s)` + "```" + `(?:json)?\s*\n?(\{.*?\})\s*\n?` + "```")
var jsonObjectRe = regexp.MustCompile(`(?s)(\{.*\})`)

// categoryAliases maps common LLM-generated category variants to canonical values.
var categoryAliases = map[string]Category{
	"external payment provider issue":          CategoryExternalPaymentProvider,
	"db degradation caused by reporting":       CategoryDBDegradation,
	"database degradation caused by reporting": CategoryDBDegradation,
	"notification delivery issue":              CategoryNotificationDelivery,
	"user authentication errors":               CategoryUserAuthentication,
	"authentication errors":                    CategoryUserAuthentication,
	"unknown or mixed":                         CategoryUnknown,
	"unknown":                                  CategoryUnknown,
}

// severityAliases maps common case variants to canonical severity values.
var severityAliases = map[string]Severity{
	"low":    SeverityLow,
	"medium": SeverityMedium,
	"high":   SeverityHigh,
}

// RepairJSON attempts to extract and fix near-valid JSON from LLM output.
// Returns the cleaned JSON string and true if repair succeeded, or empty string and false.
func RepairJSON(raw string) (string, bool) {
	cleaned := raw

	// Try to extract JSON from markdown fences
	if m := jsonFenceRe.FindStringSubmatch(raw); len(m) > 1 {
		cleaned = m[1]
	} else if m := jsonObjectRe.FindStringSubmatch(raw); len(m) > 1 {
		// Extract the outermost JSON object from surrounding prose
		cleaned = m[1]
	}

	return strings.TrimSpace(cleaned), cleaned != raw
}

// RepairEnums normalizes category and severity values in-place.
// Returns true if any repair was made.
func RepairEnums(result *TriageResult) bool {
	repaired := false

	// Normalize category
	cat := strings.ToLower(strings.TrimSpace(string(result.Category)))
	// Try direct alias lookup
	if canonical, ok := categoryAliases[cat]; ok {
		result.Category = canonical
		repaired = true
	}
	// Try underscore-to-space conversion
	if !isValidCategory(result.Category) {
		spaced := strings.ReplaceAll(cat, "_", " ")
		if canonical, ok := categoryAliases[spaced]; ok {
			result.Category = canonical
			repaired = true
		}
	}

	// Normalize severity
	sev := strings.ToLower(strings.TrimSpace(string(result.Severity)))
	if canonical, ok := severityAliases[sev]; ok && result.Severity != canonical {
		result.Severity = canonical
		repaired = true
	}

	return repaired
}
