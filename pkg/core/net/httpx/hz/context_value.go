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

package httpxhz

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/protobuf/proto"
)

// SentinezContextKey is the key type for the context.
type SentinezContextKey string

const (
	sntzRequestHTTPCtxKey  SentinezContextKey = "sntz.ctx.request.http"
	sntzRequestHTTPIDKey   SentinezContextKey = "sntz.ctx.request.id"
	sntzRequestHTTPTimeKey SentinezContextKey = "sntz.ctx.request.time"
)

// SetRequestContext returns a new context with the given message
func SetRequestContext(parent *Context, msg *common.HTTPContext) *Context {
	if msg == nil {
		msg = &common.HTTPContext{}
	}

	if parent == nil {
		return nil
	}

	msgBin, _ := proto.Marshal(msg)

	parent.ctx = context.WithValue(parent.ctx, sntzRequestHTTPCtxKey, msgBin)
	return parent
}

// GetRequestContext returns the context value.
func GetRequestContext(rctx *Context) (*common.HTTPContext, bool) {
	msgBin, ok := rctx.ctx.Value(sntzRequestHTTPCtxKey).([]byte)
	if !ok {
		return nil, false
	}

	var msg common.HTTPContext
	if err := proto.Unmarshal(msgBin, &msg); err != nil {
		return nil, false
	}

	return &msg, true
}
