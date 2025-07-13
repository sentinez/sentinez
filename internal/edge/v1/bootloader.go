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
	"context"

	"github.com/sentinez/sentinez/internal/edge/v1/logic"
	"github.com/sentinez/sentinez/internal/edge/v1/routing"
	httpxf1mdw "github.com/sentinez/sentinez/pkg/core/httpx/f1/middleware"
)

func (s *Server) bootloader(_ context.Context) error {

	protected := httpxf1mdw.Protected(s.flag.GetRuleRoot())
	host := logic.Host(s.flag.GetHost())

	s.core.Use(host)
	s.core.Use(protected)

	return routing.Serve(s.config, s.core)
}
