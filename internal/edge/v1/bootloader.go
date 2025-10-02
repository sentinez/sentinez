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

package edge

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/cache"
	"github.com/sentinez/sentinez/internal/edge/v1/dmz/logging"
	"github.com/sentinez/sentinez/internal/edge/v1/dmz/routing"
	"github.com/sentinez/sentinez/internal/edge/v1/dmz/secure"
)

func (s *Server) bootloader(appConf *common.AppConfig) error {
	// idx = 0
	s.core.Use(cache.HeaderCacheControl)
	// idx = 1
	s.core.Use(logging.WriterHandler)
	// idx = 2
	s.core.Use(secure.DomainHandler(appConf.GetEnvConf().GetHostname()))
	// idx = 3
	s.core.Use(secure.WAFHandler(appConf.GetFlag().GetRulePath()))

	return routing.Serve(s.yaml, s.core)
}
