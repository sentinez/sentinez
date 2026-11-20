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

// Package szruleset runs the rulesets of any provider on Coraza: the
// Loader builds the directives from the generated rules, NewWAF the WAF
// and Rulesets runs a request through it. The provider specifics (OWASP
// CRS: paranoia levels, anomaly scores) live in the provider package,
// e.g. szcrs.
package szruleset

import (
	"io/fs"

	"github.com/corazawaf/coraza/v3"
)

// NewWAF builds a coraza.WAF from the directives, rootFs holds the files
// they include (data files, @pmFromFile lists).
func NewWAF(rootFs fs.FS, directives string) (coraza.WAF, error) {
	conf := coraza.NewWAFConfig().
		WithRootFS(rootFs).
		WithDirectives(directives)

	return coraza.NewWAF(conf)
}
