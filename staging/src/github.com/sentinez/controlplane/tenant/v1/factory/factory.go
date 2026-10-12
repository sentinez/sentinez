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
	"fmt"

	tenanthandler "github.com/sentinez/controlplane/tenant/v1/handler"
	memberrepo "github.com/sentinez/controlplane/tenant/v1/repos/members"
	resourcerepo "github.com/sentinez/controlplane/tenant/v1/repos/resources"
	tenantrepo "github.com/sentinez/controlplane/tenant/v1/repos/tenants"
	tenantsvc "github.com/sentinez/controlplane/tenant/v1/service"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

func NewDefaultHandler(ctx context.Context,
	conf *settingpb.Config) (tenantpb.TenantServiceServer, error) {
	rscrepo, err := resourcerepo.New(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("resource repo: %w", err)
	}

	mbrrepo, err := memberrepo.New(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("member repo: %w", err)
	}

	tntrepo, err := tenantrepo.New(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("tenant repo: %w", err)
	}

	tenantsvcer := tenantsvc.New(rscrepo, mbrrepo, tntrepo)
	return tenanthandler.New(tenantsvcer), nil
}
