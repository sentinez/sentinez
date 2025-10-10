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

package wafengine

import (
	"os"

	"github.com/corazawaf/coraza/v3"
	"github.com/sentinez/sentinez/security"
)

func NewWAF(version security.CRSVersion,
	ruleBasePath string, rulesetsFlag security.CRSFlag) (coraza.WAF, error) {

	rule := security.GetRule(version, rulesetsFlag)

	rootFS := os.DirFS(ruleBasePath)
	conf := coraza.NewWAFConfig().WithRootFS(rootFS).WithDirectives(rule)

	return coraza.NewWAF(conf)
}
