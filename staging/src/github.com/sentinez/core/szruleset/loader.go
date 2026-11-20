// Copyright 2026 Duc-Hung Ho.
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

package szruleset

import (
	"bytes"
	"encoding/base64"
	"strconv"

	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
)

// Rewrite changes the configuration (SecLang) of a rule before it is
// loaded. first loads it in front of the other rules of its Load call.
type Rewrite func(rule *rulesetpb.Rule, conf []byte) (out []byte, first bool)

// Loader concatenates the generated rules of a provider into the
// directives a coraza.WAF is built from, see NewWAF.
type Loader struct {
	buf bytes.Buffer

	// IDs of the rules loaded so far.
	ids     map[uint32]struct{}
	rewrite Rewrite // nil = rules loaded as is
}

// NewLoader returns a loader applying rewrite (may be nil) to the rules
// it loads.
func NewLoader(rewrite Rewrite) *Loader {
	return &Loader{
		ids:     make(map[uint32]struct{}),
		rewrite: rewrite,
	}
}

// Load appends the rules in order, but the rules the rewrite moves first.
// A rule whose configuration is not base64 is skipped.
func (l *Loader) Load(rulesFn []func() *rulesetpb.Rule) {
	rest := make([][]byte, 0, len(rulesFn))

	for _, fn := range rulesFn {
		rule := fn()

		conf, err := base64.StdEncoding.DecodeString(rule.GetConfiguration())
		if err != nil {
			continue
		}

		if id := RuleId(rule); id != 0 {
			l.ids[id] = struct{}{}
		}

		if l.rewrite != nil {
			var first bool
			if conf, first = l.rewrite(rule, conf); first {
				l.write(conf)
				continue
			}
		}

		rest = append(rest, conf)
	}

	for _, conf := range rest {
		l.write(conf)
	}
}

// Write appends raw directives.
func (l *Loader) Write(directives string) {
	l.write([]byte(directives))
}

// Loaded reports whether the rule ID has been loaded.
func (l *Loader) Loaded(id uint32) bool {
	_, ok := l.ids[id]
	return ok
}

// Export returns the directives loaded so far.
func (l *Loader) Export() string {
	return l.buf.String()
}

func (l *Loader) write(conf []byte) {
	if len(conf) == 0 {
		return
	}

	l.buf.Write(conf)

	if !bytes.HasSuffix(conf, []byte("\n")) {
		l.buf.WriteByte('\n')
	}
}

// RuleId the ID of the rule, 0 = not a rule (marker, directive).
func RuleId(rule *rulesetpb.Rule) uint32 {
	ids := rule.GetActions().GetFields().GetId()
	if len(ids) == 0 {
		return 0
	}

	id, err := strconv.ParseUint(ids[0], 10, 32)
	if err != nil {
		return 0
	}

	return uint32(id)
}
