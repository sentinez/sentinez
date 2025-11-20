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

package corehttp

import (
	"context"
	"time"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

// SentinezContextKey is the key type for the context.
type SentinezContextKey string

const (
	RequestHTTPCtxKey  SentinezContextKey = "senz.ctx.request.http"
	RequestHTTPTimeKey SentinezContextKey = "senz.ctx.request.time"
)

// SetRequestContext returns a new context with the given message
func SetRequestContext(ctx Context, msg *edgepb.Context) Context {

	if msg == nil {
		msg = &edgepb.Context{}
	}

	if ctx == nil {
		return nil
	}

	ctx.SetExtra(msg)
	return ctx
}

// GetRequestContext returns the context value.
func GetRequestContext(rctx Context) (*edgepb.Context, bool) {
	msg := rctx.Extra()
	if msg == nil {
		return nil, false
	}

	return msg, true
}

func SetRequestTime(ctx context.Context) context.Context {
	// set request time
	ctx = context.WithValue(ctx, RequestHTTPTimeKey, time.Now().UTC())

	return ctx
}

func GetRequestTime(ctx context.Context) time.Time {
	t, ok := ctx.Value(RequestHTTPTimeKey).(time.Time)
	if ok {
		return t
	}
	return time.Time{}
}
