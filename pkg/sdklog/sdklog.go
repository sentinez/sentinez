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

package sdklog

import (
	"context"
	"fmt"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/zlog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// NewProvider returns the provider exporting the traffic logs of the func
// chains to SENZ_OTLP_EDGE_ENDPOINT, or nil when it is unset: those
// logs are then not exported.
func NewProvider(conf *settingpb.Config) (*sdklog.LoggerProvider, error) {
	endpoint := conf.Get(settingpb.Senz_SENZ_OTLP_EDGE_ENDPOINT)
	if endpoint == "" {
		return nil, nil
	}

	provider, err := zlog.NewOTLPProvider(context.Background(),
		zlog.OTLPConfigOf(endpoint, conf.GetMeta().GetServiceKey()))
	if err != nil {
		return nil, fmt.Errorf("edge log: %w", err)
	}

	return provider, nil
}
