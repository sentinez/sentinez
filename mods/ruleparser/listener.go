package seclang

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/sentinez/sentinez/mods/ruleparser/parser"
)

type ParserResult struct {
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Actions       *Action `json:"actions"`
	Configuration string  `json:"configuration"`
	Action        *Action `json:"action"`
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
	return &CustomErrorListener{antlr.NewDefaultErrorListener(), make([]error, 0)}
}

func (c *CustomErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	var err error
	if offendingSymbol == nil {
		err = fmt.Errorf("recognition error at line %d, column %d: %s", line, column, msg)
	} else {
		err = fmt.Errorf("syntax error at line %d, column %d: %v", line, column, offendingSymbol)
	}
	c.Errors = append(c.Errors, err)
}

func (t *TreeShapeListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
	// if you need to debug, enable this one below
	// fmt.Println("Entering rule:", ctx.GetText())
}

func (l *TreeShapeListener) EnterStmt(ctx *parser.StmtContext) {
	if len(l.results.Rules) > 0 {
		latest := l.results.Rules[len(l.results.Rules)-1]
		if strings.HasSuffix(latest.Configuration, "chain\"") {
			latest.Configuration += " " + ctx.GetText()
			l.results.Rules[len(l.results.Rules)-1] = latest

			action := &Action{
				Fields:    make(map[string][]string),
				Statement: ctx.GetText(),
			}
			current := action
			l.results.Rules[len(l.results.Rules)-1].Action.Children = action
			l.results.Rules[len(l.results.Rules)-1].Action = current

			return
		}
	}

	stmt := Rule{Configuration: removeFullLineComments(ctx.GetText())}
	stmt.Actions = &Action{
		Fields:    make(map[string][]string),
		Statement: stmt.Configuration,
	}
	stmt.Action = stmt.Actions

	l.results.Rules = append(l.results.Rules, stmt)
}

func (l *TreeShapeListener) ExitStmt(ctx *parser.StmtContext) {

}

func (l *TreeShapeListener) EnterAction(ctx *parser.ActionContext) {
	mapp := l.results.Rules[len(l.results.Rules)-1].Action.Fields
	action := strings.SplitN(ctx.GetText(), ":", 2)
	if len(action) > 1 {
		_, ok := mapp[action[0]]
		if !ok {
			mapp[action[0]] = []string{}
		}
		mapp[action[0]] = append(mapp[action[0]], strings.TrimSpace(action[1]))
	}
}

func (l *TreeShapeListener) EnterAction_with_params(ctx *parser.Action_with_paramsContext) {
}

func (l *TreeShapeListener) EnterAction_value(ctx *parser.Action_valueContext) {

}

func removeFullLineComments(input string) string {
	var result []string
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue // bỏ qua comment hoặc dòng trống
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}
