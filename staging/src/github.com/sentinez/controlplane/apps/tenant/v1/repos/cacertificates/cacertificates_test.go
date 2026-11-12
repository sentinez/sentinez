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
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/crypto"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_pem = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"
	_key = "-----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY-----\n"
)

const _usageClient = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_CLIENT

func newRepo(t *testing.T, b byte) *CaCertificates {
	t.Helper()

	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{b}, 32))
	sealer, err := crypto.NewSealer(key)
	require.NoError(t, err)

	return &CaCertificates{sealer: sealer}
}

func newCert(privateKey string) *tenantpb.CaCertificate {
	return &tenantpb.CaCertificate{
		Id: "cacrt_1", Name: "Acme Root CA", Pem: _pem,
		PrivateKey: privateKey,
	}
}

// readBack is the row the database gives back for the values of a create.
func readBack(values map[string]any) *tenantpb.CaCertificate {
	return &tenantpb.CaCertificate{
		Id:         values[tenantpb.CaCertificate_Id].(string),
		Pem:        values[tenantpb.CaCertificate_Pem].(string),
		PrivateKey: values[tenantpb.CaCertificate_PrivateKey].(string),
	}
}

func TestInsertValuesSealTheSecrets(t *testing.T) {
	repo := newRepo(t, 1)

	values, err := repo.insertValues(newCert(_key))

	require.NoError(t, err)
	for column, value := range values {
		text, ok := value.(string)
		if !ok {
			continue
		}

		assert.NotContains(t, text, "BEGIN", column)
	}

	assert.True(t, strings.HasPrefix(
		values[tenantpb.CaCertificate_Pem].(string), "v1:"))
	assert.True(t, strings.HasPrefix(
		values[tenantpb.CaCertificate_PrivateKey].(string), "v1:"))
	assert.Equal(t, "Acme Root CA", values[tenantpb.CaCertificate_Name])
}

func TestOpenUnsealsThePemAndNotTheKey(t *testing.T) {
	repo := newRepo(t, 1)
	values, err := repo.insertValues(newCert(_key))
	require.NoError(t, err)
	row := readBack(values)

	err = repo.open(row)

	require.NoError(t, err)
	assert.Equal(t, _pem, row.GetPem())
	assert.Empty(t, row.GetPrivateKey())
	assert.True(t, row.GetHasPrivateKey())
}

func TestHideKey(t *testing.T) {
	cert := newCert(_key)

	hideKey(cert)

	assert.Empty(t, cert.GetPrivateKey())
	assert.True(t, cert.GetHasPrivateKey())
}

func TestOpenWithoutPrivateKey(t *testing.T) {
	repo := newRepo(t, 1)
	values, err := repo.insertValues(newCert(""))
	require.NoError(t, err)
	row := readBack(values)

	err = repo.open(row)

	require.NoError(t, err)
	assert.Empty(t, values[tenantpb.CaCertificate_PrivateKey])
	assert.Equal(t, _pem, row.GetPem())
	assert.Empty(t, row.GetPrivateKey())
	assert.False(t, row.GetHasPrivateKey())
}

func TestOpenFails(t *testing.T) {
	values, err := newRepo(t, 1).insertValues(newCert(_key))
	require.NoError(t, err)

	moved := readBack(values)
	moved.Id = "cacrt_2" // the secrets of another record

	tests := []struct {
		name     string
		giveRepo *CaCertificates
		giveRow  *tenantpb.CaCertificate
	}{
		{"another key", newRepo(t, 2), readBack(values)},
		{"another record", newRepo(t, 1), moved},
		{"stored in clear", newRepo(t, 1), newCert(_key)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.giveRepo.open(tt.giveRow)

			assert.ErrorIs(t, err, crypto.ErrOpen)
		})
	}
}

func TestListConditionsWithoutFilter(t *testing.T) {
	conds := listConditions(&tenantpb.ListCaCertificateRequest{})

	assert.Empty(t, conds)
}

func TestListConditions(t *testing.T) {
	conds := listConditions(&tenantpb.ListCaCertificateRequest{
		ResourceIds:       []string{"rsrc_1", "rsrc_2"},
		Usage:             _usageClient,
		FingerprintSha256: "abc",
		Provider:          "Acme Trust",
	})

	query, args, err := sq.And(conds).ToSql()

	require.NoError(t, err)
	assert.Equal(t, "(owner_resource_id IN (?,?) AND usage = ? AND "+
		"fingerprint_sha256 = ? AND provider = ?)", query)
	assert.Equal(t,
		[]any{"rsrc_1", "rsrc_2", int32(2), "abc", "Acme Trust"}, args)
}
