// Copyright 2026 Duc-Hung Ho.
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

package sznormalize

import (
	"testing"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func tenantResource() *tenantpb.Resource {
	return &tenantpb.Resource{
		Id:       "rsrc_shop",
		TenantId: "tnt_acme",
		Name:     "shop",
		Domain:   "shop.example.com",
		Locations: []*edgepb.Location{{
			Location:  "/",
			ProxyPass: []*edgepb.Upstream{{Server: "10.0.0.1"}},
		}},
		TlsMode: edgepb.TlsMode_TLS_MODE_REENCRYPT,
	}
}

func TestEdgeResource(t *testing.T) {
	res := tenantResource()

	got := EdgeResource(res)

	want := &edgepb.Resource{
		Id:        "rsrc_shop",
		Name:      "shop",
		Domain:    "shop.example.com",
		Locations: res.GetLocations(),
		TlsMode:   edgepb.TlsMode_TLS_MODE_REENCRYPT,
	}
	assert.True(t, proto.Equal(want, got), "got %v", got)
	// the locations are copies: res does not change with the setting
	assert.NotSame(t, res.GetLocations()[0], got.GetLocations()[0])
}

func TestEdgeResourceNil(t *testing.T) {
	assert.Nil(t, EdgeResource(nil))
}

func TestTenantResource(t *testing.T) {
	srv := EdgeResource(tenantResource())

	got := TenantResource(srv)

	want := tenantResource()
	want.TenantId = ""
	assert.True(t, proto.Equal(want, got), "got %v", got)
	assert.NotSame(t, srv.GetLocations()[0], got.GetLocations()[0])
}

func TestTenantResourceNil(t *testing.T) {
	assert.Nil(t, TenantResource(nil))
}

func TestEdgeSettingWithoutBase(t *testing.T) {
	res := tenantResource()

	got := EdgeSetting(res, nil)

	assert.Equal(t, "tnt_acme", got.GetMetadata().GetTenantId())
	assert.True(t, proto.Equal(EdgeResource(res),
		got.GetIngress().GetResource()))
}

func TestEdgeSettingKeepsBase(t *testing.T) {
	tls := &edgepb.Tls{Certificates: []*edgepb.Certificate{{Id: "c1"}}}
	base := &edgepb.Setting{
		Metadata: &edgepb.Metadata{TenantId: "tnt_old", TenantName: "old"},
		Ingress: &edgepb.Ingress{
			Resource: &edgepb.Resource{Domain: "old.example.com"},
			Tls:      tls,
		},
		Delivery: &edgepb.Delivery{},
	}
	before := proto.CloneOf(base)

	got := EdgeSetting(tenantResource(), base)

	assert.Equal(t, "tnt_acme", got.GetMetadata().GetTenantId())
	assert.Equal(t, "old", got.GetMetadata().GetTenantName())
	assert.Equal(t, "shop.example.com",
		got.GetIngress().GetResource().GetDomain())
	assert.True(t, proto.Equal(tls, got.GetIngress().GetTls()))
	assert.NotNil(t, got.GetDelivery())
	assert.True(t, proto.Equal(before, base), "base changed")
}

func TestEdgeSettingKeepsBaseTenant(t *testing.T) {
	base := &edgepb.Setting{Metadata: &edgepb.Metadata{TenantId: "tnt_old"}}
	res := &tenantpb.Resource{Domain: "shop.example.com"}

	got := EdgeSetting(res, base)

	assert.Equal(t, "tnt_old", got.GetMetadata().GetTenantId())
}
