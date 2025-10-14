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

	"github.com/sentinez/sentinez/internal/common/chains"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var _ chains.Handler = (*Domain)(nil)

func NewDomain(hostname string) *Domain {
	return &Domain{
		BaseHandler: chains.New(),
		hostname:    hostname,
	}
}

type Domain struct {
	*chains.BaseHandler
	hostname string
}

func (d *Domain) Handle(ctx *httpxhz.Context) error {
	zlog.Debugf("[edge][%s] >>> visit domain", ctx.GetReqID())

	ns, ok := d.isValidSingleLevelSubdomain(string(ctx.Host()), d.hostname)
	if !ok {
		return httpxhz.Forbidden(ctx)
	}

	ctxValue, ok := httpxhz.GetRequestContext(ctx)
	if !ok {
		ctxValue = &edgepb.Context{}
	}

	ctxValue.TenantNs = ns

	ctx = httpxhz.SetRequestContext(ctx, ctxValue)

	return d.HandleNext(ctx)
}

func (d *Domain) isValidSingleLevelSubdomain(
	subdomain, root string) (string, bool) {

	subLabels := strings.Split(subdomain, ".")
	rootLabels := strings.Split(root, ".")

	if len(subLabels) != len(rootLabels)+1 {
		return "", false
	}

	rootMatch := strings.Join(subLabels[len(subLabels)-len(rootLabels):], ".")
	rootMatch = strings.Split(rootMatch, ":")[0]
	if rootMatch == root {
		return subLabels[0], true
	}

	return "", false
}
