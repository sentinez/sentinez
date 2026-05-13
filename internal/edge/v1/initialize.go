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
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/http"
	"github.com/sentinez/sentinez/internal/edge/v1/stream"
	"github.com/sentinez/sentinez/internal/shared/mem"
	"github.com/sentinez/shared/zlog"
)

func (s *Server) initialize(appConf *confpb.Config) error {
	// init cache repository
	mem.LoadConfiguration(s.setting, appConf)

	if err := stream.Init(stream.VETH0); err != nil {
		zlog.Errorf("failed to initialize stream: %v", err)
	}

	income := http.Init(appConf)
	s.core.Handle(income.Handle)

	return nil
}

func (s *Server) SetReverseProxyConstructor(fn ReverseProxyConstructor) {
	mem.SetReverseProxyConstructor(fn)
}
