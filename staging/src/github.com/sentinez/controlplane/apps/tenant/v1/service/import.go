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

package tenantsvc

import (
	"context"
	"fmt"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// ImportServer stores the server of a proxy.yaml: tenant, whose name and
// plan are updated when it exists, and the resource of srv under its
// resource id, replaced so that importing again updates it. The resource
// is served at srv.resource_domain, errNoDomain without one.
func (svc *Service) ImportServer(ctx context.Context, tenant *tenantpb.Tenant,
	srv *edgepb.Server) error {

	if err := normalizeServer(srv); err != nil {
		return err
	}

	rsrc := &tenantpb.Resource{
		Id:       srv.GetResourceId(),
		TenantId: tenant.GetId(),
		Server:   srv,
		Status:   typepb.Status_STATUS_ACTIVE,
	}

	txss, err := svc.tx.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx.Begin: %w", err)
	}

	if err := svc.tenant.WithTX(txss).Upsert(ctx, tenant); err != nil {
		_ = txss.Rollback(ctx)
		return fmt.Errorf("tenant.Upsert: %w", err)
	}

	if err := svc.resource.WithTX(txss).Upsert(ctx, rsrc); err != nil {
		_ = txss.Rollback(ctx)
		return fmt.Errorf("resource.Upsert: %w", err)
	}

	if err := txss.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}
