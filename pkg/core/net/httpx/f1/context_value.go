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

// Package httpxf1 provides the HTTP context interface and its implementation.
package httpxf1

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/protobuf/proto"
)

// SentinezContextKey is the key type for the context.
type SentinezContextKey int

const contextKey SentinezContextKey = 0

// NewCtxValue returns a new context with the given message
func NewCtxValue(msg *common.Context) context.Context {
	if msg == nil {
		msg = &common.Context{}
	}

	ctx := context.Background()
	msgBin, _ := proto.Marshal(msg)

	return context.WithValue(ctx, contextKey, msgBin)
}

// Value returns the context value.
func Value(ctx context.Context) (*common.Context, bool) {
	msgBin, ok := ctx.Value(contextKey).([]byte)
	if !ok {
		return nil, false
	}

	var msg common.Context
	if err := proto.Unmarshal(msgBin, &msg); err != nil {
		return nil, false
	}

	return &msg, true
}
