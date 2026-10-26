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
	"fmt"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
)

func (s *Server) initialize() error {

	// init cache repository
	s.memstore.LoadSetting(s.setting)
	s.memstore.LoadServer(s.server)

	chain, err := http.NewHandler(s.memstore,
		option.WithExporter(s.logExporter),
		option.WithHostname(s.conf.Get(settingpb.Senz_SENZ_HOSTNAME)),
		option.WithSecretKey(s.conf.Get(settingpb.Senz_SENZ_SECRET_KEY)),
	)
	if err != nil {
		return fmt.Errorf("build chain: %w", err)
	}

	s.chain = chain
	s.server.Handle(chain.Handle)

	return nil
}
