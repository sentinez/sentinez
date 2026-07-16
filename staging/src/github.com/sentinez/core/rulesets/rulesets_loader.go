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
	"bytes"
	"encoding/base64"
	"os"

	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
)

type RulesetsLoader struct {
	buf bytes.Buffer
}

func (rl *RulesetsLoader) Load(rulesetsFn []func() *rulepb.CoreRule) {
	for _, rule := range rulesetsFn {

		conf, err := base64.StdEncoding.DecodeString(rule().Configuration)
		if err != nil {
			continue
		}

		if !bytes.HasSuffix(conf, []byte("\n")) {
			conf = append(conf, '\n')
		}

		if _, err := rl.buf.Write(conf); err != nil {
			continue
		}
	}
}

func (rl *RulesetsLoader) Export() string {
	if rl.buf.Len() == 0 {
		return ""
	}

	if err := os.WriteFile("WAF.conf.lock", rl.buf.Bytes(), 0644); err != nil {
		return ""
	}

	return rl.buf.String()
}
