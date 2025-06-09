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

package apiserver

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"

	dcvrhandler "github.com/sentinez/sentinez/internal/core/discovery/v1/handler"
	greeterhandler "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"
)

func TestHandler(_ *testing.T) {
	mux := runtime.NewServeMux()

	_ = discoverypb.RegisterDiscoveryServiceHandlerServer(
		context.Background(), mux, dcvrhandler.NewDefault())

	_ = greeterpb.RegisterGreeterServiceHandlerServer(
		context.Background(), mux, greeterhandler.NewDefault())
}
