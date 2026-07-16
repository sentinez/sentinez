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
	"context"
	"time"

	"github.com/sentinez/core/storage/dbx/postgres"
	greeterfac "github.com/sentinez/modules/greeter/v1/factory"
	iamhdl "github.com/sentinez/modules/iam/v1/handler"
	accountrepo "github.com/sentinez/modules/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/modules/iam/v1/repos/users"
	iamsvc "github.com/sentinez/modules/iam/v1/service"
	"github.com/sentinez/modules/pkg/passkey"
	"github.com/sentinez/sentinez/api/client"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/shared/zlog"
)

// nolint:funlen
func NewDefaultHandler(ctx context.Context, appConf *settingpb.Config,
) iampb.IdentityAccessManagementServiceServer {

	service := NewDefaultService(ctx, appConf)

	geeterCli, err := client.NewLocalGreeter(
		greeterfac.NewDefaultHandler(appConf),
	)
	if err != nil {
		zlog.Errorf("iamfactory: new greeter client err=%v", err)
	}

	return iamhdl.New(service, geeterCli)
}

func NewDefaultService(
	ctx context.Context, appConf *settingpb.Config) *iamsvc.IAMService {
	userrepos, err := usersrepo.New(ctx, appConf)
	if err != nil {
		zlog.Errorf("iamfactory: init user repo err=%v", err)
	}

	accountrepos, err := accountrepo.New(ctx, appConf)
	if err != nil {
		zlog.Errorf("iamfactory: init account repo err=%v", err)
	}

	tx := postgres.NewTX(appConf)
	dataStore := passkey.NewMemoryStorage(time.Hour)

	return iamsvc.New(appConf, tx, dataStore, userrepos, accountrepos)
}
