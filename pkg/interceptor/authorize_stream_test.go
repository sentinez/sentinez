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

package authz

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"testing"
	"time"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/pkg/szconfig"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const _sync = "/sentinez.apps.centraldata.v1.CentralDataService/Sync"

// fakeStream a server stream whose context is ctx.
type fakeStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s fakeStream) Context() context.Context { return s.ctx }

// streams runs the stream interceptor for fullMethod as a caller sending
// token, and returns the context the handler saw, nil if it never ran.
func streams(token, fullMethod string) (context.Context, error) {
	ctx := context.Background()
	if token != "" {
		ctx = metadata.NewIncomingContext(ctx,
			metadata.Pairs(headers.AuthHeader, token))
	}

	var seen context.Context
	handler := func(_ any, ss grpc.ServerStream) error {
		seen = ss.Context()
		return nil
	}

	err := Stream()(nil, fakeStream{ctx: ctx},
		&grpc.StreamServerInfo{FullMethod: fullMethod}, handler)

	return seen, err
}

func serviceTokenOf(t *testing.T) string {
	t.Helper()

	szconfig.SetEnv(settingpb.Senz_SENZ_SERVICE_KEY,
		base64.StdEncoding.EncodeToString([]byte("service-secret")))

	token, err := crypto.ServiceTokenGenerator("edge-1", time.Hour)
	require.NoError(t, err)

	return token
}

func TestStream_NeedsToken(t *testing.T) {
	seen, err := streams("", _sync)

	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Nil(t, seen, "the handler must not run")
}

func TestStream_UserTokenDenied(t *testing.T) {
	for _, plane := range []string{"portal", "admin"} {
		t.Run(plane, func(t *testing.T) {
			p := _portal
			if plane == "admin" {
				p = _admin
			}

			seen, err := streams(tokenOf(t, p), _sync)

			assert.Equal(t, codes.PermissionDenied, status.Code(err))
			assert.Nil(t, seen)
		})
	}
}

func TestStream_ServiceToken(t *testing.T) {
	seen, err := streams(serviceTokenOf(t), _sync)
	require.NoError(t, err)

	auth, err := headers.GetAuth(seen)
	require.NoError(t, err)
	assert.Equal(t, "edge-1", auth.GetName())
	assert.True(t, auth.IsService())
}

// A service token opens only the methods that list the service.
func TestUnary_ServiceTokenDenied(t *testing.T) {
	token := serviceTokenOf(t)

	for _, method := range []string{_deleteUser, _getTenant} {
		t.Run(method, func(t *testing.T) {
			seen, err := calls(token, method)

			assert.Equal(t, codes.PermissionDenied, status.Code(err))
			assert.Nil(t, seen)
		})
	}
}

// withCertificate the context of a connection whose verified client
// certificate is of cn.
func withCertificate(ctx context.Context, cn string) context.Context {
	cert := &x509.Certificate{Subject: pkix.Name{CommonName: cn}}
	info := credentials.TLSInfo{State: tls.ConnectionState{
		VerifiedChains: [][]*x509.Certificate{{cert}},
	}}

	return peer.NewContext(ctx, &peer.Peer{AuthInfo: info})
}

func streamsOn(ctx context.Context, token string) error {
	ctx = metadata.NewIncomingContext(ctx,
		metadata.Pairs(headers.AuthHeader, token))
	handler := func(any, grpc.ServerStream) error { return nil }

	return Stream()(nil, fakeStream{ctx: ctx},
		&grpc.StreamServerInfo{FullMethod: _sync}, handler)
}

func TestStream_CertificateOfTheToken(t *testing.T) {
	token := serviceTokenOf(t)

	err := streamsOn(withCertificate(context.Background(), "edge-1"), token)

	require.NoError(t, err)
}

func TestStream_CertificateOfAnotherService(t *testing.T) {
	token := serviceTokenOf(t)

	err := streamsOn(withCertificate(context.Background(), "edge-2"), token)

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
