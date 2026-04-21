package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func log(cfg Config, format string, args ...any) {
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[incident-assistant] "+format+"\n", args...)
	}
}

// Run executes the full triage pipeline and returns the final JSON string.
func Run(incidentText string, cfg Config) (string, error) {
	ctx := context.Background()

	// Stage 1: Parse input
	log(cfg, "parsing incident text...")
	parsed := Parse(incidentText)
	log(cfg, "parsed signals: services=%v endpoints=%v keywords=%d signals=%d",
		parsed.Services, parsed.Endpoints, len(parsed.Keywords), len(parsed.Signals))

	// Stage 2: Load auxiliary knowledge
	log(cfg, "loading auxiliary knowledge...")
	sysDesc, err := LoadSystemDescription()
	if err != nil {
		return "", fmt.Errorf("failed to load system description: %w", err)
	}

	allIncidents, err := LoadPastIncidents()
	if err != nil {
		return "", fmt.Errorf("failed to load past incidents: %w", err)
	}

	// Stage 3: Select relevant past incidents
	log(cfg, "selecting relevant past incidents...")
	relevant := SelectRelevantIncidents(&parsed, allIncidents)
	if len(relevant) > 0 {
		ids := make([]string, len(relevant))
		for i := range relevant {
			ids[i] = relevant[i].ID
		}
		log(cfg, "selected past incidents: %v", ids)
	} else {
		log(cfg, "no relevant past incidents found (proceeding without)")
	}

	// Stage 4: Build context and prompt
	log(cfg, "building generation prompt...")
	userPrompt := BuildGenerationPrompt(incidentText, &parsed, sysDesc, relevant)

	// Stage 5: Generate
	log(cfg, "calling LLM (model=%s)...", cfg.Model)
	llm := NewLLMCaller(cfg.APIKey, cfg.Model)
	rawOutput, err := llm.Generate(ctx, systemRole, userPrompt)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}
	log(cfg, "received LLM response (%d bytes)", len(rawOutput))

	// Stage 6: Validate
	log(cfg, "validating output...")
	result, validationErr := ValidateRawJSON(rawOutput)
	if validationErr == nil {
		log(cfg, "validation passed")
		return formatJSON(result)
	}
	log(cfg, "validation failed: %v", validationErr)

	// Stage 7: Repair attempt
	log(cfg, "attempting repair...")
	repaired := rawOutput

	// Try JSON extraction repair
	if cleaned, changed := RepairJSON(rawOutput); changed {
		log(cfg, "extracted JSON from surrounding text")
		repaired = cleaned
	}

	// Try to parse the repaired JSON
	result, validationErr = ValidateRawJSON(repaired)
	if validationErr == nil {
		log(cfg, "repair succeeded (JSON extraction)")
		return formatJSON(result)
	}

	// Try enum normalization on the parsed struct
	var tryResult TriageResult
	if json.Unmarshal([]byte(repaired), &tryResult) == nil {
		if RepairEnums(&tryResult) {
			log(cfg, "applied enum normalization")
		}
		if valErr := Validate(&tryResult); valErr == nil {
			log(cfg, "repair succeeded (enum normalization)")
			return formatJSON(&tryResult)
		}
	}

	log(cfg, "repair insufficient, validation still fails: %v", validationErr)

	// Stage 8: Regeneration attempt
	log(cfg, "attempting regeneration with refined prompt...")
	regenPrompt := BuildRegenerationPrompt(rawOutput, validationErr.Error())
	rawOutput2, err := llm.Generate(ctx, systemRole, regenPrompt)
	if err != nil {
		return "", fmt.Errorf("regeneration failed: %w", err)
	}
	log(cfg, "received regenerated response (%d bytes)", len(rawOutput2))

	// Validate regenerated output (with repair)
	repaired2 := rawOutput2
	if cleaned, changed := RepairJSON(rawOutput2); changed {
		repaired2 = cleaned
	}

	result, validationErr = ValidateRawJSON(repaired2)
	if validationErr == nil {
		log(cfg, "regeneration succeeded")
		return formatJSON(result)
	}

	// Try enum normalization on regenerated output
	if json.Unmarshal([]byte(repaired2), &tryResult) == nil {
		RepairEnums(&tryResult)
		if valErr := Validate(&tryResult); valErr == nil {
			log(cfg, "regeneration succeeded after enum repair")
			return formatJSON(&tryResult)
		}
	}

	// Recovery exhausted
	return "", fmt.Errorf("recovery exhausted: output still invalid after regeneration: %v", validationErr)
}

func formatJSON(result *TriageResult) (string, error) {
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}
