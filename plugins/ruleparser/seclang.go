package ruleparser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/antlr4-go/antlr/v4"
	"github.com/sentinez/sentinez/plugins/ruleparser/parser"
)

func WriteJSONToFile(filename string, data any) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json failed: %w", err)
	}

	err = os.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

func Parse(inputPath string) (*ParserResult, error) {
	input, err := antlr.NewFileStream(inputPath)
	if err != nil {
		return nil, err
	}
	lexer := parser.NewSecLangLexer(input)

	lexerErrors := NewCustomErrorListener()
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(lexerErrors)

	parserErrors := NewCustomErrorListener()
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewSecLangParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(parserErrors)

	p.BuildParseTrees = true
	tree := p.Configuration()

	listener := NewTreeShapeListener()

	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// WriteJSONToFile(outputPath, listener.results)

	return &listener.results, nil
}
