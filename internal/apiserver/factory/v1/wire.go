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

package factory

import (
	"github.com/google/wire"

	services "github.com/sentinez/sentinez/internal/apiserver/services/v1"

	dcvrdomain "github.com/sentinez/sentinez/internal/core/discovery/v1/domain"
	dcvrhandler "github.com/sentinez/sentinez/internal/core/discovery/v1/handler"
	dcvrrepo "github.com/sentinez/sentinez/internal/core/discovery/v1/repos"

	greeterdomain "github.com/sentinez/sentinez/internal/core/greeter/v1/domain"
	greeterhandler "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"

	tenanthandler "github.com/sentinez/sentinez/internal/core/tenant/v1/handler"

	iamdomain "github.com/sentinez/sentinez/internal/core/iam/v1/domain"
	iamhandler "github.com/sentinez/sentinez/internal/core/iam/v1/handler"

	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
)

func NewDefaultDiscovery() httpgw.ServiceRegistrar {
	wire.Build(
		dcvrrepo.New,
		dcvrdomain.New,
		dcvrhandler.New,
		services.NewDiscovery,
	)
	return nil
}

func NewDefaultGreeter() httpgw.ServiceRegistrar {
	wire.Build(
		greeterdomain.New,
		greeterhandler.New,
		services.NewGreeter,
	)
	return nil
}

func NewDefaultIAM() httpgw.ServiceRegistrar {
	wire.Build(
		iamdomain.New,
		iamhandler.New,
		services.NewIAM,
	)
	return nil
}

func NewDefaultTenant() httpgw.ServiceRegistrar {
	wire.Build(
		tenanthandler.New,
		services.NewTenant,
	)
	return nil
}
