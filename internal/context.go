package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

const systemRole = `You are an expert on-call incident triage assistant for a payment platform.
Your job is to analyze incident descriptions and produce a structured triage result.
You must respond with ONLY a valid JSON object — no markdown fences, no prose, no explanation.
The response language must be English.`

func BuildGenerationPrompt(
	incidentText string,
	parsed *ParsedSignals,
	sysDesc *SystemDescription,
	relevantIncidents []PastIncident,
) string {
	var b strings.Builder

	b.WriteString("## Task\n\n")
	b.WriteString("Analyze the following production incident and return a structured triage result as a JSON object.\n\n")

	b.WriteString("## Output format\n\n")
	b.WriteString("Return ONLY a valid JSON object with these exact fields:\n\n")
	b.WriteString("```\n")
	b.WriteString(`{
  "category": "<one of the allowed categories>",
  "summary": "<what is happening and who is likely affected>",
  "affected": "<who is likely affected — users, services, or components>",
  "severity": "low|medium|high",
  "hypotheses": [
    {
      "title": "<hypothesis title>",
      "reasoning": "<why this hypothesis is plausible>",
      "next_steps": ["<step 1>", "<step 2>"]
    }
  ]
}
`)
	b.WriteString("```\n\n")

	b.WriteString("## Constraints\n\n")
	b.WriteString("- category must be exactly one of:\n")
	for _, c := range ValidCategories {
		fmt.Fprintf(&b, "  - %s\n", c)
	}
	b.WriteString("- severity must be exactly one of: low, medium, high\n")
	b.WriteString("- hypotheses: at most 3\n")
	b.WriteString("- each hypothesis must have title, reasoning, and next_steps\n")
	b.WriteString("- each next_steps must contain 2 to 3 non-empty strings\n")
	b.WriteString("- respond in English only\n")
	b.WriteString("- return JSON only — no markdown fences, no surrounding text\n\n")

	// System description
	b.WriteString("## System description\n\n")
	sysJSON, _ := json.MarshalIndent(sysDesc, "", "  ")
	b.Write(sysJSON)
	b.WriteString("\n\n")

	// Relevant past incidents
	if len(relevantIncidents) > 0 {
		b.WriteString("## Relevant past incidents\n\n")
		for i := range relevantIncidents {
			inc := &relevantIncidents[i]
			fmt.Fprintf(&b, "### %s: %s\n", inc.ID, inc.Title)
			fmt.Fprintf(&b, "Category: %s\n", inc.Category)
			fmt.Fprintf(&b, "Summary: %s\n", inc.Summary)
			fmt.Fprintf(&b, "Affected: %s\n", inc.LikelyAffected)
			if len(inc.DiagnosticHints) > 0 {
				b.WriteString("Diagnostic hints:\n")
				for _, h := range inc.DiagnosticHints {
					fmt.Fprintf(&b, "- %s\n", h)
				}
			}
			b.WriteString("\n")
		}
	}

	// Parsed signals
	b.WriteString("## Parsed signals from the incident\n\n")
	parsedJSON, _ := json.MarshalIndent(parsed, "", "  ")
	b.Write(parsedJSON)
	b.WriteString("\n\n")

	// Raw incident text
	b.WriteString("## Incident description\n\n")
	b.WriteString(incidentText)
	b.WriteString("\n")

	return b.String()
}

func BuildRegenerationPrompt(
	invalidOutput string,
	validationError string,
) string {
	var b strings.Builder

	b.WriteString("Your previous response was invalid. Fix it and return ONLY a valid JSON object.\n\n")

	b.WriteString("## Validation error\n\n")
	b.WriteString(validationError)
	b.WriteString("\n\n")

	b.WriteString("## Your previous (invalid) output\n\n")
	b.WriteString(invalidOutput)
	b.WriteString("\n\n")

	b.WriteString("## Required output format\n\n")
	b.WriteString("Return ONLY a valid JSON object with these exact fields:\n\n")
	b.WriteString(`{
  "category": "<one of the allowed categories>",
  "summary": "<string>",
  "affected": "<string>",
  "severity": "low|medium|high",
  "hypotheses": [
    {
      "title": "<string>",
      "reasoning": "<string>",
      "next_steps": ["<string>", "<string>"]
    }
  ]
}
`)
	b.WriteString("\n")

	b.WriteString("## Constraints\n\n")
	b.WriteString("- category must be exactly one of:\n")
	for _, c := range ValidCategories {
		fmt.Fprintf(&b, "  - %s\n", c)
	}
	b.WriteString("- severity must be exactly one of: low, medium, high\n")
	b.WriteString("- hypotheses: at most 3\n")
	b.WriteString("- each hypothesis must have title, reasoning, and next_steps\n")
	b.WriteString("- each next_steps must contain 2 to 3 non-empty strings\n")
	b.WriteString("- respond in English only\n")
	b.WriteString("- return JSON only — no markdown fences, no surrounding text\n")

	return b.String()
}
