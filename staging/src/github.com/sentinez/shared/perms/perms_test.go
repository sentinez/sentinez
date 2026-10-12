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

package perms

import (
	"testing"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
)

// nolint:funlen
func TestAllow(t *testing.T) {
	tests := []struct {
		name      string
		method    *typepb.XMethod
		plane     typepb.ControlPlane
		wantError bool
	}{
		{
			name:      "Nil method",
			method:    nil,
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: false,
		},
		{
			name: "Ignored method",
			method: &typepb.XMethod{
				Ignore: true,
			},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: false,
		},
		{
			name:      "Empty consoles list",
			method:    &typepb.XMethod{},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: false,
		},
		{
			name: "Console match",
			method: &typepb.XMethod{
				ControlPlanes: []typepb.ControlPlane{
					typepb.ControlPlane_CONTROL_PLANE_PORTAL,
				},
			},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: false,
		},
		{
			name: "Console mismatch",
			method: &typepb.XMethod{
				ControlPlanes: []typepb.ControlPlane{
					typepb.ControlPlane_CONTROL_PLANE_ADMIN,
				},
			},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: true,
		},
		{
			name: "Multiple consoles - match",
			method: &typepb.XMethod{
				ControlPlanes: []typepb.ControlPlane{
					typepb.ControlPlane_CONTROL_PLANE_PORTAL,
					typepb.ControlPlane_CONTROL_PLANE_ADMIN,
				},
			},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: false,
		},
		{
			name: "Multiple consoles - no match",
			method: &typepb.XMethod{
				ControlPlanes: []typepb.ControlPlane{
					typepb.ControlPlane_CONTROL_PLANE_ADMIN,
				},
			},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: true,
		},
		{
			name: "Ignored method with consoles still allows access",
			method: &typepb.XMethod{
				Ignore: true,
				ControlPlanes: []typepb.ControlPlane{
					typepb.ControlPlane_CONTROL_PLANE_ADMIN,
				},
			},
			plane:     typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Allow(tt.method, tt.plane)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
