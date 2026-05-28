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

	tenanthandler "github.com/sentinez/modules/tenant/v1/handler"
	resourcerepo "github.com/sentinez/modules/tenant/v1/repos/resources"
	tenantsvc "github.com/sentinez/modules/tenant/v1/service"
	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/tenant/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
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
