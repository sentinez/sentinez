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

package corers

import (
	"fmt"
	"io/fs"

	"github.com/corazawaf/coraza/v3"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
)

// NewWAF builds a coraza.WAF from the CRS setting of a server.
func NewWAF(
	version Version, rootFs fs.FS, setting *corerulesetpb.CoreRuleset,
) (coraza.WAF, error) {
	rule, err := GenerateRulesets(version, setting)
	if err != nil {
		return nil, fmt.Errorf("generate rulesets: %w", err)
	}

	conf := coraza.NewWAFConfig().WithRootFS(rootFs).WithDirectives(rule)

	return coraza.NewWAF(conf)
}
