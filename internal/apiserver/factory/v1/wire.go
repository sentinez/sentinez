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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/infra/utils"

	accountrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
	iamprivateservice "github.com/sentinez/sentinez/internal/core/iam/v1/services/private"
	iampublicservice "github.com/sentinez/sentinez/internal/core/iam/v1/services/public"

	services "github.com/sentinez/sentinez/internal/apiserver/services/v1"

	greeterhandler "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"

	tenanthandler "github.com/sentinez/sentinez/internal/core/tenant/v1/handler"

	iamhandler "github.com/sentinez/sentinez/internal/core/iam/v1/handler"

	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
)

func NewDefaultGreeter() httpgw.ServiceRegistrar {
	wire.Build(
		greeterhandler.New,
		services.NewGreeter,
	)
	return nil
}

func NewDefaultIAM(conf *common.Config) (httpgw.ServiceRegistrar, error) {
	wire.Build(
		utils.NewPgxPool,
		usersrepo.New,
		accountrepo.New,
		iampublicservice.New,
		iamprivateservice.New,
		iamhandler.New,
		services.NewIAM,
	)
	return nil, nil
}

func NewDefaultTenant() httpgw.ServiceRegistrar {
	wire.Build(
		tenanthandler.New,
		services.NewTenant,
	)
	return nil
}
