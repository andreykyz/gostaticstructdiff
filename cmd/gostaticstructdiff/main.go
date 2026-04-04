package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/andreykyz/gostaticstructdiff/generator"
	"github.com/andreykyz/gostaticstructdiff/parser"
)

const version = "0.1.1"

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Input Go file (required)")
	outputFile := flag.String("output", "", "Output file (default: <input>_diff.go)")
	structName := flag.String("struct", "", "Specific struct to generate (default: all)")
	tagKey := flag.String("tag", "structtomap", "Tag key to look for (default: structtomap)")
	includeAll := flag.Bool("all", false, "Include all fields regardless of tags")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate type-safe diff structures and patch functions from Go structs annotated with configurable tags.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input models/user.go\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input models/user.go -output user_diff.go -struct User\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input models/user.go -tag mapstructure -all\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input models/user.go -verbose\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("gostaticstructdiff version %s\n", version)
		return
	}

	// Validate input
	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Set default output filename if not provided
	if *outputFile == "" {
		*outputFile = generateOutputFilename(*inputFile)
	}

	if *verbose {
		fmt.Printf("Input file: %s\n", *inputFile)
		fmt.Printf("Output file: %s\n", *outputFile)
		if *structName != "" {
			fmt.Printf("Struct filter: %s\n", *structName)
		}
		fmt.Printf("Tag key: %s\n", *tagKey)
		if *includeAll {
			fmt.Printf("Include all fields: true\n")
		}
	}

	// Process the file
	if err := processFile(*inputFile, *outputFile, *structName, *tagKey, *includeAll, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Println("Generation completed successfully")
	}
}

// processFile reads the input file, parses structs, generates diff code, and writes to output.
func processFile(inputFile, outputFile, structName, tagKey string, includeAll, verbose bool) error {
	// Parse the input file with options
	opts := parser.ParseOptions{
		TagKey:     tagKey,
		IncludeAll: includeAll,
	}
	structs, imports, err := parser.ParseFileWithOptions(inputFile, opts)
	if err != nil {
		return fmt.Errorf("failed to parse input file: %w", err)
	}

	if verbose {
		if includeAll {
			fmt.Printf("Found %d struct(s) (all fields included)\n", len(structs))
		} else {
			fmt.Printf("Found %d struct(s) with %s tags\n", len(structs), tagKey)
		}
		for _, s := range structs {
			fmt.Printf("  - %s (%d fields)\n", s.Name, len(s.Fields))
		}
	}

	// Filter by struct name if specified
	if structName != "" {
		filtered := make([]parser.StructInfo, 0)
		for _, s := range structs {
			if s.Name == structName {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("struct %q not found in input file", structName)
		}
		structs = filtered
		if verbose {
			fmt.Printf("Filtered to struct: %s\n", structName)
		}
	}

	// Determine package name from input file
	packageName, err := extractPackageName(inputFile)
	if err != nil {
		return fmt.Errorf("failed to determine package name: %w", err)
	}

	// Generate code (imports are passed from the parsed file)
	code, err := generator.Generate(structs, packageName, imports, version)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// generateOutputFilename generates a default output filename based on input.
// Example: "models/user.go" -> "models/user_diff.go"
func generateOutputFilename(input string) string {
	// Simple implementation: insert "_diff" before ".go"
	// Could be more sophisticated with path handling
	if len(input) > 3 && input[len(input)-3:] == ".go" {
		return input[:len(input)-3] + "_diff.go"
	}
	return input + "_diff.go"
}

// extractPackageName reads the package name from a Go file.
func extractPackageName(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("package declaration not found in %s", filename)
}
