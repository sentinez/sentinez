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

package controlplane

import (
	"github.com/sentinez/core/szgrpc/grpcgateway"
	"github.com/sentinez/sentinez/pkg/interceptor"
	"google.golang.org/grpc"
)

// chain returns the options of a local conn: the unary interceptors
// before, then the authorization of the unary calls and of the streams. An
// interceptor before it, e.g. the audit log, also sees the calls that
// authorization denies.
func chain(
	before []grpc.UnaryServerInterceptor,
) []grpcgateway.LocalOption {
	return append(
		[]grpcgateway.LocalOption{grpcgateway.WithUnaryInterceptors(before...)},
		interceptor.LocalOptions()...)
}
