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

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/waf/v1"
	"github.com/sentinez/sentinez/plugins/ruleparser"
	templatez "github.com/sentinez/sentinez/tools/template"
)

func GenerateRulesGoFile(outputPath string, data *waf.CoreRulesets) error {

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpl, err := template.New("sentinez_rules").Parse(templatez.SentinezRule)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(outputPath, buf.Bytes(), 0644)
		return err
	}

	return os.WriteFile(outputPath, formatted, 0644)
}

func main() {
	var out = ""
	flag.StringVar(&out, "out", out, "directory for the generated rules file")
	flag.Parse()

	if out != "" {
		out = out + "/"
	}

	result, err := ruleparser.Parse("testdata/test_41_negated_operator_n.conf")
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	var rules waf.CoreRulesets
	if err = json.Unmarshal(data, &rules); err != nil {
		panic(err)
	}

	fmt.Println("Parsed Rules:", rules.GetRules()[0].GetConfiguration())

	decodedBytes, err := base64.StdEncoding.DecodeString(rules.GetRules()[0].GetConfigurationBase64())
	if err != nil {
		log.Fatalf("decode failed: %v", err)
	}
	decodedStatement := string(decodedBytes)

	rule := fmt.Sprintf("%s", decodedStatement)
	fmt.Println("Decoded Rule:", rule)

	err = GenerateRulesGoFile(out+"sentinez_rules.gen.go", &rules)
	if err != nil {
		panic(err)
	}
}
