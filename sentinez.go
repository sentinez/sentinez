// Copyright 2026 Sentinéz Labs.
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

package sentinez

import (
	"embed"
	"io/fs"

	corers "github.com/sentinez/core/rulesets"
)

var (
	//go:embed deploy/ruleroot/v4-16-0/*.data
	fsWAF4160 embed.FS

	//go:embed deploy/ruleroot/v4-17-0/*.data
	fsWAF4170 embed.FS
)

func WAF4160() (corers.Version, fs.FS) {
	sub, _ := fs.Sub(fsWAF4160, "deploy/ruleroot/v4-16-0")
	return corers.WAF4160, sub
}

func WAF4170() (corers.Version, fs.FS) {
	sub, _ := fs.Sub(fsWAF4160, "deploy/ruleroot/v4-17-0")
	return corers.WAF4170, sub
}
