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

package headers

import (
	"context"
	"net"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	_forwardedFor = "x-forwarded-for"
	_userAgent    = "user-agent"
	// _gatewayUserAgent is the User-Agent of the HTTP request, which
	// grpc-gateway forwards under its metadata prefix.
	_gatewayUserAgent = runtime.MetadataPrefix + _userAgent
)

// ClientAddr returns the address of the client: the last of the
// X-Forwarded-For the gateway sets, which is the peer it saw, not what the
// client wrote in front of it; else the host of the peer, else "unknown".
func ClientAddr(ctx context.Context) string {
	forwarded := metadata.ValueFromIncomingContext(ctx, _forwardedFor)
	if len(forwarded) != 0 {
		hops := strings.Split(forwarded[len(forwarded)-1], ",")
		if addr := strings.TrimSpace(hops[len(hops)-1]); addr != "" {
			return addr
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			return host
		}
	}

	return "unknown"
}

// UserAgent returns the user agent of the client: the one of the HTTP
// request grpc-gateway forwarded, else the one of the gRPC client, else "".
func UserAgent(ctx context.Context) string {
	for _, key := range [...]string{_gatewayUserAgent, _userAgent} {
		if v := metadata.ValueFromIncomingContext(ctx, key); len(v) != 0 {
			return v[0]
		}
	}

	return ""
}
