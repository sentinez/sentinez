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
	"time"

	"github.com/sentinez/sentinez/pkg/zlog"

	"github.com/sentinez/sentinez/api/client"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	greeterfac "github.com/sentinez/sentinez/internal/core/greeter/v1/factory"
	iamhdl "github.com/sentinez/sentinez/internal/core/iam/v1/handler"
	accountrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
	iamsvc "github.com/sentinez/sentinez/internal/core/iam/v1/services"
	"github.com/sentinez/sentinez/pkg/passkey"
	"github.com/sentinez/sentinez/pkg/storage/database/postgres"
)

// nolint:funlen
func NewDefaultHandler(appConf *common.AppConfig,
) iam.IdentityAccessManagementServiceServer {

	service := NewDefaultService(appConf)

	geeterCli, err := client.NewLocalGreeterService(
		greeterfac.NewDefaultHandler(appConf),
	)
	if err != nil {
		zlog.Errorf("iamfactory: new greeter client err=%v", err)
	}

	return iamhdl.New(service, geeterCli)
}

func NewDefaultService(appConf *common.AppConfig) *iamsvc.IAMService {
	userrepos, err := usersrepo.New(appConf)
	if err != nil {
		zlog.Errorf("iamfactory: init user repo err=%v", err)
	}

	accountrepos, err := accountrepo.New(appConf)
	if err != nil {
		zlog.Errorf("iamfactory: init account repo err=%v", err)
	}

	tx := postgres.NewTX(appConf)
	dataStore := passkey.NewMemoryStorage(time.Hour * 2)

	return iamsvc.New(appConf, tx, dataStore, userrepos, accountrepos)
}
