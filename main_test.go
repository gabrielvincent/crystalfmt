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
		if !info.IsDir() && strings.HasSuffix(path, "_input.cr") {
			basePath := strings.TrimSuffix(path, "_input.cr")
			expectedPath := basePath + "_expected.cr"

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
	testName := filepath.Base(strings.TrimSuffix(inputPath, "_input.cr"))

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
		want := strings.TrimSuffix(string(expected), "\n")

		// Compare results
		if got != want {
			diff := generateDiff(want, got)
			t.Errorf("%s: Formatting mismatch: %s", testName, diff)
		} else {
			fmt.Println(testName+":", "pass")
		}
	})
}

func generateDiff(want, got string) string {
	var diff strings.Builder

	// First compare lengths
	diff.WriteString(fmt.Sprintf("Lengths: want=%d got=%d\n", len(want), len(got)))

	// Find first difference
	minLen := min(len(got), len(want))

	firstDiffPos := -1
	for i := range minLen {
		if want[i] != got[i] {
			firstDiffPos = i
			break
		}
	}

	if firstDiffPos != -1 {
		// Show the context around the first difference
		startPos := max(0, firstDiffPos-20)
		endPosWant := min(len(want), firstDiffPos+20)
		endPosGot := min(len(got), firstDiffPos+20)

		diff.WriteString(fmt.Sprintf("First difference at position %d:\n", firstDiffPos))

		// Show hex values of differing bytes
		diff.WriteString(fmt.Sprintf("want[%d]=0x%02x (%q), got[%d]=0x%02x (%q)\n",
			firstDiffPos, want[firstDiffPos], string(want[firstDiffPos]),
			firstDiffPos, got[firstDiffPos], string(got[firstDiffPos])))

		// Show context
		diff.WriteString("Context:\n")
		diff.WriteString("want: ")
		for i := startPos; i < endPosWant; i++ {
			if i == firstDiffPos {
				diff.WriteString("[" + formatChar(want[i]) + "]")
			} else {
				diff.WriteString(formatChar(want[i]))
			}
		}
		diff.WriteString("\n")

		diff.WriteString("got:  ")
		for i := startPos; i < endPosGot; i++ {
			if i == firstDiffPos {
				diff.WriteString("[" + formatChar(got[i]) + "]")
			} else {
				diff.WriteString(formatChar(got[i]))
			}
		}
		diff.WriteString("\n")
	} else if len(want) != len(got) {
		// If no character differences were found but lengths differ
		diff.WriteString("No character differences found in the common portion, but lengths differ.\n")

		if len(want) > len(got) {
			diff.WriteString("want has additional characters: ")
			for i := len(got); i < min(len(got)+40, len(want)); i++ {
				diff.WriteString(formatChar(want[i]))
			}
			if len(want) > len(got)+40 {
				diff.WriteString("...")
			}
		} else {
			diff.WriteString("got has additional characters: ")
			for i := len(want); i < min(len(want)+40, len(got)); i++ {
				diff.WriteString(formatChar(got[i]))
			}
			if len(got) > len(want)+40 {
				diff.WriteString("...")
			}
		}
		diff.WriteString("\n")
	} else {
		// If we get here, the strings are identical
		diff.WriteString("Strings are identical by character comparison, but Go reports they are different.\n")
		diff.WriteString("This might indicate a Unicode normalization issue.\n")

		// Print the first few bytes of each string as hex
		const sampleSize = 100
		diff.WriteString(fmt.Sprintf("First %d bytes as hex:\n", min(sampleSize, len(want))))
		diff.WriteString("want: ")
		for i := range min(sampleSize, len(want)) {
			diff.WriteString(fmt.Sprintf("%02x ", want[i]))
		}
		diff.WriteString("\ngot:  ")
		for i := range min(sampleSize, len(got)) {
			diff.WriteString(fmt.Sprintf("%02x ", got[i]))
		}
		diff.WriteString("\n")
	}

	return diff.String()
}

// Helper function to format special characters in a readable way
func formatChar(b byte) string {
	switch b {
	case '\n':
		return "\\n"
	case '\r':
		return "\\r"
	case '\t':
		return "\\t"
	case ' ':
		return "·" // Use a visible character for space
	default:
		if b < 32 || b > 126 {
			return fmt.Sprintf("\\x%02x", b)
		}
		return string(b)
	}
}
