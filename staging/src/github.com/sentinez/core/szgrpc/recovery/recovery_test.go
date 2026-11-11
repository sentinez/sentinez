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

package recovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func call(handler grpc.UnaryHandler) (any, error) {
	info := &grpc.UnaryServerInfo{FullMethod: "/svc.Service/Method"}

	return Unary()(context.Background(), nil, info, handler)
}

func TestUnary_Panic(t *testing.T) {
	var auth *struct{ Name string }

	resp, err := call(func(context.Context, any) (any, error) {
		return auth.Name, nil // nil dereference
	})

	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "nil pointer")
}

func TestUnary_Pass(t *testing.T) {
	resp, err := call(func(context.Context, any) (any, error) {
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}
