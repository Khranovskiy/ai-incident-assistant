package internal

import "strings"

const (
	scoreService  = 3
	scoreEndpoint = 3
	scoreKeyword  = 2
	scoreProvider = 2
	scoreSignal   = 1

	minRelevanceThreshold = 3
)

// SelectRelevantIncidents returns up to 2 most relevant past incidents.
// Returns nil if no incident meets the minimum relevance threshold.
func SelectRelevantIncidents(parsed ParsedSignals, incidents []PastIncident) []PastIncident {
	type scored struct {
		incident PastIncident
		score    int
	}

	var results []scored
	for _, inc := range incidents {
		s := scoreIncident(parsed, inc)
		if s >= minRelevanceThreshold {
			results = append(results, scored{inc, s})
		}
	}

	// Sort descending by score (simple insertion sort, max 4 items)
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	// Take top 2
	var top []PastIncident
	for i := 0; i < len(results) && i < 2; i++ {
		top = append(top, results[i].incident)
	}
	return top
}

func scoreIncident(parsed ParsedSignals, inc PastIncident) int {
	score := 0

	// Service overlap
	for _, ps := range parsed.Services {
		for _, is := range inc.Services {
			if strings.EqualFold(ps, is) {
				score += scoreService
			}
		}
	}

	// Endpoint overlap
	for _, pe := range parsed.Endpoints {
		for _, ie := range inc.Endpoints {
			if strings.EqualFold(pe, ie) {
				score += scoreEndpoint
			}
		}
	}

	// Provider/protocol terms are scored separately and excluded from keyword overlap
	// to avoid double-counting. Only specific provider names, not generic words.
	providerTerms := map[string]bool{"paygate": true, "smtp": true, "sms": true}

	// Keyword overlap (skip terms that match a provider term)
	for _, pk := range parsed.Keywords {
		if providerTerms[strings.ToLower(pk)] {
			continue
		}
		for _, ik := range inc.Keywords {
			if strings.EqualFold(pk, ik) {
				score += scoreKeyword
			}
		}
	}

	// Provider/protocol term overlap
	for term := range providerTerms {
		inParsed := false
		for _, pk := range parsed.Keywords {
			if strings.EqualFold(pk, term) {
				inParsed = true
				break
			}
		}
		inIncident := false
		for _, ik := range inc.Keywords {
			if strings.EqualFold(ik, term) {
				inIncident = true
				break
			}
		}
		if inParsed && inIncident {
			score += scoreProvider
		}
	}

	// Signal overlap (fuzzy: check if any parsed signal substring matches any incident signal)
	for _, ps := range parsed.Signals {
		psLower := strings.ToLower(ps)
		for _, is := range inc.Signals {
			isLower := strings.ToLower(is)
			if strings.Contains(psLower, isLower) || strings.Contains(isLower, psLower) {
				score += scoreSignal
				break
			}
		}
	}

	return score
}
