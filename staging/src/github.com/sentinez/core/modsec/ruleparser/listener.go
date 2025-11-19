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
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/sentinez/core/modsec/ruleparser/parser"
)

type ParserResult struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

type Rule struct {
	current       *Action
	Actions       *Action `json:"actions"`
	Configuration string  `json:"configuration"`
	Level         string  `json:"level"`
}

type Action struct {
	Statement string              `json:"statement"`
	Fields    map[string][]string `json:"fields"`
	Children  *Action             `json:"children"`
}

type TreeShapeListener struct {
	*parser.BaseSecLangParserListener
	results ParserResult
}

func NewTreeShapeListener() *TreeShapeListener {
	return new(TreeShapeListener)
}

type CustomErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []error
}

func NewCustomErrorListener() *CustomErrorListener {
	return &CustomErrorListener{
		antlr.NewDefaultErrorListener(), make([]error, 0)}
}

func (c *CustomErrorListener) SyntaxError(
	_ antlr.Recognizer,
	offendingSymbol any,
	line, column int,
	msg string,
	_ antlr.RecognitionException,
) {
	var err error
	if offendingSymbol == nil {
		err = fmt.Errorf("recognition error at line %d, column %d: %s",
			line, column, msg)
	} else {
		err = fmt.Errorf("syntax error at line %d, column %d: %v",
			line, column, offendingSymbol)
	}
	c.Errors = append(c.Errors, err)
}

func (l *TreeShapeListener) EnterEveryRule(_ antlr.ParserRuleContext) {
	// if you need to debug, enable this one below
	// fmt.Println("Entering rule:", ctx.GetText())
}

func (l *TreeShapeListener) EnterStmt(ctx *parser.StmtContext) {
	start := ctx.GetStart().GetStart() // start char index
	stop := ctx.GetStop().GetStop()
	inputStream := ctx.GetStart().GetInputStream()
	raw := inputStream.GetText(start, stop)
	processed := removeFullLineComments(raw)
	if processed == "" {
		return // skip empty statements
	}

	if len(l.results.Rules) > 0 {
		latest := l.results.Rules[len(l.results.Rules)-1]
		if strings.HasSuffix(latest.Configuration, "chain\"") {
			latest.Configuration += "\n" + processed
			l.results.Rules[len(l.results.Rules)-1] = latest

			action := &Action{
				Fields:    make(map[string][]string),
				Statement: processed,
			}
			current := action
			l.results.Rules[len(l.results.Rules)-1].current.Children = action
			l.results.Rules[len(l.results.Rules)-1].current = current

			return
		}
	}

	configuration := processed
	stmt := Rule{Configuration: configuration}
	stmt.Actions = &Action{
		Fields:    make(map[string][]string),
		Statement: stmt.Configuration,
	}
	stmt.current = stmt.Actions

	l.results.Rules = append(l.results.Rules, stmt)
}

func (l *TreeShapeListener) ExitStmt(_ *parser.StmtContext) {

}

func (l *TreeShapeListener) EnterAction(ctx *parser.ActionContext) {
	latest := len(l.results.Rules) - 1
	mapp := l.results.Rules[latest].current.Fields
	action := strings.SplitN(ctx.GetText(), ":", 2)
	if len(action) > 1 {
		_, ok := mapp[action[0]]
		if !ok {
			mapp[action[0]] = []string{}
		}
		mapp[action[0]] = append(mapp[action[0]], strings.TrimSpace(action[1]))

		processed := strings.TrimSuffix(strings.TrimPrefix(action[1], "'"), "'")
		if action[0] == "ver" && l.results.Version == "" {
			l.results.Version = processed
		}
		if strings.Contains(action[1], "paranoia-level") {
			l.results.Rules[latest].Level = processed
		}

	}
}

func (l *TreeShapeListener) EnterAction_with_params(
	_ *parser.Action_with_paramsContext) {
}

func (l *TreeShapeListener) EnterAction_value(_ *parser.Action_valueContext) {

}

func removeFullLineComments(input string) string {
	var result []string
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		result = append(result, line)
	}

	input = strings.Join(result, "\n")
	return input
}
