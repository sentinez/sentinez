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
	"cmp"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"google.golang.org/protobuf/proto"
)

// EdgeSetting is the setting the edge serves the tenant resource res
// with: a copy of base, or an empty setting when nil, whose tenant and
// ingress resource are those of res. The rest of base (certificates,
// security, delivery) is kept; base itself is not changed.
func EdgeSetting(res *tenantpb.Resource, base *edgepb.Setting,
) *edgepb.Setting {
	st := &edgepb.Setting{}
	if base != nil {
		st = proto.CloneOf(base)
	}

	if st.Metadata == nil {
		st.Metadata = &edgepb.Metadata{}
	}

	st.Metadata.TenantId = cmp.Or(res.GetTenantId(),
		st.Metadata.GetTenantId())

	if st.Ingress == nil {
		st.Ingress = &edgepb.Ingress{}
	}

	st.Ingress.Resource = EdgeResource(res)

	return st
}

// EdgeResource is the tenant resource res as the edge takes it: the
// resource of its setting, with copies of the locations of res; nil when
// res is.
func EdgeResource(res *tenantpb.Resource) *edgepb.Resource {
	if res == nil {
		return nil
	}

	return &edgepb.Resource{
		Id:        res.GetId(),
		Name:      res.GetName(),
		Domain:    res.GetDomain(),
		Locations: cloneLocations(res.GetLocations()),
		TlsMode:   res.GetTlsMode(),
	}
}

// TenantResource is the tenant resource of the edge resource srv: its id,
// name, domain, TLS mode and copies of its locations; nil when srv is.
func TenantResource(srv *edgepb.Resource) *tenantpb.Resource {
	if srv == nil {
		return nil
	}

	return &tenantpb.Resource{
		Id:        srv.GetId(),
		Name:      srv.GetName(),
		Domain:    srv.GetDomain(),
		Locations: cloneLocations(srv.GetLocations()),
		TlsMode:   srv.GetTlsMode(),
	}
}

func cloneLocations(locs []*edgepb.Location) []*edgepb.Location {
	if len(locs) == 0 {
		return nil
	}

	out := make([]*edgepb.Location, 0, len(locs))
	for _, loc := range locs {
		out = append(out, proto.CloneOf(loc))
	}

	return out
}
