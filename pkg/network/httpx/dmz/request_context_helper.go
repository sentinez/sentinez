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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"

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

	msgBin, _ := proto.Marshal(msg)

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
	if err := proto.Unmarshal(msgBin, &msg); err != nil {
		return nil, false
	}

	return &msg, true
}

func setRequestTime(ctx context.Context) context.Context {
	// set request time
	ctx = context.WithValue(ctx, senzRequestHTTPTimeKey, time.Now().UTC())

	return ctx
}

// GenContextKey nolint:funlen
func GenContextKey(ctx *Context) string {
	method := string(ctx.req.Method())
	host := string(ctx.req.Host())
	path := string(ctx.Path())

	args := ctx.req.QueryArgs()
	var keys []string
	args.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	sort.Strings(keys)

	sortedQuery := ""
	for _, k := range keys {
		sortedQuery += fmt.Sprintf("%s=%s&", k, args.Peek(k))
	}

	ct := string(ctx.req.Request.Header.ContentType())

	body := ctx.req.Request.Body()
	if len(body) > 1024 {
		body = body[:1024]
	}
	bodyHash := ""
	if len(body) > 0 {
		sum := sha256.Sum256(body)
		bodyHash = hex.EncodeToString(sum[:])
	}

	rawKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		method, host, path, sortedQuery, ct, bodyHash,
	)

	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}
