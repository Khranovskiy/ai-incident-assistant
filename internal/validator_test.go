package internal

import "testing"

func TestValidate_AcceptsValidResult(t *testing.T) {
	result := &TriageResult{
		Category: CategoryDBDegradation,
		Summary:  "Reporting service queries saturating shared DB, causing payment timeouts.",
		Affected: "Customers making payments via /payments/create",
		Severity: SeverityHigh,
		Hypotheses: []Hypothesis{
			{
				Title:     "Reporting queries starving payment transactions",
				Reasoning: "Long-running reporting queries hold locks and exhaust DB connections.",
				NextSteps: []string{
					"Check pg_stat_activity for long-running queries from reporting-service.",
					"Kill offending queries and verify response times recover.",
				},
			},
		},
	}

	if err := Validate(result); err != nil {
		t.Errorf("expected valid result to pass, got: %v", err)
	}
}

func TestValidate_RejectsInvalidCategory(t *testing.T) {
	result := &TriageResult{
		Category: "not_a_real_category",
		Summary:  "Something is broken.",
		Affected: "All users",
		Severity: SeverityMedium,
		Hypotheses: []Hypothesis{
			{
				Title:     "Unknown cause",
				Reasoning: "Need more data.",
				NextSteps: []string{"Check logs.", "Escalate."},
			},
		},
	}

	if err := Validate(result); err == nil {
		t.Error("expected invalid category to fail validation")
	}
}
