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

package tenantsvc

import (
	"context"
	"crypto/x509/pkix"
	"testing"
	"time"

	cacertmock "github.com/sentinez/controlplane/apps/tenant/v1/repos/cacertificates/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _shopDomain = "shop.example.com"

// certOf is c stored for _resourceId under usage, as the repo reads it.
func certOf(c *testCa, usage tenantpb.CaCertificateUsage,
	hasKey bool) *tenantpb.CaCertificate {
	return &tenantpb.CaCertificate{
		Usage:         usage,
		Pem:           c.certPEM,
		NotBefore:     timestamppb.New(c.cert.NotBefore),
		NotAfter:      timestamppb.New(c.cert.NotAfter),
		HasPrivateKey: hasKey,
	}
}

// serverFor a server certificate of domain, with its key.
func serverFor(t *testing.T, domain string) *tenantpb.CaCertificate {
	t.Helper()

	return certOf(newTestCa(t, caOptions{
		subject:  pkix.Name{CommonName: domain},
		notCa:    true,
		dnsNames: []string{domain},
	}), _usageServer, true)
}

// expiredCa an upstream CA expired before _now.
func expiredCa(t *testing.T) *tenantpb.CaCertificate {
	t.Helper()

	return certOf(newTestCa(t, caOptions{
		subject:  pkix.Name{CommonName: "Old CA"},
		notAfter: _now.Add(-time.Minute),
	}), _usageUpstreamCa, false)
}

// certsOf a repo holding certs for _resourceId.
func certsOf(t *testing.T,
	certs ...*tenantpb.CaCertificate) *cacertmock.MockICaCertificate {
	t.Helper()

	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().List(mock.Anything, &tenantpb.ListCaCertificateRequest{
		ResourceIds: []string{_resourceId},
	}).Return(&tenantpb.ListCaCertificateResponse{
		CaCertificates: certs,
	}, nil).Maybe()

	return repo
}

type tlsReadyCase struct {
	name      string
	giveMode  edgepb.TlsMode
	giveCerts []*tenantpb.CaCertificate
	want      bool
}

func tlsReadyCases(t *testing.T) []tlsReadyCase {
	t.Helper()

	var (
		server   = serverFor(t, _shopDomain)
		other    = serverFor(t, "other.example.com")
		noKey    = certOf(leafNamed(t, _shopDomain), _usageServer, false)
		client   = certOf(leafNamed(t, "edge"), _usageUpstreamClient, true)
		terminal = edgepb.TlsMode_TLS_MODE_TERMINATE
		reencr   = edgepb.TlsMode_TLS_MODE_REENCRYPT
		auto     = edgepb.TlsMode_TLS_MODE_UNSPECIFIED
	)

	return []tlsReadyCase{
		{"passthrough needs nothing",
			edgepb.TlsMode_TLS_MODE_PASSTHROUGH, nil, true},
		{"terminate with a server certificate", terminal,
			[]*tenantpb.CaCertificate{server}, true},
		{"terminate without one", terminal, nil, false},
		{"server certificate of another domain", terminal,
			[]*tenantpb.CaCertificate{other}, false},
		{"server certificate without its key", terminal,
			[]*tenantpb.CaCertificate{noKey}, false},
		{"terminate ignores the upstream", terminal,
			[]*tenantpb.CaCertificate{server, expiredCa(t)}, true},
		{"re-encrypt with a valid client", reencr,
			[]*tenantpb.CaCertificate{server, client}, true},
		{"re-encrypt with expired CAs only", reencr,
			[]*tenantpb.CaCertificate{server, expiredCa(t)}, false},
		{"auto re-encrypts too", auto,
			[]*tenantpb.CaCertificate{server, expiredCa(t)}, false},
		{"auto without a server certificate", auto,
			[]*tenantpb.CaCertificate{client}, false},
	}
}

func TestTlsReady(t *testing.T) {
	for _, tt := range tlsReadyCases(t) {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(nil, nil, nil, nil, certsOf(t, tt.giveCerts...))
			svc.now = func() time.Time { return _now }

			got, err := svc.tlsReady(t.Context(), &tenantpb.Resource{
				Id: _resourceId, Domain: _shopDomain, TlsMode: tt.giveMode,
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateResourcePendsWithoutTls(t *testing.T) {
	want := storedResource()
	want.TlsMode = edgepb.TlsMode_TLS_MODE_TERMINATE
	want.Status = _statusPending

	// the status is written though not in the mask
	svc := New(nil, resourcesWith(t, want, []string{
		tenantpb.Resource_TlsMode, tenantpb.Resource_Status,
	}), nil, nil, certsOf(t))
	_, err := svc.UpdateResource(asAdmin(t), &tenantpb.UpdateResourceRequest{
		Id:         _resourceId,
		Resource:   &tenantpb.Resource{TlsMode: want.GetTlsMode()},
		UpdateMask: maskOf(tenantpb.Resource_TlsMode),
	})

	require.NoError(t, err)
}

func TestCreateResourceIsActiveOnlyInPassthrough(t *testing.T) {
	tests := []struct {
		name       string
		giveMode   edgepb.TlsMode
		wantStatus typepb.Status
	}{
		{"passthrough", edgepb.TlsMode_TLS_MODE_PASSTHROUGH, _statusActive},
		{"terminate", edgepb.TlsMode_TLS_MODE_TERMINATE, _statusPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := resourcemock.NewMockIResource(t)
			repo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, r *tenantpb.Resource,
				) (*tenantpb.Resource, error) {
					return r, nil
				})

			resp, err := New(nil, repo, nil, nil, nil).CreateResource(
				asAdmin(t), &tenantpb.CreateResourceRequest{
					TenantId: _tenantId,
					Status:   _statusActive,
					Resource: &edgepb.Resource{
						Domain: _shopDomain, TlsMode: tt.giveMode,
					},
				})

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.GetResource().GetStatus())
		})
	}
}

func TestDeleteCaCertificatePendsTheResource(t *testing.T) {
	stored := storedResource()
	want := storedResource()
	want.Status = _statusPending

	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, _resourceId).Return(stored, nil)
	resources.EXPECT().Update(mock.Anything, want,
		[]string{tenantpb.Resource_Status}).Return(nil)

	// the last server certificate is deleted
	certs := certsOf(t)
	certs.EXPECT().Get(mock.Anything, _caId).Return(storedCa(true), nil)
	certs.EXPECT().Delete(mock.Anything, _caId).Return(nil)

	_, err := New(nil, resources, nil, nil, certs).DeleteCaCertificate(
		asAdmin(t), &tenantpb.DeleteCaCertificateRequest{Id: _caId})

	require.NoError(t, err)
}

func TestCreateCaCertificateActivatesThePendingResource(t *testing.T) {
	server := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: _shopDomain}, notCa: true,
		dnsNames: []string{_shopDomain},
	})
	pending := storedResource()
	pending.Domain = _shopDomain
	pending.Status = _statusPending
	want := proto.CloneOf(pending)
	want.Status = _statusActive

	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, _resourceId).Return(pending, nil)
	resources.EXPECT().Update(mock.Anything, want,
		[]string{tenantpb.Resource_Status}).Return(nil)

	tx, conn := mockTx(t)
	conn.ExpectCommit()
	// the resource then holds the certificate, matched before the List
	// of createsEach
	certs := certsOf(t, certOf(server, _usageServer, true))
	createsEach(certs)

	svc := New(tx, resources, nil, nil, certs)
	svc.now = func() time.Time { return _now }
	_, err := svc.CreateCaCertificate(asAdmin(t),
		&tenantpb.CreateCaCertificateRequest{
			ResourceId: _resourceId,
			Usage:      _usageServer,
			Pem:        server.certPEM,
			PrivateKey: server.keyPEM,
		})

	require.NoError(t, err)
}
