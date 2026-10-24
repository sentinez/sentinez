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

package szrunner

import (
	"strings"

	"github.com/sentinez/core/internal/console"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// Banner returns the ASCII banner of the service described by meta.
func Banner(meta *typepb.XMeta) string {
	service := strings.Replace(meta.GetServiceName(), "_", " // ", 1)
	return console.GenFigure(service, meta.GetServiceKey())
}
