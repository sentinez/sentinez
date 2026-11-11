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

package cacertrepo

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// ICaCertificate stores the CA certificates of the resources. A certificate
// is immutable once stored, so there is no Update: to change one, delete it
// and create it again.
//
// The pem and the private key are sealed in the database. The pem comes
// back in clear; the private key never comes back: a certificate comes out
// with has_private_key only, and nothing but the repo can unseal the key.
//
//nolint:lll
type ICaCertificate interface {
	Create(ctx context.Context, cert *tenantpb.CaCertificate) (*tenantpb.CaCertificate, error)
	Get(ctx context.Context, id string) (*tenantpb.CaCertificate, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ICaCertificate

	List(ctx context.Context, req *tenantpb.ListCaCertificateRequest) (*tenantpb.ListCaCertificateResponse, error)
}

var _ ICaCertificate = (*CaCertificates)(nil)

// New creates the repository backed by Postgres; sealer seals the pem and
// the private key before they are stored.
func New(ctx context.Context, appConf *settingpb.Config,
	sealer *crypto.Sealer) (ICaCertificate, error) {

	storage, err := postgres.New[tenantpb.CaCertificate](ctx, appConf,
		schema.TenantCaCertificates())
	if err != nil {
		return nil, err
	}

	return &CaCertificates{storage: storage, sealer: sealer}, nil
}

type CaCertificates struct {
	storage dbx.Database[tenantpb.CaCertificate]
	sealer  *crypto.Sealer
}

func (r *CaCertificates) WithTX(tx *postgres.TxSession) ICaCertificate {
	return &CaCertificates{
		storage: postgres.WithTx(tx, r.storage),
		sealer:  r.sealer,
	}
}

// Create implements ICaCertificate. The returned certificate has its id, and
// no private key.
func (r *CaCertificates) Create(ctx context.Context,
	cert *tenantpb.CaCertificate) (*tenantpb.CaCertificate, error) {

	cert.Id = szrand.NewUUID(szprefix.CaCertificate)

	values, err := r.insertValues(cert)
	if err != nil {
		return nil, err
	}

	query := postgres.InsertBuilder(r.storage, values)
	if _, err = r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	hideKey(cert)

	return cert, nil
}

// insertValues are the columns of cert, with its secrets sealed for its id.
func (r *CaCertificates) insertValues(
	cert *tenantpb.CaCertificate) (postgres.M, error) {

	pem, err := r.sealer.Seal(cert.GetPem(), cert.GetId())
	if err != nil {
		return nil, fmt.Errorf("seal pem: %w", err)
	}

	// no key is stored as is, rather than as the seal of nothing
	var key string
	if cert.GetPrivateKey() != "" {
		key, err = r.sealer.Seal(cert.GetPrivateKey(), cert.GetId())
		if err != nil {
			return nil, fmt.Errorf("seal private key: %w", err)
		}
	}

	return plainValues(cert, pem, key), nil
}

func plainValues(
	cert *tenantpb.CaCertificate, pem, key string) postgres.M {

	owner := cert.GetOwner()

	return postgres.M{
		tenantpb.CaCertificate_Id:                cert.GetId(),
		tenantpb.CaCertificate_Name:              cert.GetName(),
		tenantpb.CaCertificate_Description:       cert.GetDescription(),
		tenantpb.CaCertificate_Usage:             int32(cert.GetUsage()),
		tenantpb.CaCertificate_Status:            int32(cert.GetStatus()),
		tenantpb.CaCertificate_Pem:               pem,
		tenantpb.CaCertificate_PrivateKey:        key,
		tenantpb.CaCertificate_Subject:           cert.GetSubject(),
		tenantpb.CaCertificate_Issuer:            cert.GetIssuer(),
		tenantpb.CaCertificate_Provider:          cert.GetProvider(),
		tenantpb.CaCertificate_SerialNumber:      cert.GetSerialNumber(),
		tenantpb.CaCertificate_FingerprintSha256: cert.GetFingerprintSha256(),
		tenantpb.CaCertificate_NotBefore:         cert.GetNotBefore().AsTime(),
		tenantpb.CaCertificate_NotAfter:          cert.GetNotAfter().AsTime(),
		tenantpb.CaCertificate_SelfSigned:        cert.GetSelfSigned(),
		tenantpb.CaCertificate_OwnerResourceId:   owner.GetResourceId(),
		tenantpb.CaCertificate_OwnerTenantId:     owner.GetTenantId(),
	}
}

// Delete implements ICaCertificate.
func (r *CaCertificates) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

// Get implements ICaCertificate.
func (r *CaCertificates) Get(
	ctx context.Context, id string) (*tenantpb.CaCertificate, error) {

	builder := r.selectQuery(nil).
		Where(sq.Eq{tenantpb.CaCertificate_Id: id})

	cert, err := r.storage.Select(ctx, builder,
		tenantpb.ScanOneCaCertificate)
	if err != nil {
		return nil, err
	}

	if err = r.open(cert); err != nil {
		return nil, err
	}

	return cert, nil
}

// List implements ICaCertificate.
func (r *CaCertificates) List(ctx context.Context,
	req *tenantpb.ListCaCertificateRequest,
) (*tenantpb.ListCaCertificateResponse, error) {

	conds := listConditions(req)

	certs, err := r.rows(ctx, req.GetPage(), conds)
	if err != nil {
		return nil, err
	}

	var total int64
	if req.GetPage().GetTotal() {
		total, err = r.total(ctx, conds)
		if err != nil {
			return nil, err
		}
	}

	return &tenantpb.ListCaCertificateResponse{
		CaCertificates: certs,
		Total:          total,
	}, nil
}

// rows reads the page of certificates matching conds, with their secrets
// unsealed.
func (r *CaCertificates) rows(ctx context.Context, page *typepb.Pages,
	conds []sq.Sqlizer) ([]*tenantpb.CaCertificate, error) {

	builder := r.selectQuery(page)
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	szlog.Debug("[cacertificates] query: ", query, " args: ", args)

	certs, err := r.storage.CollectRows(
		ctx, builder, tenantpb.ScanCaCertificate)
	if err != nil {
		return nil, err
	}

	for _, cert := range certs {
		if err = r.open(cert); err != nil {
			return nil, err
		}
	}

	return certs, nil
}

// open unseals the pem of cert, which has been read from the database. The
// private key stays sealed: nothing reads it back, so it is not unsealed.
func (r *CaCertificates) open(cert *tenantpb.CaCertificate) error {
	pem, err := r.sealer.Open(cert.GetPem(), cert.GetId())
	if err != nil {
		return fmt.Errorf("open pem of %s: %w", cert.GetId(), err)
	}

	cert.Pem = pem
	hideKey(cert)

	return nil
}

// hideKey leaves cert with the flag of its private key and not the key.
func hideKey(cert *tenantpb.CaCertificate) {
	cert.HasPrivateKey = cert.GetPrivateKey() != ""
	cert.PrivateKey = ""
}

// listConditions maps the filters of the request to WHERE conditions,
// all of them must match.
func listConditions(req *tenantpb.ListCaCertificateRequest) []sq.Sqlizer {
	var conds []sq.Sqlizer

	// a slice becomes owner_resource_id IN (...)
	if ids := req.GetResourceIds(); len(ids) != 0 {
		conds = append(conds,
			sq.Eq{tenantpb.CaCertificate_OwnerResourceId: ids})
	}

	if usage := req.GetUsage(); usage !=
		tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UNSPECIFIED {
		conds = append(conds,
			sq.Eq{tenantpb.CaCertificate_Usage: int32(usage)})
	}

	if fp := req.GetFingerprintSha256(); fp != "" {
		conds = append(conds,
			sq.Eq{tenantpb.CaCertificate_FingerprintSha256: fp})
	}

	if provider := req.GetProvider(); provider != "" {
		conds = append(conds,
			sq.Eq{tenantpb.CaCertificate_Provider: provider})
	}

	return conds
}

// total counts the rows matching the list conditions.
func (r *CaCertificates) total(
	ctx context.Context, conds []sq.Sqlizer,
) (int64, error) {
	builder := sq.Select("COUNT(*)").From(r.storage.Table())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	var total int64
	if err := r.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *CaCertificates) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		tenantpb.CaCertificateColumns()...)
}
