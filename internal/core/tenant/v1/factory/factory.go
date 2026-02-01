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
	"context"

	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	tenanthandler "github.com/sentinez/sentinez/internal/core/tenant/v1/handler"
	resourcerepo "github.com/sentinez/sentinez/internal/core/tenant/v1/repos/resources"
	tenantsvc "github.com/sentinez/sentinez/internal/core/tenant/v1/service"
	"github.com/sentinez/shared/zlog"
)

func NewDefaultHandler(ctx context.Context,
	conf *confpb.Config) tenantpb.TenantServiceServer {

	rscrepo, err := resourcerepo.New(ctx, conf)
	if err != nil {
		zlog.Fatalf("tenantpb factory: new resource err: %v", err)
	}

	tenantsvcer := tenantsvc.New(rscrepo)
	return tenanthandler.New(tenantsvcer)
}
