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

	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// GroupBuilder provides a fluent API for building RuleBaseds.
type GroupBuilder struct {
	group *rulepb.RuleBased
}

// NewGroup starts a new RuleBased builder with the given logical operator.
func NewGroup(op rulepb.Logic) *GroupBuilder {
	return &GroupBuilder{
		group: &rulepb.RuleBased{
			Node: &rulepb.RuleBased_Node{
				Operator: op,
				Rules:    []*rulepb.Rule{},
				Groups:   []*rulepb.RuleBased_Node{},
			},
		},
	}
}

func (b *GroupBuilder) WithID(id string) *GroupBuilder {
	b.group.Id = id
	return b
}

func (b *GroupBuilder) WithName(name string) *GroupBuilder {
	b.group.Name = name
	return b
}

func (b *GroupBuilder) WithDescription(desc string) *GroupBuilder {
	b.group.Description = desc
	return b
}

func (b *GroupBuilder) AddRule(r *rulepb.Rule) *GroupBuilder {
	node := b.group.GetNode()
	node.Rules = append(node.GetRules(), r)
	return b
}

func (b *GroupBuilder) AddGroup(g *rulepb.RuleBased) *GroupBuilder {
	if g != nil && g.GetNode() != nil {
		node := b.group.GetNode()
		node.Groups = append(node.GetGroups(), g.GetNode())
	}
	return b
}

func (b *GroupBuilder) Build() *rulepb.RuleBased {
	return b.group
}

// RuleBuilder provides a fluent API for building individual Rules.
type RuleBuilder struct {
	rule *rulepb.Rule
}

// NewRule starts a new Rule builder.
func NewRule() *RuleBuilder {
	return &RuleBuilder{
		rule: &rulepb.Rule{},
	}
}

func (b *RuleBuilder) WithID(id string) *RuleBuilder {
	b.rule.Id = id
	return b
}

func (b *RuleBuilder) WithName(name string) *RuleBuilder {
	b.rule.Name = name
	return b
}

func (b *RuleBuilder) WithCondition(
	src rulepb.FieldSource,
	op rulepb.Operator,
	val any,
	key ...string,
) *RuleBuilder {
	// structpb.NewValue handles []interface{} as ListValue.
	// We convert common slice types to []interface{}
	// to ensure correct behavior.
	finalVal := val
	switch v := val.(type) {
	case []string:
		items := make([]any, len(v))
		for i, s := range v {
			items[i] = s
		}
		finalVal = items
	case []any:
		finalVal = v
	}

	v, err := structpb.NewValue(finalVal)
	if err != nil {
		v, _ = structpb.NewValue(fmt.Sprintf("%v", val))
	}

	k := ""
	if len(key) > 0 {
		k = key[0]
	}

	b.rule.Condition = &rulepb.Condition{
		Source:   src,
		Operator: op,
		Value:    v,
		Key:      k,
	}
	return b
}

func (b *RuleBuilder) Build() *rulepb.Rule {
	return b.rule
}

// Helper functions for easy access
func And(rules ...*rulepb.Rule) *GroupBuilder {
	g := NewGroup(rulepb.Logic_LOGIC_AND)
	for _, r := range rules {
		g.AddRule(r)
	}
	return g
}

func Or(rules ...*rulepb.Rule) *GroupBuilder {
	g := NewGroup(rulepb.Logic_LOGIC_OR)
	for _, r := range rules {
		g.AddRule(r)
	}
	return g
}

func Not(node any) *GroupBuilder {
	g := NewGroup(rulepb.Logic_LOGIC_NOT)
	switch v := node.(type) {
	case *rulepb.Rule:
		g.AddRule(v)
	case *rulepb.RuleBased:
		g.AddGroup(v)
	}
	return g
}
