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

// Package origin provides the origin source of the uri
package origin

import (
	"sync"

	edgeconfig "github.com/sentinez/sentinez/cmd/edge/v1/apps/config"
)

var pathMap sync.Map

func SetOrigin(config *edgeconfig.Routes) {
	for _, route := range config.Routes {
		pathMap.Store(route.PathPrefix, route.Target)
	}
}

func Source(path string) string {
	target, ok := pathMap.Load(path)
	if !ok {
		return ""
	}

	return target.(string)
}
