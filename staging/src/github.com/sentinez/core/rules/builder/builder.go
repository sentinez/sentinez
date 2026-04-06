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

package builder

import (
	"fmt"

	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// RuleBuilder provides a fluent API to build a Rule protobuf message.
type RuleBuilder struct {
	rule *ruleenginepb.Rule
}

// NewRule initializes a new RuleBuilder with a given ID and Name.
func NewRule(id, name string) *RuleBuilder {
	return &RuleBuilder{
		rule: &ruleenginepb.Rule{
			Id:      id,
			Name:    name,
			Enabled: true,
		},
	}
}

// RuleFromJSON deserializes a JSON string into a RuleBuilder.
func RuleFromJSON(data []byte) (*RuleBuilder, error) {
	rule := &ruleenginepb.Rule{}
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(data, rule); err != nil {
		return nil, fmt.Errorf("builder: failed to import from json: %w", err)
	}
	return &RuleBuilder{rule: rule}, nil
}

func (b *RuleBuilder) Description(desc string) *RuleBuilder {
	b.rule.Description = desc
	return b
}

func (b *RuleBuilder) Priority(p int32) *RuleBuilder {
	b.rule.Priority = p
	return b
}

func (b *RuleBuilder) Disable() *RuleBuilder {
	b.rule.Enabled = false
	return b
}

// Condition configures the conditional match parameters.
func (b *RuleBuilder) Condition(
	src ruleenginepb.FieldSource,
	op ruleenginepb.Operator,
	key string,
	val any,
) *RuleBuilder {
	v, err := structpb.NewValue(val)
	if err != nil {
		panic(fmt.Errorf("builder: invalid value for condition: %w", err))
	}

	b.rule.Condition = &ruleenginepb.Condition{
		Source:   src,
		Operator: op,
		Key:      key,
		Value:    v,
	}
	return b
}

// Action append an execution action to the rule.
func (b *RuleBuilder) Action(
	id string,
	t ruleenginepb.ActionType,
	params map[string]any,
) *RuleBuilder {
	p, err := structpb.NewStruct(params)
	if err != nil {
		panic(fmt.Errorf("builder: invalid params for action: %w", err))
	}

	b.rule.Actions = append(b.rule.Actions, &ruleenginepb.Action{
		Id:     id,
		Type:   t,
		Params: p,
	})
	return b
}

// Build returns the finalized Rule protobuf message.
func (b *RuleBuilder) Build() *ruleenginepb.Rule {
	return b.rule
}

// ToJSON exports the rule as a compliant protobuf JSON object.
func (b *RuleBuilder) ToJSON() ([]byte, error) {
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	return opts.Marshal(b.rule)
}

// ExprBuilder provides a fluent API to build an Expr protobuf message.
type ExprBuilder struct {
	expr *ruleenginepb.Expr
}

// NewExpr initializes a new ExprBuilder with a given ID and Name.
func NewExpr(id, name string) *ExprBuilder {
	return &ExprBuilder{
		expr: &ruleenginepb.Expr{
			Id:      id,
			Name:    name,
			Enabled: true,
		},
	}
}

// nolint
// ExprFromJSON deserializes a JSON string into an ExprBuilder.
func ExprFromJSON(data []byte) (*ExprBuilder, error) {
	expr := &ruleenginepb.Expr{}
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(data, expr); err != nil {
		return nil, fmt.Errorf("builder: failed to import expr from json: %w", err)
	}
	return &ExprBuilder{expr: expr}, nil
}

func (b *ExprBuilder) Description(desc string) *ExprBuilder {
	b.expr.Description = desc
	return b
}

func (b *ExprBuilder) Disable() *ExprBuilder {
	b.expr.Enabled = false
	return b
}

// AddRule adds an initial rule to the expression chain.
func (b *ExprBuilder) AddRule(rule *ruleenginepb.Rule) *ExprBuilder {
	b.expr.Rules = append(b.expr.Rules, rule)
	return b
}

// AddLogicAndRule appends a logic operator and a trailing rule.
func (b *ExprBuilder) AddLogicAndRule(
	logic ruleenginepb.Logic,
	rule *ruleenginepb.Rule,
) *ExprBuilder {
	b.expr.Logics = append(b.expr.Logics, logic)
	b.expr.Rules = append(b.expr.Rules, rule)
	return b
}

// Build returns the finalized Expr protobuf message.
func (b *ExprBuilder) Build() *ruleenginepb.Expr {
	return b.expr
}

// ToJSON exports the expr collection as a compliant protobuf JSON object.
func (b *ExprBuilder) ToJSON() ([]byte, error) {
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	return opts.Marshal(b.expr)
}
