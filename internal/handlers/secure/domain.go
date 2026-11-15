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

package secure

import (
	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/bytesconv"
	"github.com/sentinez/pkg/sznormalize"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/settings"
)

// DomainBasedId is the chain stage name of DomainBased.
const DomainBasedId = "DMA"

var (
	_ chains.Plugin                      = (*DomainBased)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*DomainBased)(nil)
)

// DomainBased resolves the namespace of the request: its host, the
// resource domain of the server of a setting (server.resourceDomain).
type DomainBased struct {
	settings *settings.Setting
}

func (*DomainBased) ID() string { return DomainBasedId }

func (d *DomainBased) Provision(env *envplugin.Env) error {
	d.settings = env.Store().Setting()
	return nil
}

func (d *DomainBased) Handle(
	ctx szhttp.Context, next chains.Handler) error {

	xctx := ctx.X()
	if xctx == nil {
		return szhttp.BadRequest(ctx)
	}

	ns := sznormalize.ResourceDomain(
		bytesconv.B2s(szhttp.StripPort(ctx.Host())),
	)

	if misdirected(ctx, ns) {
		return szhttp.MisdirectedRequest(ctx)
	}

	st, err := d.settings.Load(ns)
	if err != nil {
		return szhttp.Forbidden(ctx)
	}

	xctx.Namespace = ns
	// the ids the next stages fill into their logs
	ctx.SetTenantId(st.GetMetadata().GetTenantId())
	ctx.SetResourceId(st.GetServer().GetResourceId())

	return next.Handle(ctx)
}

// misdirected reports whether the request came on a TLS conn made for
// another domain than its host ns. A browser reuses the conn of a domain
// for any other its certificate covers (HTTP/2 and HTTP/3 connection
// coalescing): a wildcard certificate of one domain would then serve the
// domains without one, or with another. 421 makes it open a conn for ns.
func misdirected(ctx szhttp.Context, ns string) bool {
	sni := ctx.ServerName()

	return sni != "" && sznormalize.ResourceDomain(sni) != ns
}
