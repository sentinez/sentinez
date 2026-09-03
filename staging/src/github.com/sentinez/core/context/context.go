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

package corecontext

import (
	"context"

	networkpb "github.com/sentinez/sentinez/api/gen/go/sentinez/network/v1"
)

type contextKey string

const (
	transport contextKey = "senz.transport"
)

func WithTransportValue(
	parent context.Context, value *networkpb.Transport) context.Context {
	return context.WithValue(parent, transport, value)
}

func GetTransport(ctx context.Context) *networkpb.Transport {
	v, ok := ctx.Value(transport).(*networkpb.Transport)
	if !ok {
		return nil
	}

	return v
}
