// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ruleparser

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	inputPath := "testdata/test_41_negated_operator_n.conf"
	outputPath := "testdata/test_41_negated_operator_n.json"

	result, err := Parse(inputPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_ = WriteJSONToFile(outputPath, result)

	// Check if the output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file %s does not exist", outputPath)
	}
}
