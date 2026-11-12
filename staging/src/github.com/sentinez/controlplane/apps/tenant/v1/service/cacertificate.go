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
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	cacertrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/cacertificates"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// CreateCaCertificate implements tenantpb.TenantServiceServer. A bundle is
// stored as one record per certificate, all or none; a certificate that the
// resource already has is refused.
func (svc *Service) CreateCaCertificate(ctx context.Context,
	req *tenantpb.CreateCaCertificateRequest,
) (*tenantpb.CreateCaCertificateResponse, error) {

	resource, err := svc.authz.RequireResource(
		ctx, req.GetResourceId(), authz.Write)
	if err != nil {
		return nil, err
	}

	certs, err := newCaCertificates(req, resource, time.Now())
	if err != nil {
		return nil, err
	}

	created, err := svc.storeCaCertificates(ctx, certs)
	if err != nil {
		return nil, err
	}

	return &tenantpb.CreateCaCertificateResponse{CaCertificates: created}, nil
}

// GetCaCertificate implements tenantpb.TenantServiceServer. The repo never
// gives the private key back, only whether there is one.
func (svc *Service) GetCaCertificate(ctx context.Context,
	req *tenantpb.GetCaCertificateRequest,
) (*tenantpb.GetCaCertificateResponse, error) {

	cert, err := svc.getCaCertificate(ctx, req.GetId(), authz.Read)
	if err != nil {
		return nil, err
	}

	return &tenantpb.GetCaCertificateResponse{CaCertificate: cert}, nil
}

// DeleteCaCertificate implements tenantpb.TenantServiceServer.
func (svc *Service) DeleteCaCertificate(ctx context.Context,
	req *tenantpb.DeleteCaCertificateRequest,
) (*tenantpb.DeleteCaCertificateResponse, error) {

	_, err := svc.getCaCertificate(ctx, req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = svc.cacert.Delete(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("cacert.Delete: %w", err)
	}

	return &tenantpb.DeleteCaCertificateResponse{}, nil
}

// ListCaCertificate implements tenantpb.TenantServiceServer. Like Get, it
// tells whether there is a private key and never gives it.
func (svc *Service) ListCaCertificate(ctx context.Context,
	req *tenantpb.ListCaCertificateRequest,
) (*tenantpb.ListCaCertificateResponse, error) {

	req, ok, err := authz.Scoped(ctx, svc.authz, req)
	if err != nil || !ok {
		return &tenantpb.ListCaCertificateResponse{}, err
	}

	resp, err := svc.cacert.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cacert.List: %w", err)
	}

	return resp, nil
}

// _usageServer is the usage of the TLS certificate the edge serves.
const _usageServer = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_SERVER

// errServerNeedsKey a certificate to serve, stored without its private key.
var errServerNeedsKey = szerror.StatusInvalidArgumentF(
	"a server certificate needs its private_key")

// errKeyCount a request with more than one private key.
var errKeyCount = szerror.StatusInvalidArgumentF(
	"the request holds more than one PEM private key")

// splitRequestKey is the certificates of the request, and its private key, if
// any: in private_key, or in the pem itself as in a single pem file.
func splitRequestKey(
	req *tenantpb.CreateCaCertificateRequest,
) (string, *pem.Block, error) {

	bundle, keys := splitPrivateKeys(req.GetPem())

	if field := req.GetPrivateKey(); strings.TrimSpace(field) != "" {
		_, fieldKeys := splitPrivateKeys(field)
		if len(fieldKeys) == 0 {
			return "", nil, szerror.StatusInvalidArgumentF(
				"private_key holds no PEM private key")
		}

		keys = append(keys, fieldKeys...)
	}

	switch len(keys) {
	case 0:
		return bundle, nil, nil
	case 1:
		return bundle, keys[0], nil
	}

	return "", nil, errKeyCount
}

// newCaCertificates is the records to store for the request, owned by
// resource.
func newCaCertificates(req *tenantpb.CreateCaCertificateRequest,
	resource *tenantpb.Resource, now time.Time,
) ([]*tenantpb.CaCertificate, error) {

	// the certificate the edge serves is not a CA, and is of no use without
	// its key
	server := req.GetUsage() == _usageServer

	bundle, key, err := splitRequestKey(req)
	if err != nil {
		return nil, err
	}

	if server && key == nil {
		return nil, errServerNeedsKey
	}

	certs, err := parseCaBundle(bundle, now, server)
	if err != nil {
		return nil, err
	}

	if err = attachPrivateKey(certs, key); err != nil {
		return nil, err
	}

	for i, cert := range certs {
		fillCaCertificate(cert, req, resource, i, len(certs))
	}

	return certs, nil
}

// fillCaCertificate sets what the request and the owner give to the i-th of
// n parsed certificates.
func fillCaCertificate(cert *tenantpb.CaCertificate,
	req *tenantpb.CreateCaCertificateRequest, resource *tenantpb.Resource,
	i, n int,
) {
	cert.Name = caCertificateName(req.GetName(), cert.GetName(), i, n)
	cert.Description = req.GetDescription()
	cert.Usage = req.GetUsage()
	cert.Status = typepb.Status_STATUS_ACTIVE
	cert.Owner = &typepb.Owner{
		ResourceId: resource.GetId(),
		TenantId:   resource.GetTenantId(),
	}
}

// caCertificateName is the name of the i-th of n certificates: the one given
// by the caller, numbered when there are several, else the parsed one.
func caCertificateName(given, parsed string, i, n int) string {
	switch {
	case given == "":
		return parsed
	case n == 1:
		return given
	default:
		return fmt.Sprintf("%s #%d", given, i+1)
	}
}

// storeCaCertificates creates certs in one transaction, which is rolled back
// when one of them exists already or can not be created.
func (svc *Service) storeCaCertificates(ctx context.Context,
	certs []*tenantpb.CaCertificate) ([]*tenantpb.CaCertificate, error) {

	txss, err := svc.tx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx.Begin: %w", err)
	}

	repo := svc.cacert.WithTX(txss)
	created := make([]*tenantpb.CaCertificate, 0, len(certs))

	for _, cert := range certs {
		stored, err := createNewCaCertificate(ctx, repo, cert)
		if err != nil {
			_ = txss.Rollback(ctx)
			return nil, err
		}

		created = append(created, stored)
	}

	if err = txss.Commit(ctx); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	return created, nil
}

// createNewCaCertificate creates cert unless its resource has the same
// fingerprint already: AlreadyExists.
func createNewCaCertificate(ctx context.Context,
	repo cacertrepo.ICaCertificate, cert *tenantpb.CaCertificate,
) (*tenantpb.CaCertificate, error) {

	same, err := repo.List(ctx, &tenantpb.ListCaCertificateRequest{
		ResourceIds:       []string{cert.GetOwner().GetResourceId()},
		FingerprintSha256: cert.GetFingerprintSha256(),
	})
	if err != nil {
		return nil, fmt.Errorf("cacert.List: %w", err)
	}

	if len(same.GetCaCertificates()) != 0 {
		return nil, szerror.StatusAlreadyExistsF(
			"ca certificate already exists: fingerprint_sha256=%s",
			cert.GetFingerprintSha256())
	}

	stored, err := repo.Create(ctx, cert)
	if err != nil {
		return nil, fmt.Errorf("cacert.Create: %w", err)
	}

	return stored, nil
}

// getCaCertificate returns the certificate, or a NotFound status when it is
// missing, after checking the caller has at least role least in the tenant
// of its resource.
func (svc *Service) getCaCertificate(ctx context.Context, id string,
	least typepb.Role) (*tenantpb.CaCertificate, error) {

	cert, err := svc.cacert.Get(ctx, id)
	if szerror.IsNoRows(err) {
		return nil, szerror.StatusNotFoundF(
			"ca certificate not found: id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("cacert.Get: %w", err)
	}

	if err = svc.authz.RequireOwner(ctx, cert.GetOwner(), least); err != nil {
		return nil, err
	}

	return cert, nil
}
