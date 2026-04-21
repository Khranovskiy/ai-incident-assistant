package internal

import (
	"regexp"
	"strings"
)

var knownServices = []string{
	"api-gateway",
	"auth-service",
	"payment-service",
	"billing-service",
	"notification-service",
	"reporting-service",
}

var endpointRe = regexp.MustCompile(`(/[a-zA-Z0-9_\-]+(?:/[a-zA-Z0-9_\-]+)*)`)

var keywordPatterns = []string{
	"timeout", "5xx", "504", "502", "500", "401", "403", "404",
	"high cpu", "high memory", "latency", "slow",
	"connection error", "connection refused",
	"invalid token", "invalid signature", "invalid credentials",
	"smtp", "sms", "email", "notification",
	"paygate", "provider",
	"long-running queries", "long running queries",
	"database", "postgresql", "db",
	"login", "auth", "authentication",
	"payment", "card payment", "transaction",
	"gateway timeout",
}

var timeRe = regexp.MustCompile(`\d{1,2}:\d{2}\s*(?:UTC|GMT|[A-Z]{2,4})?`)

func Parse(text string) ParsedSignals {
	lower := strings.ToLower(text)

	var services []string
	for _, svc := range knownServices {
		if strings.Contains(lower, svc) {
			services = append(services, svc)
		}
	}

	// Extract endpoints
	endpointMatches := endpointRe.FindAllString(text, -1)
	var endpoints []string
	seen := map[string]bool{}
	for _, ep := range endpointMatches {
		// Skip very short matches that are likely not endpoints
		if len(ep) < 3 || !strings.Contains(ep[1:], "/") && len(ep) < 6 {
			continue
		}
		if !seen[ep] {
			endpoints = append(endpoints, ep)
			seen[ep] = true
		}
	}

	var keywords []string
	for _, kw := range keywordPatterns {
		if strings.Contains(lower, kw) {
			keywords = append(keywords, kw)
		}
	}

	// Extract signals: lines that contain error-like or symptom-like content
	var signals []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lineLower := strings.ToLower(trimmed)
		for _, indicator := range []string{
			"error", "fail", "timeout", "cannot", "not receive",
			"do not receive", "degradat", "spike", "increase",
			"high cpu", "high memory", "401", "403", "404", "500", "502", "504",
			"connection error", "invalid", "missing",
		} {
			if strings.Contains(lineLower, indicator) {
				signals = append(signals, trimmed)
				break
			}
		}
	}

	timeHints := timeRe.FindAllString(text, -1)

	return ParsedSignals{
		Services:  nonNil(services),
		Endpoints: nonNil(endpoints),
		Keywords:  nonNil(keywords),
		Signals:   nonNil(signals),
		TimeHints: nonNil(timeHints),
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
