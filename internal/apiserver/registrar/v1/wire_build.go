//go:build wireinject
// +build wireinject

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

package registrar

import (
	"github.com/google/wire"
	dcvrhandler "github.com/sentinez/sentinez/internal/core/discovery/v1/handler"
	greeterhandler "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"
	iamhandlers "github.com/sentinez/sentinez/internal/core/iam/v1/handlers"
	tenanthandler "github.com/sentinez/sentinez/internal/core/tenant/v1/handler"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
)

func NewDefaultDiscovery() httpgw.ServiceRegistrar {
	wire.Build(
		dcvrhandler.NewDefault,
		NewDiscovery,
	)
	return nil
}

func NewDefaultGreeter() httpgw.ServiceRegistrar {
	wire.Build(
		greeterhandler.NewDefault,
		NewGreeter,
	)
	return nil
}

func NewDefaultIAM() httpgw.ServiceRegistrar {
	wire.Build(
		iamhandlers.NewDefault,
		NewIAM,
	)
	return nil
}

func NewDefaultTenant() httpgw.ServiceRegistrar {
	wire.Build(
		tenanthandler.NewDefault,
		NewTenant,
	)
	return nil
}
