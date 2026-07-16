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
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
)

func NewTable(appConf *settingpb.Config, tableName string) string {
	tableName = fmt.Sprintf("%s.%s.%s",
		appConf.GetFlag().GetEnvMode(), core.BaseName, tableName)
	return strings.ReplaceAll(tableName, ".", "_")
}

func NewPrimaryKey(tableName string) string {
	return fmt.Sprintf("%s.%s.", strings.ToLower(core.Code), tableName)
}

func IsValidTableName(tableName string) bool {
	matched, err := regexp.MatchString(core.EnvPattern, tableName)
	if err != nil {
		fmt.Println("Regex error:", err)
		return false
	}

	if matched {
		return true
	}

	return false
}
