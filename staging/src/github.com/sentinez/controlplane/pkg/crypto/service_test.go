// Copyright 2025 Duc-Hung Ho.
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

package crypto

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/sentinez/pkg/szconfig"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setKeys(t *testing.T, secret, service string) {
	t.Helper()

	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY,
		base64.StdEncoding.EncodeToString([]byte(secret)))
	szconfig.SetEnv(settingpb.Senz_SENZ_SERVICE_KEY,
		base64.StdEncoding.EncodeToString([]byte(service)))
}

func TestServiceTokenRoundTrip(t *testing.T) {
	setKeys(t, "user-secret", "service-secret")

	token, err := ServiceTokenGenerator("edge-1", time.Hour)
	require.NoError(t, err)

	got, ok := ServiceTokenVerifier("Bearer " + token)

	require.True(t, ok)
	assert.Equal(t, "edge-1", got.GetName())
	assert.Equal(t, typepb.ControlPlane_CONTROL_PLANE_SERVICE,
		got.GetControlPlane())
}

func TestServiceTokenExpired(t *testing.T) {
	setKeys(t, "user-secret", "service-secret")

	token, err := ServiceTokenGenerator("edge-1", -time.Minute)
	require.NoError(t, err)

	_, ok := ServiceTokenVerifier(token)

	assert.False(t, ok)
}

func TestServiceTokenSignedWithOtherKey(t *testing.T) {
	setKeys(t, "user-secret", "service-secret")
	token, err := ServiceTokenGenerator("edge-1", time.Hour)
	require.NoError(t, err)

	setKeys(t, "user-secret", "another-secret")
	_, ok := ServiceTokenVerifier(token)

	assert.False(t, ok)
}

func TestServiceTokenNeedsKey(t *testing.T) {
	setKeys(t, "user-secret", "")

	_, err := ServiceTokenGenerator("edge-1", time.Hour)
	_, ok := ServiceTokenVerifier("x")

	require.ErrorIs(t, err, errServiceKey)
	assert.False(t, ok)
}

// The two kinds of token never pass for each other, even when the
// operator reused the same key for both.
func TestTokenKindsAreSeparate(t *testing.T) {
	setKeys(t, "same-secret", "same-secret")

	user, err := TokenGenerator(&typepb.Context{
		Name:         "root",
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_ADMIN,
		ExpireAt:     timestamppb.New(time.Now().Add(time.Hour)),
	})
	require.NoError(t, err)
	service, err := ServiceTokenGenerator("edge-1", time.Hour)
	require.NoError(t, err)

	_, userAsService := ServiceTokenVerifier(user)
	_, serviceAsUser := BearerTokenVerifier(service)

	assert.False(t, userAsService)
	assert.False(t, serviceAsUser)
}
