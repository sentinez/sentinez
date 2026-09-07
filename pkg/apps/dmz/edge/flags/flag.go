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
	"sync"

	"github.com/sentinez/shared/flagx"

	"github.com/sentinez/shared/zlog"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/spf13/pflag"
)

var onceEdge sync.Once

// Parse flag args for grpc service
func Parse() *settingpb.Flag {
	onceEdge.Do(func() {
		flagx.Get().ProxyConfig = "./proxy.yaml"
		flagx.Get().EnvFile = "./.env"
		flagx.Get().CertFile = "./_wildcard.sentinez.vn+1.pem"
		flagx.Get().CertKeyFile = "./_wildcard.sentinez.vn+1-key.pem"

		pflag.StringVar(&flagx.Get().EnvFile, settingpb.XFlag_EnvFile,
			flagx.Get().GetEnvFile(), "environment variables config file")

		pflag.StringVar(&flagx.Get().ProxyConfig, settingpb.XFlag_ProxyConfig,
			flagx.Get().GetProxyConfig(), "origin config yaml configuration")

		pflag.StringVar(&flagx.Get().CertFile, settingpb.XFlag_CertFile,
			flagx.Get().GetCertFile(), "TLS certificate file")

		pflag.StringVar(&flagx.Get().CertKeyFile, settingpb.XFlag_CertKeyFile,
			flagx.Get().GetCertKeyFile(), "TLS certificate key")

		flagx.Parse(edgepb.GetMetaEdge())
	})

	if err := flagx.Validate(flagx.Get()); err != nil {
		zlog.Fatal(err)
	}

	return flagx.Get()
}
