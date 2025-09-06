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

package iamfac

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	iamhandler "github.com/sentinez/sentinez/internal/core/iam/v1/handler"
	accountrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
	iamservices "github.com/sentinez/sentinez/internal/core/iam/v1/services"
	"github.com/sentinez/sentinez/pkg/infra/database/postgres"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// nolint:funlen
func NewDefaultIAMHdl(
	rctx *common.RunnerCtx) iam.IdentityAccessManagementServiceServer {

	userrepos, err := usersrepo.New(rctx)
	if err != nil {
		zlog.Errorf("iamfactory: init user repo err=%v", err)
	}

	accountrepos, err := accountrepo.New(rctx)
	if err != nil {
		zlog.Errorf("iamfactory: init account repo err=%v", err)
	}

	tx := postgres.NewTX(rctx.GetConfig())
	service := iamservices.New(rctx.GetConfig(), tx, userrepos, accountrepos)

	return iamhandler.New(service)
}
