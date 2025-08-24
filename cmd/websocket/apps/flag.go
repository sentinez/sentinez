// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apps provides the app setting for apiserver service
package apps

import (
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/ws/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/spf13/pflag"
)

var onceWS sync.Once

// apiServerFlags global variable
var apiServerFlags = &common.FlagWS{
	Address: ":7778",
}

// ParseFlag flag args for apiserver service
func ParseFlag() *common.FlagWS {
	onceWS.Do(func() {
		pflag.StringVarP(&apiServerFlags.Address, "address", "a",
			apiServerFlags.GetAddress(), "host address")
	})

	flags.Parse(ws.Metadata_ws)

	return apiServerFlags
}
