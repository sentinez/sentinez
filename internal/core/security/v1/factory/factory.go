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

package securityfac

import (
	"context"

	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/security/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	securityhdl "github.com/sentinez/sentinez/internal/core/security/v1/handler"
	"github.com/sentinez/sentinez/internal/core/security/v1/repos/rulegroups"
	"github.com/sentinez/sentinez/internal/core/security/v1/repos/rules"
	securitysvc "github.com/sentinez/sentinez/internal/core/security/v1/service"
	"github.com/sentinez/shared/zlog"
)

// NewDefaultHandler initializes the Security handler with its dependencies.
func NewDefaultHandler(ctx context.Context,
	appConf *confpb.Config,
) securitypb.SecurityServiceServer {
	service := NewDefaultService(ctx, appConf)
	return securityhdl.New(service)
}

// NewDefaultService initializes the Security service.
func NewDefaultService(ctx context.Context,
	appConf *confpb.Config,
) *securitysvc.SecurityService {
	rulesRepo, err := rules.New(ctx, appConf)
	if err != nil {
		zlog.Fatalf("failed to init rules repo: %v", err)
	}

	rulegroupsRepo, err := rulegroups.New(ctx, appConf)
	if err != nil {
		zlog.Fatalf("failed to init rulegroups repo: %v", err)
	}

	return securitysvc.New(appConf, rulesRepo, rulegroupsRepo)
}
