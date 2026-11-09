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

package table

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sentinez/core"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// _envRe the table names of an env: lowercase words, those after the
// first may hold digits, e.g. the suffix of a view, _30m.
var _envRe = regexp.MustCompile(
	`^(dev|sandbox|prod)_sentinez_[a-z]+(_[a-z0-9]+)*$`)

func NewTable(appConf *settingpb.Config, tableName string) string {
	tableName = fmt.Sprintf("%s.%s.%s",
		appConf.GetFlag().GetEnvMode(), core.BaseName, tableName)
	return strings.ReplaceAll(tableName, ".", "_")
}

func IsValidTableName(tableName string) bool {
	return _envRe.MatchString(tableName)
}
