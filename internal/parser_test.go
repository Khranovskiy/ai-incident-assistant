package internal

import (
	"slices"
	"testing"
)

func TestParse_ExtractsCanonicalSignals(t *testing.T) {
	text := `Sharp increase in response time for /payments/create (up to 5-7 seconds).
DB dashboards show high CPU and many long-running queries from reporting-service.
Some customers receive 504 Gateway Timeout from api-gateway.`

	p := Parse(text)

	for _, want := range []string{"reporting-service", "api-gateway"} {
		if !slices.Contains(p.Services, want) {
			t.Errorf("expected service %q in %v", want, p.Services)
		}
	}

	if !slices.Contains(p.Endpoints, "/payments/create") {
		t.Errorf("expected endpoint /payments/create in %v", p.Endpoints)
	}

	for _, want := range []string{"504", "high cpu"} {
		if !slices.Contains(p.Keywords, want) {
			t.Errorf("expected keyword %q in %v", want, p.Keywords)
		}
	}

	if len(p.Signals) == 0 {
		t.Error("expected at least one signal")
	}
}
