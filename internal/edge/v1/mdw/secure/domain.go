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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/chains"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func isValidSingleLevelSubdomain(subdomain, root string) (string, bool) {
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

func NewDomain(hostname string) *Domain {
	return &Domain{
		Base:     &chains.Base{},
		hostname: hostname,
	}
}

type Domain struct {
	*chains.Base
	hostname string
}

func (d *Domain) Handle(ctx *httpxhz.Context) error {
	zlog.Info("[edge][handler] >>> domain")

	ns, ok := isValidSingleLevelSubdomain(string(ctx.Host()), d.hostname)
	if !ok {
		httpxhz.Forbidden(ctx)
		return nil
	}

	ctxValue, ok := httpxhz.GetRequestContext(ctx)
	if !ok {
		ctxValue = &common.HTTPContext{}
	}

	ctxValue.TenantNs = ns

	ctx = httpxhz.SetRequestContext(ctx, ctxValue)

	return d.HandleNext(ctx)
}
