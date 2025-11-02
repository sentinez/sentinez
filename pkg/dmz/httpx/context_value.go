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

package httpxdmz

import (
	"context"

	"github.com/sentinez/sentinez/pkg/x/protobuf/protox"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

// SentinezContextKey is the key type for the context.
type SentinezContextKey string

const (
	senzRequestHTTPCtxKey  SentinezContextKey = "senz.ctx.request.http"
	senzRequestHTTPTimeKey SentinezContextKey = "senz.ctx.request.time"
)

// SetRequestContext returns a new context with the given message
func SetRequestContext(parent *Context, msg *edgepb.Context) *Context {
	if msg == nil {
		msg = &edgepb.Context{}
	}

	if parent == nil {
		return nil
	}

	msgBin, _ := protox.Marshal(msg)

	parent.ctx = context.WithValue(parent.ctx, senzRequestHTTPCtxKey, msgBin)
	return parent
}

// GetRequestContext returns the context value.
func GetRequestContext(rctx *Context) (*edgepb.Context, bool) {
	msgBin, ok := rctx.ctx.Value(senzRequestHTTPCtxKey).([]byte)
	if !ok {
		return nil, false
	}

	var msg edgepb.Context
	if err := protox.Unmarshal(msgBin, &msg); err != nil {
		return nil, false
	}

	return &msg, true
}
