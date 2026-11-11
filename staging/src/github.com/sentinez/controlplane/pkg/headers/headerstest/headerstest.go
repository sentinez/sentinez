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

// Package headerstest builds the contexts of callers for tests of the
// services, as the authorization interceptor would.
package headerstest

import (
	"context"

	"github.com/sentinez/controlplane/pkg/headers"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// User returns ctx with userId signed in to the portal.
func User(ctx context.Context, userId string) context.Context {
	return as(ctx, userId, typepb.ControlPlane_CONTROL_PLANE_PORTAL)
}

// Admin returns ctx with the admin control plane signed in.
func Admin(ctx context.Context) context.Context {
	return as(ctx, "sentinez.admin", typepb.ControlPlane_CONTROL_PLANE_ADMIN)
}

func as(ctx context.Context, userId string, plane typepb.ControlPlane,
) context.Context {
	return headers.WithAuth(ctx, &headers.Auth{
		Context: &typepb.Context{UserId: userId, ControlPlane: plane},
	})
}
