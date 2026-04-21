package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/khranovskiy/incident-assistant/internal"
)

func main() {
	text := flag.String("text", "", "Incident description text")
	file := flag.String("file", "", "Path to file containing incident description")
	verbose := flag.Bool("verbose", false, "Print diagnostic logs to stderr")
	flag.Parse()

	// Validate flags
	if *text == "" && *file == "" {
		fmt.Fprintln(os.Stderr, "error: provide either --text or --file")
		os.Exit(1)
	}
	if *text != "" && *file != "" {
		fmt.Fprintln(os.Stderr, "error: provide either --text or --file, not both")
		os.Exit(1)
	}

	// Load input
	var input string
	if *file != "" {
		data, err := os.ReadFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
			os.Exit(1)
		}
		input = string(data)
	} else {
		input = *text
	}

	if input == "" {
		fmt.Fprintln(os.Stderr, "error: incident description is empty")
		os.Exit(1)
	}

	// Load config
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: ANTHROPIC_API_KEY environment variable is required")
		os.Exit(1)
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	cfg := internal.Config{
		APIKey:  apiKey,
		Model:   model,
		Verbose: *verbose,
	}

	result, err := internal.Run(input, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}
