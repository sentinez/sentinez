// Copyright 2025 Sentinez Labs.
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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/stdcmn/zflag"
	"github.com/sentinez/sentinez/pkg/stdcmn/zlog"

	"github.com/spf13/pflag"
)

var onceEdge sync.Once

// Parse flag args for grpc service
func Parse() *common.Flag {
	onceEdge.Do(func() {
		zflag.Get().RulePath = "./resources/waf/data/v4-16-0"
		zflag.Get().ProxyConfig = "./cmd/edge/v1/proxy.yaml"
		zflag.Get().EnvFile = "./cmd/edge/v1/.env"
		zflag.Get().CertificateFile = "./localhost.pem"
		zflag.Get().CertKeyFile = "./localhost-key.pem"

		pflag.StringVar(&zflag.Get().EnvFile, "env-file",
			zflag.Get().GetEnvFile(), "environment variables config file")

		pflag.StringVar(&zflag.Get().RulePath, "rule-path",
			zflag.Get().GetRulePath(), "core rulesets root path for rules")

		pflag.StringVar(&zflag.Get().ProxyConfig, "proxy-config",
			zflag.Get().GetProxyConfig(), "origin config yaml configuration")

		pflag.StringVar(&zflag.Get().CertificateFile, "cert-file",
			zflag.Get().GetCertificateFile(), "TLS certificate file .pem")

		pflag.StringVar(&zflag.Get().CertKeyFile, "cert-key",
			zflag.Get().GetCertKeyFile(), "TLS certificate key .pem")

		zflag.Parse(edge.GetMetaEdge())
	})

	if err := zflag.Validate(zflag.Get()); err != nil {
		zlog.Fatal(err)
	}

	return zflag.Get()
}
