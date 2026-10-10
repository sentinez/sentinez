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
	"errors"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sentinez/core/modsec/ruleparser"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
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
	var result strings.Builder
	for _, val := range names {
		result.WriteString(caser.String(val))
	}

	return result.String()
}

func normalizeVersion(input string) string {
	input = strings.ReplaceAll(input, ".", "_")
	input = strings.ReplaceAll(input, "/", "_")
	input = strings.ReplaceAll(input, "-", "_")
	return input
}

// ruleFile is the data rendered by the rules template.
type ruleFile struct {
	*rulepb.CoreRulesets
	Infos []*rulepb.RuleInfo
}

func generateRulesGoFile(outputPath string, data *ruleFile) error {

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

	var result strings.Builder
	for _, val := range nameArr {
		result.WriteString(caser.String(val))
	}

	return result.String()
}

func parse(filePath string) (*ruleFile, error) {
	result, err := ruleparser.Parse(filePath)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	rules := rulepb.CoreRulesets{Name: PascalCaseFileName(filePath)}
	if err = json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return &ruleFile{
		CoreRulesets: &rules,
		Infos:        ruleInfos(filePath, result.Rules),
	}, nil
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

// expandFiles resolves each pattern (e.g. "rules/*.conf") into a sorted,
// de-duplicated list of matching files.
func expandFiles(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]struct{})
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", pattern, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no file matches %q", pattern)
		}
		for _, match := range matches {
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			files = append(files, match)
		}
	}
	return files, nil
}

func generate(out, file string) error {
	rules, err := parse(file)
	if err != nil {
		return fmt.Errorf("parse %s: %w", file, err)
	}

	name := filepath.Join(out, normalizeFineName(file))
	err = generateRulesGoFile(name+".sentinez_rules.gen.go", rules)
	if err != nil {
		return fmt.Errorf("generate %s: %w", file, err)
	}
	return nil
}

func run() error {
	var out, file = "", ""
	flag.StringVar(&out, "out", out, "directory for the generated rules file")
	flag.StringVar(&file, "file", file,
		"coreruleset config file or glob pattern (e.g. 'rules/*.conf')")
	flag.Parse()

	// Positional args cover patterns already expanded by the shell.
	patterns := flag.Args()
	if file != "" {
		patterns = append([]string{file}, patterns...)
	}
	if len(patterns) == 0 {
		return errors.New("missing input coreruleset file config path")
	}

	files, err := expandFiles(patterns)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := generate(out, f); err != nil {
			return err
		}
		log.Printf("generated %s", f)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
