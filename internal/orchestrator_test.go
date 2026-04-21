package internal

import "testing"

// TestRun_PropagatesLLMError verifies that when the LLM call fails (e.g. due to
// an invalid API key), Run returns a non-nil error and does not panic.
func TestRun_PropagatesLLMError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: makes a real network call")
	}
	cfg := Config{
		APIKey: "invalid-key-for-testing",
		Model:  "claude-sonnet-4-20250514",
	}
	_, err := Run("payments are failing with 504 errors", cfg)
	if err == nil {
		t.Error("expected error with invalid API key")
	}
}
