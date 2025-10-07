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
	"github.com/sentinez/sentinez/internal/edge/v1/mdw/logging"
	"github.com/sentinez/sentinez/internal/edge/v1/mdw/routing"
	"github.com/sentinez/sentinez/internal/edge/v1/mdw/secure"
	"github.com/sentinez/sentinez/internal/edge/v1/mdw/static"
)

func (s *Server) initialize(appConf *common.AppConfig) error {
	// init cache repository
	cache.Initialized(s.yaml, appConf)

	hostname := appConf.GetEnvConf().GetHostname()

	begin := static.NewStatic()

	begin.SetNext(logging.NewLogger()).
		SetNext(secure.NewDomain(hostname)).
		SetNext(secure.NewWAF()).
		SetNext(routing.NewRouter())

	s.core.Handle(begin.Handle)

	return nil
}
