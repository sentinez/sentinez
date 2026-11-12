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

package apiserver

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sznormalize"
	"github.com/sentinez/sentinez/api/client/consul"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	edgeyaml "github.com/sentinez/sentinez/pkg/apps/dmz/edge/yaml"
)

var errNoServer = errors.New("proxy.yaml has no server")

// _proxyImportTimeout bounds the import, which runs before the apiserver
// listens.
const _proxyImportTimeout = 10 * time.Second

// tenantImporter stores the server of a proxy.yaml.
type tenantImporter interface {
	ImportServer(ctx context.Context, tenant *tenantpb.Tenant,
		srv *edgepb.Server) error
}

// securityImporter stores the security block of a proxy.yaml.
type securityImporter interface {
	ImportSecurity(ctx context.Context, owner *typepb.Owner, name string,
		sec *edgepb.Security) error
}

// importProxy reads proxy.yaml from the Consul key of the proxy_config
// flag and stores it into the tenant and security services. Without
// SENZ_CONSUL_URI or the key there is nothing to import.
func importProxy(ctx context.Context, conf *settingpb.Config,
	tenant tenantImporter, security securityImporter) error {

	uri := conf.Get(settingpb.Senz_SENZ_CONSUL_URI)
	if uri == "" {
		szlog.Warnf("proxy import: no SENZ_CONSUL_URI, skipped")
		return nil
	}

	client, err := consul.New(uri)
	if err != nil {
		return fmt.Errorf("consul: %w", err)
	}

	// startup waits for the import: never longer than _proxyImportTimeout
	ctx, cancel := context.WithTimeout(ctx, _proxyImportTimeout)
	defer cancel()

	key := conf.GetFlag().GetProxyConfig()
	data, err := client.GetRaw(ctx, key)
	if errors.Is(err, consul.ErrKeyNotFound) {
		szlog.Warnf("proxy import: no consul key %q, skipped", key)
		return nil
	}
	if err != nil {
		return err
	}

	return storeProxy(ctx, data, tenant, security)
}

// storeProxy stores the proxy.yaml data into the tenant and security
// services.
func storeProxy(ctx context.Context, data []byte,
	tenant tenantImporter, security securityImporter) error {

	setting, err := edgeyaml.Parse(data)
	if err != nil {
		return err
	}
	sznormalize.NormalizeEdgeSetting(setting)
	srv := setting.GetServer()
	if srv == nil {
		return errNoServer
	}

	// the ids the edge gives the requests of a setting without them
	srv.ResourceId = cmp.Or(srv.GetResourceId(), defaults.ResourceId)
	meta := setting.GetMetadata()
	tenantId := cmp.Or(meta.GetTenantId(), defaults.TenantId)

	err = tenant.ImportServer(ctx, &tenantpb.Tenant{
		Id:   tenantId,
		Name: cmp.Or(meta.GetTenantName(), tenantId),
		Plan: meta.GetPlan(),
	}, srv)
	if err != nil {
		return fmt.Errorf("import server: %w", err)
	}

	owner := &typepb.Owner{TenantId: tenantId, ResourceId: srv.GetResourceId()}
	err = security.ImportSecurity(ctx, owner, srv.GetName(),
		setting.GetSecurity())
	if err != nil {
		return err
	}

	szlog.Infof("proxy import: server %q stored as %s",
		srv.GetName(), srv.GetResourceId())
	return nil
}
