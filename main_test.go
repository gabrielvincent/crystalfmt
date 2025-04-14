package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/crystal"
)

func TestFormatter(t *testing.T) {
	// Run all test cases in the testdata directory
	err := filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process input files
		if !info.IsDir() && strings.HasSuffix(path, ".input") {
			basePath := strings.TrimSuffix(path, ".input")
			expectedPath := basePath + ".expected"

			// Run this specific test case
			runFormatTest(t, path, expectedPath)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Error walking test files: %v", err)
	}
}

func runFormatTest(t *testing.T, inputPath, expectedPath string) {
	// Extract test name for better error reporting
	testName := filepath.Base(strings.TrimSuffix(inputPath, ".input"))

	// Create a subtest for each test case
	t.Run(testName, func(t *testing.T) {
		// Read input file
		input, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("Failed to read input file %s: %v", inputPath, err)
		}

		// Read expected output file
		expected, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read expected file %s: %v", expectedPath, err)
		}

		// Set up Tree-sitter parser
		parser := sitter.NewParser()
		parser.SetLanguage(crystal.GetLanguage())
		tree, err := parser.ParseCtx(context.Background(), nil, input)
		if err != nil {
			fmt.Println("--- Parsing failed:", err)
		}

		f := Formatter{
			strBuilder:         &strings.Builder{},
			source:             input,
			lineStartPositions: buildLineStartPositions(input),
			indentSize:         4,
		}

		f.formatNode(tree.RootNode(), 0)

		got := f.strBuilder.String()
		want := string(expected)

		// Compare results
		if got != want {
			t.Errorf("Formatting mismatch for %s:\nGOT:\n%s\nWANT:\n%s",
				testName, got, want)
		}
	})
}
