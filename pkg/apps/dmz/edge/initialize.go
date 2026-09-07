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
	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http"
)

func (s *Server) initialize(appConf *settingpb.Config) error {
	s.mem.Start(appConf)

	// init cache repository
	s.mem.LoadServer(s.core)

	income := http.Init(appConf, s.mem)
	s.core.Handle(income.Handle)

	return nil
}

func (s *Server) SetOptions(opts ...corehttp.ServerOption) {
	s.options = append(s.options, opts...)
}
