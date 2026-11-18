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

package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/sentinez/controlplane/pkg/auditowner"
	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/pkg/szlog"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	secpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	auditpb "github.com/sentinez/sentinez/api/proto/sentinez/types/audit/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	_iamService   = "/sentinez.apps.iam.v1.IdentityAccessManagementService/"
	_loginMethod  = _iamService + "Login"
	_statusMethod = _iamService + "Status"

	_createSecRuleMethod = "/sentinez.apps.security.v1.SecurityService/" +
		"CreateSecRule"
)

// recorder is a szlog.Log keeping the events logged at info.
type recorder struct {
	szlog.Log
	entries []*auditpb.AuditLog
}

func (r *recorder) Info(_ string, event proto.Message) {
	r.entries = append(r.entries, proto.Clone(event).(*auditpb.AuditLog))
}

func invoke(t *testing.T, r *recorder, method string, req proto.Message,
	err error) {
	t.Helper()

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("x-forwarded-for", "6.6.6.6, 203.0.113.7",
			"grpcgateway-user-agent", "curl/8"))
	ctx = headers.WithAuth(ctx, &headers.Auth{})

	_, got := Audit(r)(ctx, req, &grpc.UnaryServerInfo{FullMethod: method},
		func(context.Context, any) (any, error) { return nil, err })
	require.Equal(t, err, got)
}

func TestAuditRecordsAuditedMethod(t *testing.T) {
	r := &recorder{}
	req := &iampb.LoginRequest{EmailOrUsername: "bob", Password: "hunter22"}

	invoke(t, r, _loginMethod, req, nil)

	require.Len(t, r.entries, 1)
	got := r.entries[0]
	assert.Equal(t, _loginMethod, got.GetMethod())
	assert.Equal(t, uint32(codes.OK), got.GetCode())
	assert.Equal(t, "203.0.113.7", got.GetRemoteAddr())
	assert.Equal(t, "curl/8", got.GetUserAgent())
	assert.Contains(t, got.GetId(), "aud_")
	assert.Equal(t, "bob",
		got.GetRequest().GetFields()["emailOrUsername"].GetStringValue())
	assert.Equal(t, "REDACTED",
		got.GetRequest().GetFields()["password"].GetStringValue())
	assert.Equal(t, "hunter22", req.GetPassword(), "request kept")
}

func TestAuditRecordsCheckedOwner(t *testing.T) {
	r := &recorder{}
	ctx := headers.WithAuth(context.Background(), &headers.Auth{})
	req := &secpb.CreateSecRuleRequest{SecRule: &secpb.SecRule{
		Owner: &typepb.Owner{TenantId: "tnt_req", ResourceId: "res_req"},
	}}
	info := &grpc.UnaryServerInfo{FullMethod: _createSecRuleMethod}

	_, err := Audit(r)(ctx, req, info,
		func(ctx context.Context, _ any) (any, error) {
			auditowner.Set(ctx, "tnt_db", "")
			return nil, nil
		})

	require.NoError(t, err)
	require.Len(t, r.entries, 1)
	assert.Equal(t, "tnt_db", r.entries[0].GetTenantId())
	assert.Equal(t, "res_req", r.entries[0].GetResourceId())
}

func TestAuditRecordsFailedCall(t *testing.T) {
	r := &recorder{}
	err := status.Error(codes.PermissionDenied, "nope")

	invoke(t, r, _loginMethod, &iampb.LoginRequest{}, err)

	require.Len(t, r.entries, 1)
	assert.Equal(t, uint32(codes.PermissionDenied), r.entries[0].GetCode())
	assert.Equal(t, "nope", r.entries[0].GetError())
}

func TestAuditSkipsMethodWithoutAudit(t *testing.T) {
	tests := []struct {
		name       string
		giveMethod string
	}{
		{name: "no audit option", giveMethod: _statusMethod},
		{name: "unknown method", giveMethod: "/unknown.Service/Method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}

			invoke(t, r, tt.giveMethod, &iampb.StatusRequest{}, errors.New("x"))

			assert.Empty(t, r.entries)
		})
	}
}

// auditAuthorized runs an audited call through Audit then Authorize, as the
// api server chains them, with token as the bearer token.
func auditAuthorized(t *testing.T, token string) *auditpb.AuditLog {
	t.Helper()

	r := &recorder{}
	ctx := metadata.NewIncomingContext(t.Context(),
		metadata.Pairs(headers.AuthHeader, token))
	info := &grpc.UnaryServerInfo{FullMethod: _loginMethod}
	handler := func(ctx context.Context, req any) (any, error) {
		return Authorize()(ctx, req, info,
			func(context.Context, any) (any, error) { return nil, nil })
	}

	_, err := Audit(r)(ctx, &iampb.LoginRequest{}, info, handler)
	require.NoError(t, err)
	require.Len(t, r.entries, 1)

	return r.entries[0]
}

func TestAuditRecordsCallerAuthorizeSaw(t *testing.T) {
	got := auditAuthorized(t,
		tokenOf(t, typepb.ControlPlane_CONTROL_PLANE_PORTAL))

	assert.Equal(t, "user_1", got.GetUserId())
	assert.Equal(t, "tester", got.GetName())
}

func TestAuditRecordsNoCallerWithBadToken(t *testing.T) {
	got := auditAuthorized(t, "not-a-token")

	assert.Empty(t, got.GetUserId())
}
