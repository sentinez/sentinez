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
	"go/format"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/waf/v1"
	"github.com/sentinez/sentinez/plugins/ruleparser"
	templatez "github.com/sentinez/sentinez/tools/template"
)

func base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func generateRulesGoFile(outputPath string, data *waf.CoreRulesets) error {

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpl := template.New("sentinez_rules").Funcs(template.FuncMap{
		"base64Encode": base64Encode,
	})

	tmpl, err := tmpl.Parse(templatez.SentinezRuleFunc)
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

func parse() *waf.CoreRulesets {
	result, err := ruleparser.Parse("testdata/REQUEST-932-APPLICATION-ATTACK-RCE.conf")
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

	return &rules
}

func main() {
	var out = ""
	flag.StringVar(&out, "out", out, "directory for the generated rules file")
	flag.Parse()

	if out != "" {
		out = out + "/"
	}

	rules := parse()

	err := generateRulesGoFile(out+"sentinez_rules_func.gen.go", rules)
	if err != nil {
		panic(err)
	}
}
