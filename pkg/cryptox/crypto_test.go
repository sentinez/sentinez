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

package cryptox

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGenAndVerifyToken(t *testing.T) {
	secBase64 := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))

	conf := &common.EnvConfig{SecretKey: secBase64}

	token, err := TokenGenerator(conf, &common.Context{
		Name:     "test gen & verify",
		ExpireAt: timestamppb.New(time.Now().Add(time.Hour)),
	})
	if err != nil {
		t.Error(err)
	}

	tp, ok := BearerTokenVerifier(conf, token)
	if !ok {
		t.Error("fail to verify bearer token")
	}

	t.Log("token payload: ", tp.String())
}
