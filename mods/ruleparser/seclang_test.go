package seclang

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	inputPath := "testdata/test_41_negated_operator_n.conf"
	outputPath := "testdata/test_41_negated_operator_n.json"

	err := Parse(inputPath, outputPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check if the output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file %s does not exist", outputPath)
	}
}
