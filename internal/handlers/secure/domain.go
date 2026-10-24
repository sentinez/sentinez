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
	"strings"

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/shared/bytesconv"
)

// DomainBasedId is the chain stage name of DomainBased.
const DomainBasedId = "DMA"

var (
	_ chains.Plugin                      = (*DomainBased)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*DomainBased)(nil)
)

// DomainBased resolves the tenant namespace from single-level
// subdomains of the Env hostname.
type DomainBased struct {
	hostname string
}

func (*DomainBased) ID() string { return DomainBasedId }

func (d *DomainBased) Provision(env *envplugin.Env) error {
	d.hostname = env.Hostname()
	return nil
}

func (d *DomainBased) Handle(
	ctx szhttp.Context, next chains.Handler) error {

	// szlog.Debug("[edge] >>> visit domain")

	xctx := ctx.X()
	if xctx == nil {
		return szhttp.BadRequest(ctx)
	}

	ns, ok := d.isValidSingleLevelSubdomain(
		bytesconv.B2s(ctx.Host()), d.hostname)
	if !ok {
		return szhttp.Forbidden(ctx)
	}

	xctx.Namespace = ns

	return next.Handle(ctx)
}

func (d *DomainBased) isValidSingleLevelSubdomain(
	subdomain, root string) (string, bool) {

	// Remove port if present
	if colon := strings.IndexByte(subdomain, ':'); colon >= 0 {
		subdomain = subdomain[:colon]
	}

	// Check if it ends with "." + root
	suffix := "." + root
	if !strings.HasSuffix(subdomain, suffix) {
		return "", false
	}

	// Extract subdomain part (before root)
	subPart := subdomain[:len(subdomain)-len(suffix)]

	// Ensure single-level (no extra dots) and non-empty
	if subPart == "" || strings.IndexByte(subPart, '.') >= 0 {
		return "", false
	}

	return subPart, true
}
