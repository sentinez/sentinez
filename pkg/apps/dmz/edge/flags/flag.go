// Copyright 2025 Sentinéz Labs.
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

// Package edgeflags provides the app setting for apiserver service
package edgeflags

import (
	"fmt"
	"sync"

	"github.com/sentinez/core/szrunner"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szflag"
	"github.com/spf13/pflag"
)

var _onceEdge sync.Once

// Parse flag args for grpc service
func Parse() (*settingpb.Flag, error) {
	_onceEdge.Do(func() {
		szflag.Get().ProxyConfig = "./proxy.yaml"
		szflag.Get().EnvFile = "./.env"
		szflag.Get().CertFile = "./_wildcard.sentinez.vn+1.pem"
		szflag.Get().CertKeyFile = "./_wildcard.sentinez.vn+1-key.pem"
		szflag.Get().Engine = "quic"

		pflag.StringVar(&szflag.Get().EnvFile, settingpb.XFlag_EnvFile,
			szflag.Get().GetEnvFile(), "environment variables config file")

		pflag.StringVar(&szflag.Get().ProxyConfig, settingpb.XFlag_ProxyConfig,
			szflag.Get().GetProxyConfig(), "origin config yaml configuration")

		pflag.StringVar(&szflag.Get().CertFile, settingpb.XFlag_CertFile,
			szflag.Get().GetCertFile(), "TLS certificate file")

		pflag.StringVar(&szflag.Get().CertKeyFile, settingpb.XFlag_CertKeyFile,
			szflag.Get().GetCertKeyFile(), "TLS certificate key")

		pflag.StringVar(&szflag.Get().Engine, settingpb.XFlag_Engine,
			szflag.Get().GetEngine(), "HTTP server engine")

		szflag.Parse(szrunner.Banner(edgepb.GetMetaEdge()))
	})

	if err := szflag.Validate(szflag.Get()); err != nil {
		return nil, fmt.Errorf("validate flags: %w", err)
	}

	return szflag.Get(), nil
}
