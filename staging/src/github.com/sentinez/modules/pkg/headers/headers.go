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

	"github.com/sentinez/modules/pkg/crypto"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/perms"
	"google.golang.org/grpc/metadata"
)

const AuthHeader string = "Authorization"

type Auth struct {
	*typepb.Context
}

func (a *Auth) Check(method *typepb.XMethod) error {
	return perms.Allow(method, a.GetConsole())
}

func GetAuth(ctx context.Context) (*Auth, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	accessToken := md.Get(AuthHeader)
	if len(accessToken) == 0 {
		return nil, errorx.StatusUnauthorizedF("Invalid Access Token")
	}

	pl, ok := crypto.BearerTokenVerifier(accessToken[0])
	if !ok {
		return nil, errorx.StatusUnauthorizedF("Invalid Access Token")
	}

	return &Auth{Context: pl}, nil
}
