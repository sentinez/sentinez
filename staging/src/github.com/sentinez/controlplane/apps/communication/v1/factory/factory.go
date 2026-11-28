// Copyright 2026 Duc-Hung Ho.
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

// Package communicationfac builds the communication handler.
package communicationfac

import (
	"context"
	"fmt"

	communicationhdl "github.com/sentinez/controlplane/apps/communication/v1/handler"
	notificationrepo "github.com/sentinez/controlplane/apps/communication/v1/repos/notifications"
	tenantfac "github.com/sentinez/controlplane/apps/tenant/v1/factory"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// NewDefaultHandler returns the communication handler of the setting; its
// Notify stores a notification and pushes it to the open streams.
func NewDefaultHandler(ctx context.Context,
	conf *settingpb.Config) (*communicationhdl.Communication, error) {
	notifications, err := notificationrepo.New(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("notification repo: %w", err)
	}

	az, err := tenantfac.NewAuthorizer(ctx, conf)
	if err != nil {
		return nil, err
	}

	return communicationhdl.New(notifications, az), nil
}
