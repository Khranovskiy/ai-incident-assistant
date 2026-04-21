package internal

import (
	"fmt"
	"os"
)

func log(cfg Config, format string, args ...any) {
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[incident-assistant] "+format+"\n", args...)
	}
}

// Run executes the full triage pipeline and returns the final JSON string.
// TODO: wire parsing, knowledge loading, context building, generation, validation, repair.
func Run(incidentText string, cfg Config) (string, error) {
	return "", fmt.Errorf("not implemented: pipeline stages will be wired in subsequent tasks")
}
