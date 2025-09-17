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

package tenantfac

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	tenanthandler "github.com/sentinez/sentinez/internal/core/tenant/v1/handler"
)

func NewDefaultTenantHdl(_ *common.AppConfig) tenant.TenantServiceServer {

	return tenanthandler.New()
}
