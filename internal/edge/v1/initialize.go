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
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/h/logging"
	"github.com/sentinez/sentinez/internal/edge/v1/h/routing"
	"github.com/sentinez/sentinez/internal/edge/v1/h/secure"
	"github.com/sentinez/sentinez/internal/edge/v1/h/static"
	"github.com/sentinez/sentinez/internal/edge/v1/h/waitingroom"
	"github.com/sentinez/sentinez/pkg/dmz/mem"
	"github.com/sentinez/sentinez/shared/zlog"
)

func (s *Server) initialize(appConf *confpb.Config) error {
	// init cache repository
	mem.Initialized(s.setting, appConf)

	hostname := appConf.GetEnv().GetHostname()

	begin := waitingroom.New()

	begin.SetNext(static.NewStatic()).
		SetNext(logging.NewLogger(zlog.LevelInfo)).
		SetNext(secure.NewDomain(hostname)).
		SetNext(secure.NewWAF(zlog.LevelInfo)).
		SetNext(routing.NewStandardRouter())

	s.core.Handle(begin.Handle)

	return nil
}
