// Copyright 2026 Sentinéz Labs.
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

package network

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type connIdKey struct{}

// ConnId returns the connection id that the QUIC listener put on the
// connection context, or "" when ctx has none.
func ConnId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	id, _ := ctx.Value(connIdKey{}).(string)

	return id
}

func withConnId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, connIdKey{}, id)
}

func newConnId() string {
	id := make([]byte, 8)
	_, _ = rand.Read(id)

	return hex.EncodeToString(id)
}
