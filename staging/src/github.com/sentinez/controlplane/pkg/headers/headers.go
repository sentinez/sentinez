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

package headers

import (
	"context"
	"slices"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szperms"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/grpc/metadata"
)

const AuthHeader string = "Authorization"

type authKey struct{}

type Auth struct {
	*typepb.Context
}

// Check reports whether the caller may use method. It fails with
// PermissionDenied when the control plane of the caller is not allowed.
//
// A service is stricter than a user: it may call only the methods that list
// the service control plane, never one open to every control plane.
func (a *Auth) Check(method *typepb.XMethod) error {
	if a.IsService() && !slices.Contains(
		method.GetControlPlanes(), typepb.ControlPlane_CONTROL_PLANE_SERVICE,
	) && !method.GetPublic() {
		return szerror.StatusForbiddenF("access denied: service not allowed")
	}

	if err := szperms.Allow(method, a.GetControlPlane()); err != nil {
		return szerror.StatusForbiddenF("%v", err)
	}

	return nil
}

// IsAdmin reports whether the caller signed in to the admin control plane.
func (a *Auth) IsAdmin() bool {
	return a.GetControlPlane() == typepb.ControlPlane_CONTROL_PLANE_ADMIN
}

// IsService reports whether the caller is an internal service.
func (a *Auth) IsService() bool {
	return a.GetControlPlane() == typepb.ControlPlane_CONTROL_PLANE_SERVICE
}

// WithAuth returns a context that carries auth for GetAuth.
func WithAuth(ctx context.Context, auth *Auth) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

// GetAuth returns the caller of ctx: the one the authorization interceptor
// verified, otherwise the one of the bearer token in the metadata.
func GetAuth(ctx context.Context) (*Auth, error) {
	if auth, ok := ctx.Value(authKey{}).(*Auth); ok {
		return auth, nil
	}

	md, _ := metadata.FromIncomingContext(ctx)
	accessToken := md.Get(AuthHeader)
	if len(accessToken) == 0 {
		return nil, szerror.StatusUnauthorizedF("Invalid Access Token")
	}

	if pl, ok := crypto.BearerTokenVerifier(accessToken[0]); ok {
		return &Auth{Context: pl}, nil
	}

	if pl, ok := crypto.ServiceTokenVerifier(accessToken[0]); ok {
		return &Auth{Context: pl}, nil
	}

	return nil, szerror.StatusUnauthorizedF("Invalid Access Token")
}
