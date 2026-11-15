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

//go:build ruleguard

// Package ruleguard keeps the dsl of the lint rules (hack/lint/rules.go,
// in the root module) in go.mod: ruleguard resolves it from the module it
// lints. The ruleguard tag keeps it out of the builds.
package ruleguard

import _ "github.com/quasilyte/go-ruleguard/dsl"
