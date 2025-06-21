// Copyright 2025 Sentinez Labs.
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

// Package edgeflags provides the app setting for apiserver service
package edgeflags

import (
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/spf13/pflag"
)

var onceEdge sync.Once

var edgeFlags = &sentinez.FlagEdge{
	Address:  "0.0.0.0:7777",
	RulePath: "./default.conf",
	RuleRoot: "./boot/coreruleset",
}

// ParseFlag flag args for grpc service
func ParseFlag() *sentinez.FlagEdge {
	onceEdge.Do(func() {
		flags.SetConsole(edge.ASCII, names.EdgeV1, "dev")

		pflag.StringVarP(&edgeFlags.Address, "address", "a",
			edgeFlags.GetAddress(), "host address")

		pflag.StringVar(&edgeFlags.RulePath, "rule_path",
			edgeFlags.RulePath, "crs .config file path")

		pflag.StringVar(&edgeFlags.RuleRoot, "rule_root",
			edgeFlags.RuleRoot, "crs root path for rules")
	})

	_ = flags.Parse()

	return edgeFlags
}
