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

package stdctx

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/std/stdcrypto"
	"github.com/sentinez/sentinez/pkg/std/stderr"
	"google.golang.org/grpc/metadata"
)

const AuthHeader string = "Authorization"

func GetAuthContext(ctx context.Context,
	conf *common.EnvConfig) (*common.Context, error) {

	md, _ := metadata.FromIncomingContext(ctx)
	accessToken := md.Get(AuthHeader)
	if len(accessToken) == 0 {
		return nil, stderr.UnauthorizedF("Invalid Access Token")
	}

	pl, ok := stdcrypto.BearerTokenVerifier(conf, accessToken[0])
	if !ok {
		return nil, stderr.UnauthorizedF("Invalid Access Token")
	}

	return pl, nil
}
