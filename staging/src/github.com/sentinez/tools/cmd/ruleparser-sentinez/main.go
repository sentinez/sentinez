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
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sentinez/core/modsec/ruleparser"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	templatez "github.com/sentinez/tools/internal/template"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func normalizeName(input string) string {
	caser := cases.Title(language.English)

	names := strings.Split(input, " ")
	result := ""
	for _, val := range names {
		result += caser.String(val)
	}

	return result
}

func normalizeVersion(input string) string {
	input = strings.ReplaceAll(input, ".", "_")
	input = strings.ReplaceAll(input, "/", "_")
	input = strings.ReplaceAll(input, "-", "_")
	return input
}

func generateRulesGoFile(outputPath string, data *rulepb.CoreRulesets) error {

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpl := template.New("sentinez_rules").Funcs(template.FuncMap{
		"base64Encode":     base64Encode,
		"normalizeVersion": normalizeVersion,
		"normalizeName":    normalizeName,
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

func PascalCaseFileName(filePath string) string {
	caser := cases.Title(language.English)

	name := normalizeFineName(filePath)
	name = strings.ToLower(name)
	nameArr := strings.Split(name, "_")

	result := ""
	for _, val := range nameArr {
		result += caser.String(val)
	}

	return result
}

func parse(filePath string) *rulepb.CoreRulesets {

	result, err := ruleparser.Parse(filePath)
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	rules := rulepb.CoreRulesets{Name: PascalCaseFileName(filePath)}
	if err = json.Unmarshal(data, &rules); err != nil {
		panic(err)
	}

	return &rules
}

func normalizeFineName(file string) string {
	elements := strings.Split(file, "/")

	configFile := elements[0]
	if len(elements) > 0 {
		configFile = elements[len(elements)-1]
	}

	name := strings.Split(configFile, ".")
	return strings.ReplaceAll(name[0], "-", "_")
}

func main() {
	var out, file = "", ""
	flag.StringVar(&out, "out", out, "directory for the generated rules file")
	flag.StringVar(&file, "file", file, "coreruleset configuration file")
	flag.Parse()

	if file == "" {
		log.Fatal("missing input coreruleset file config path")
	}

	if out != "" {
		out = out + "/"
	}

	rules := parse(file)

	name := out + normalizeFineName(file)
	err := generateRulesGoFile(name+".sentinez_rules.gen.go", rules)
	if err != nil {
		panic(err)
	}
}
