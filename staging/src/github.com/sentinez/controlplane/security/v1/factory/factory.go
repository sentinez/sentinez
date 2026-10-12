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
	"fmt"

	securityhdl "github.com/sentinez/controlplane/security/v1/handler"
	"github.com/sentinez/controlplane/security/v1/repos/coreruleset"
	"github.com/sentinez/controlplane/security/v1/repos/ratelimit"
	"github.com/sentinez/controlplane/security/v1/repos/secrule"
	securitysvc "github.com/sentinez/controlplane/security/v1/service"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// NewDefaultHandler initializes the Security handler with its dependencies.
func NewDefaultHandler(ctx context.Context,
	appConf *settingpb.Config,
) (securitypb.SecurityServiceServer, error) {
	service, err := NewDefaultService(ctx, appConf)
	if err != nil {
		return nil, err
	}

	return securityhdl.New(service), nil
}

// NewDefaultService initializes the Security service.
func NewDefaultService(ctx context.Context,
	appConf *settingpb.Config,
) (*securitysvc.SecurityService, error) {
	secRuleRepo, err := secrule.New(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("secrule repo: %w", err)
	}

	rateLimitRepo, err := ratelimit.New(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("ratelimit repo: %w", err)
	}

	coreRulesetRepo, err := coreruleset.New(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("coreruleset repo: %w", err)
	}

	return securitysvc.New(
		appConf, secRuleRepo, rateLimitRepo, coreRulesetRepo,
	), nil
}
