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
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/pkg/szconfig"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/collector/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	_ "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_iam        = "/sentinez.apps.iam.v1.IdentityAccessManagementService/"
	_login      = _iam + "Login"
	_deleteUser = _iam + "DeleteUser"
	_getTenant  = "/sentinez.apps.tenant.v1.TenantService/GetTenant"

	_portal = typepb.ControlPlane_CONTROL_PLANE_PORTAL
	_admin  = typepb.ControlPlane_CONTROL_PLANE_ADMIN
)

// calls runs the interceptor for fullMethod as a caller sending token, and
// returns the context the handler saw, nil if the handler never ran.
func calls(token, fullMethod string) (context.Context, error) {
	ctx := context.Background()
	if token != "" {
		ctx = metadata.NewIncomingContext(ctx,
			metadata.Pairs(headers.AuthHeader, token))
	}

	var seen context.Context
	handler := func(ctx context.Context, _ any) (any, error) {
		seen = ctx
		return nil, nil
	}

	_, err := Unary()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: fullMethod},
		handler)

	return seen, err
}

func tokenOf(t *testing.T, plane typepb.ControlPlane) string {
	t.Helper()

	secret := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY, secret)

	token, err := crypto.TokenGenerator(&typepb.Context{
		Name:         "tester",
		UserId:       "user_1",
		ControlPlane: plane,
		ExpireAt:     timestamppb.New(time.Now().Add(time.Hour)),
	})
	require.NoError(t, err)

	return token
}

// methods splits every method of the sentinez.apps services by whether it
// is public.
func methods() (public, protected []string) {
	files := protoregistry.GlobalFiles
	files.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(string(f.Package()), "sentinez.apps.") {
			return true
		}

		for i := range f.Services().Len() {
			svc := f.Services().Get(i)
			for j := range svc.Methods().Len() {
				name := "/" + string(svc.FullName()) + "/" +
					string(svc.Methods().Get(j).Name())
				if methodOf(name).GetPublic() {
					public = append(public, name)
				} else {
					protected = append(protected, name)
				}
			}
		}

		return true
	})

	return public, protected
}

func TestUnary_ProtectedNeedsToken(t *testing.T) {
	_, protected := methods()
	require.NotEmpty(t, protected)

	for _, tt := range protected {
		t.Run(tt, func(t *testing.T) {
			seen, err := calls("", tt)

			assert.Equal(t, codes.Unauthenticated, status.Code(err))
			assert.Nil(t, seen, "the handler must not run")
		})
	}
}

func TestUnary_PublicNeedsNoToken(t *testing.T) {
	public, _ := methods()
	require.NotEmpty(t, public)

	for _, tt := range public {
		t.Run(tt, func(t *testing.T) {
			seen, err := calls("", tt)

			require.NoError(t, err)
			assert.NotNil(t, seen, "the handler must run")
		})
	}
}

func TestUnary_ControlPlane(t *testing.T) {
	tests := []struct {
		name       string
		givePlane  typepb.ControlPlane
		giveMethod string
		wantCode   codes.Code
	}{
		{
			name:       "portal on an admin-only method",
			givePlane:  _portal,
			giveMethod: _deleteUser,
			wantCode:   codes.PermissionDenied,
		},
		{
			name:       "admin on an admin-only method",
			givePlane:  _admin,
			giveMethod: _deleteUser,
			wantCode:   codes.OK,
		},
		{
			name:       "portal on a shared method",
			givePlane:  _portal,
			giveMethod: _getTenant,
			wantCode:   codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calls(tokenOf(t, tt.givePlane), tt.giveMethod)

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestUnary_BadToken(t *testing.T) {
	_, err := calls("Bearer not-a-token", _getTenant)

	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUnary_HandlerSeesCaller(t *testing.T) {
	token := tokenOf(t, _portal)

	seen, err := calls(token, _getTenant)
	require.NoError(t, err)

	auth, err := headers.GetAuth(seen)
	require.NoError(t, err)
	assert.Equal(t, "user_1", auth.GetUserId())
}

func TestUnary_PublicSeesValidCaller(t *testing.T) {
	token := tokenOf(t, _portal)

	seen, err := calls(token, _login)
	require.NoError(t, err)

	auth, err := headers.GetAuth(seen)
	require.NoError(t, err)
	assert.Equal(t, "user_1", auth.GetUserId())
}

func TestUnary_UnknownMethodNeedsToken(t *testing.T) {
	_, err := calls("", "/sentinez.apps.nothing.v1.Nothing/Nothing")

	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}
