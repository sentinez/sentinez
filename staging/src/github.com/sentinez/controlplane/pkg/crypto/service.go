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
	"errors"
	"time"

	"github.com/sentinez/pkg/szconfig"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

const (
	// _serviceIssuer signs the service tokens and _serviceAudience is who
	// they are for: they differ from the user ones, so neither kind passes
	// for the other even if both secrets were the same.
	_serviceIssuer   = "sentinez.service"
	_serviceAudience = "sentinez.service"
)

// errServiceKey the SENZ_SERVICE_KEY is not set.
var errServiceKey = errors.New("crypto: service key is empty")

// ServiceTokenGenerator signs a token of the service name, valid for ttl,
// with SENZ_SERVICE_KEY. Only the operator tooling signs them: the services
// holding a token cannot make another.
func ServiceTokenGenerator(name string, ttl time.Duration) (string, error) {
	key := szconfig.GetEnv(settingpb.Senz_SENZ_SERVICE_KEY)
	if key == "" {
		return "", errServiceKey
	}

	payload := &typepb.Context{
		Name:         name,
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_SERVICE,
	}

	return signToken(key, payload, time.Now().Add(ttl),
		_serviceIssuer, _serviceAudience)
}

// ServiceTokenVerifier returns the service of a bearer token signed with
// SENZ_SERVICE_KEY. It fails for a token of any other control plane.
func ServiceTokenVerifier(bearerToken string) (*typepb.Context, bool) {
	key := szconfig.GetEnv(settingpb.Senz_SENZ_SERVICE_KEY)
	if key == "" {
		return nil, false
	}

	ctx, ok := verifyToken(bearerToken, key, _serviceIssuer,
		_serviceAudience)
	if !ok || ctx.GetControlPlane() !=
		typepb.ControlPlane_CONTROL_PLANE_SERVICE {
		return nil, false
	}

	return ctx, true
}
