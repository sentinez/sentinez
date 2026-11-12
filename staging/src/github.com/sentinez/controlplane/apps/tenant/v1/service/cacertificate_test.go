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
	"errors"
	"testing"

	"crypto/x509/pkix"

	"github.com/jackc/pgx/v5"
	cacertmock "github.com/sentinez/controlplane/apps/tenant/v1/repos/cacertificates/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_caId          = "cacrt_1"
	_usageUpstream = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM
)

// caService is a Service for user_1, whose role in _tenantId is role, and
// whose resource _resourceId belongs to resourceTenant.
func caService(t *testing.T, role typepb.Role, resourceTenant string,
	tx *postgres.Tx, certs *cacertmock.MockICaCertificate) *Service {
	t.Helper()

	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, _resourceId).Return(
		&tenantpb.Resource{Id: _resourceId, TenantId: resourceTenant},
		nil).Maybe()

	return New(tx, resources, memberships(t, "user_1", role), nil, certs)
}

// storedCa is a certificate of _resourceId, as the repo reads it: it never
// holds the private key, only whether there is one.
func storedCa(hasKey bool) *tenantpb.CaCertificate {
	return &tenantpb.CaCertificate{
		Id: _caId,
		Owner: &typepb.Owner{
			ResourceId: _resourceId, TenantId: _tenantId,
		},
		Pem:           "pem",
		HasPrivateKey: hasKey,
	}
}

// createsEach makes the repo create what it is given, under an id.
func createsEach(repo *cacertmock.MockICaCertificate) {
	repo.EXPECT().WithTX(mock.Anything).Return(repo)
	repo.EXPECT().List(mock.Anything, mock.Anything).
		Return(&tenantpb.ListCaCertificateResponse{}, nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, cert *tenantpb.CaCertificate,
		) (*tenantpb.CaCertificate, error) {
			cert.Id = _caId

			return cert, nil
		})
}

func TestCreateCaCertificateSplitsTheBundle(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"}, parent: root,
	})
	tx, conn := mockTx(t)
	conn.ExpectCommit()
	repo := cacertmock.NewMockICaCertificate(t)
	createsEach(repo)

	resp, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, tx, repo).
		CreateCaCertificate(asUser(t, "user_1"),
			&tenantpb.CreateCaCertificateRequest{
				ResourceId:  _resourceId,
				Name:        "acme",
				Description: "the acme chain",
				Usage:       _usageUpstream,
				Pem:         root.certPEM + inter.certPEM,
			})

	require.NoError(t, err)
	require.Len(t, resp.GetCaCertificates(), 2)
	assert.NoError(t, conn.ExpectationsWereMet())

	first, second := resp.GetCaCertificates()[0], resp.GetCaCertificates()[1]
	assert.Equal(t, "acme #1", first.GetName())
	assert.Equal(t, "acme #2", second.GetName())
	assert.Equal(t, "the acme chain", first.GetDescription())
	assert.Equal(t, typepb.Status_STATUS_ACTIVE, first.GetStatus())
	assert.Equal(t, _resourceId, first.GetOwner().GetResourceId())
	assert.Equal(t, _tenantId, first.GetOwner().GetTenantId())
	assert.False(t, first.GetHasPrivateKey())
}

func TestCreateCaCertificateNamesAfterTheCertificate(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	tx, conn := mockTx(t)
	conn.ExpectCommit()
	repo := cacertmock.NewMockICaCertificate(t)
	createsEach(repo)

	resp, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, tx, repo).
		CreateCaCertificate(asUser(t, "user_1"),
			&tenantpb.CreateCaCertificateRequest{
				ResourceId: _resourceId, Pem: root.certPEM,
			})

	require.NoError(t, err)
	assert.Equal(t, "Acme Root CA", resp.GetCaCertificates()[0].GetName())
}

func TestCreateCaCertificateGivesTheKeyToTheRepoOnly(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	tx, conn := mockTx(t)
	conn.ExpectCommit()
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().WithTX(mock.Anything).Return(repo)
	repo.EXPECT().List(mock.Anything, mock.Anything).
		Return(&tenantpb.ListCaCertificateResponse{}, nil)
	// the repo gets the key, to seal it, and does not give it back
	repo.EXPECT().Create(mock.Anything, mock.MatchedBy(
		func(cert *tenantpb.CaCertificate) bool {
			return cert.GetPrivateKey() == root.keyPEM
		})).RunAndReturn(
		func(_ context.Context, _ *tenantpb.CaCertificate,
		) (*tenantpb.CaCertificate, error) {
			return &tenantpb.CaCertificate{
				Id: _caId, HasPrivateKey: true,
			}, nil
		})

	resp, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, tx, repo).
		CreateCaCertificate(asUser(t, "user_1"),
			&tenantpb.CreateCaCertificateRequest{
				ResourceId: _resourceId, Pem: root.certPEM,
				PrivateKey: root.keyPEM,
			})

	require.NoError(t, err)
	assert.Empty(t, resp.GetCaCertificates()[0].GetPrivateKey())
	assert.True(t, resp.GetCaCertificates()[0].GetHasPrivateKey())
}

func TestCreateServerCertificateNeedsTheKey(t *testing.T) {
	leaf := leafNamed(t, "www.example.com")
	// no transaction, no repo call: nothing is stored
	svc := caService(t, typepb.Role_ROLE_LEADER, _tenantId, nil,
		cacertmock.NewMockICaCertificate(t))

	_, err := svc.CreateCaCertificate(asUser(t, "user_1"),
		&tenantpb.CreateCaCertificateRequest{
			ResourceId: _resourceId, Pem: leaf.certPEM,
			Usage: _usageServer,
		})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateServerCertificateIsNotACa(t *testing.T) {
	leaf := leafNamed(t, "www.example.com")
	tx, conn := mockTx(t)
	conn.ExpectCommit()
	repo := cacertmock.NewMockICaCertificate(t)
	createsEach(repo)

	resp, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, tx, repo).
		CreateCaCertificate(asUser(t, "user_1"),
			&tenantpb.CreateCaCertificateRequest{
				ResourceId: _resourceId, Pem: leaf.certPEM,
				PrivateKey: leaf.keyPEM,
				Usage:      _usageServer,
			})

	require.NoError(t, err)
	require.Len(t, resp.GetCaCertificates(), 1)
	assert.Equal(t, "www.example.com", resp.GetCaCertificates()[0].GetName())
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestCreateLeafCertificateIsRefusedWithoutServerUsage(t *testing.T) {
	leaf := leafNamed(t, "www.example.com")
	svc := caService(t, typepb.Role_ROLE_LEADER, _tenantId, nil,
		cacertmock.NewMockICaCertificate(t))

	_, err := svc.CreateCaCertificate(asUser(t, "user_1"),
		&tenantpb.CreateCaCertificateRequest{
			ResourceId: _resourceId, Pem: leaf.certPEM, PrivateKey: leaf.keyPEM,
			Usage: tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM,
		})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateCaCertificateInvalidInput(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	other := rootNamed(t, "Other", "Other Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"}, parent: root,
	})

	tests := []struct {
		name    string
		givePem string
		giveKey string
	}{
		{"not a certificate", root.keyPEM, ""},
		{"a bundle of two and a key of another", root.certPEM + inter.certPEM,
			other.keyPEM},
		{"a key of another certificate", root.certPEM, other.keyPEM},
		{"a private_key that is not a key", root.certPEM, root.certPEM},
		{"a key in both fields", root.certPEM + root.keyPEM, root.keyPEM},
		{"two keys", root.certPEM + root.keyPEM + other.keyPEM, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no transaction, no repo call: nothing is stored
			svc := caService(t, typepb.Role_ROLE_LEADER, _tenantId, nil,
				cacertmock.NewMockICaCertificate(t))

			_, err := svc.CreateCaCertificate(asUser(t, "user_1"),
				&tenantpb.CreateCaCertificateRequest{
					ResourceId: _resourceId, Pem: tt.givePem,
					PrivateKey: tt.giveKey,
				})

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestCreateCaCertificateAlreadyExists(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	tx, conn := mockTx(t)
	conn.ExpectRollback()
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().WithTX(mock.Anything).Return(repo)
	repo.EXPECT().List(mock.Anything, mock.MatchedBy(
		func(req *tenantpb.ListCaCertificateRequest) bool {
			return len(req.GetResourceIds()) == 1 &&
				req.GetResourceIds()[0] == _resourceId &&
				req.GetFingerprintSha256() != ""
		})).Return(&tenantpb.ListCaCertificateResponse{
		CaCertificates: []*tenantpb.CaCertificate{storedCa(false)},
	}, nil)

	_, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, tx, repo).
		CreateCaCertificate(asUser(t, "user_1"),
			&tenantpb.CreateCaCertificateRequest{
				ResourceId: _resourceId, Pem: root.certPEM,
			})

	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestCreateCaCertificateRollsBackWhenOneFails(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"}, parent: root,
	})
	tx, conn := mockTx(t)
	conn.ExpectRollback()
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().WithTX(mock.Anything).Return(repo)
	repo.EXPECT().List(mock.Anything, mock.Anything).
		Return(&tenantpb.ListCaCertificateResponse{}, nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(
		storedCa(false), nil).Once()
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(
		nil, errors.New("insert")).Once()

	_, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, tx, repo).
		CreateCaCertificate(asUser(t, "user_1"),
			&tenantpb.CreateCaCertificateRequest{
				ResourceId: _resourceId, Pem: root.certPEM + inter.certPEM,
			})

	require.Error(t, err)
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestCreateCaCertificateNeedsARole(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")

	tests := []struct {
		name               string
		giveRole           typepb.Role
		giveResourceTenant string
		wantCode           codes.Code
	}{
		{"a member can not create", typepb.Role_ROLE_MEMBER, _tenantId,
			codes.PermissionDenied},
		{"a resource of others", typepb.Role_ROLE_ADMIN, _otherTenantId,
			codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := caService(t, tt.giveRole, tt.giveResourceTenant, nil,
				cacertmock.NewMockICaCertificate(t))

			_, err := svc.CreateCaCertificate(asUser(t, "user_1"),
				&tenantpb.CreateCaCertificateRequest{
					ResourceId: _resourceId, Pem: root.certPEM,
				})

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestGetCaCertificateTellsWhetherThereIsAKey(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().Get(mock.Anything, _caId).Return(storedCa(true), nil)

	resp, err := caService(t, typepb.Role_ROLE_MEMBER, _tenantId, nil, repo).
		GetCaCertificate(asUser(t, "user_1"),
			&tenantpb.GetCaCertificateRequest{Id: _caId})

	require.NoError(t, err)
	assert.Empty(t, resp.GetCaCertificate().GetPrivateKey())
	assert.True(t, resp.GetCaCertificate().GetHasPrivateKey())
	assert.Equal(t, "pem", resp.GetCaCertificate().GetPem())
}

func TestGetCaCertificateNotFound(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().Get(mock.Anything, _caId).Return(nil, pgx.ErrNoRows)

	_, err := caService(t, typepb.Role_ROLE_ADMIN, _tenantId, nil, repo).
		GetCaCertificate(asUser(t, "user_1"),
			&tenantpb.GetCaCertificateRequest{Id: _caId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetCaCertificateOfOthers(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().Get(mock.Anything, _caId).Return(&tenantpb.CaCertificate{
		Id: _caId,
		Owner: &typepb.Owner{
			ResourceId: "rsrc_other", TenantId: _otherTenantId,
		},
	}, nil)
	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, "rsrc_other").Return(
		&tenantpb.Resource{Id: "rsrc_other", TenantId: _otherTenantId}, nil)
	svc := New(nil, resources,
		memberships(t, "user_1", typepb.Role_ROLE_ADMIN), nil, repo)

	_, err := svc.GetCaCertificate(asUser(t, "user_1"),
		&tenantpb.GetCaCertificateRequest{Id: _caId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteCaCertificate(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().Get(mock.Anything, _caId).Return(storedCa(false), nil)
	repo.EXPECT().Delete(mock.Anything, _caId).Return(nil)

	_, err := caService(t, typepb.Role_ROLE_LEADER, _tenantId, nil, repo).
		DeleteCaCertificate(asUser(t, "user_1"),
			&tenantpb.DeleteCaCertificateRequest{Id: _caId})

	require.NoError(t, err)
}

func TestDeleteCaCertificateNeedsWrite(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t) // Delete must not run
	repo.EXPECT().Get(mock.Anything, _caId).Return(storedCa(false), nil)

	_, err := caService(t, typepb.Role_ROLE_MEMBER, _tenantId, nil, repo).
		DeleteCaCertificate(asUser(t, "user_1"),
			&tenantpb.DeleteCaCertificateRequest{Id: _caId})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestDeleteCaCertificateNotFound(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().Get(mock.Anything, _caId).Return(nil, pgx.ErrNoRows)

	_, err := caService(t, typepb.Role_ROLE_ADMIN, _tenantId, nil, repo).
		DeleteCaCertificate(asUser(t, "user_1"),
			&tenantpb.DeleteCaCertificateRequest{Id: _caId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListCaCertificateIsScoped(t *testing.T) {
	repo := cacertmock.NewMockICaCertificate(t)
	repo.EXPECT().List(mock.Anything, mock.MatchedBy(
		func(req *tenantpb.ListCaCertificateRequest) bool {
			return len(req.GetResourceIds()) == 1 &&
				req.GetResourceIds()[0] == _resourceId
		})).Return(&tenantpb.ListCaCertificateResponse{
		Total: 2,
		CaCertificates: []*tenantpb.CaCertificate{
			storedCa(true), storedCa(false),
		},
	}, nil)
	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().List(mock.Anything, mock.Anything).Return(
		&tenantpb.ListResourceResponse{
			Resources: []*tenantpb.Resource{{Id: _resourceId}},
		}, nil)
	svc := New(nil, resources,
		memberships(t, "user_1", typepb.Role_ROLE_MEMBER), nil, repo)

	resp, err := svc.ListCaCertificate(asUser(t, "user_1"),
		&tenantpb.ListCaCertificateRequest{})

	require.NoError(t, err)
	require.Len(t, resp.GetCaCertificates(), 2)
	assert.EqualValues(t, 2, resp.GetTotal())
	assert.True(t, resp.GetCaCertificates()[0].GetHasPrivateKey())
	assert.False(t, resp.GetCaCertificates()[1].GetHasPrivateKey())
}

func TestListCaCertificateWithoutMembership(t *testing.T) {
	svc := New(nil, resourcemock.NewMockIResource(t),
		memberships(t, "user_1", typepb.Role_ROLE_ADMIN), nil,
		cacertmock.NewMockICaCertificate(t))

	resp, err := svc.ListCaCertificate(asUser(t, "user_9"), // List must not run
		&tenantpb.ListCaCertificateRequest{})

	require.NoError(t, err)
	assert.Empty(t, resp.GetCaCertificates())
}
