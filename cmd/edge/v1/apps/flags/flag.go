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
	"github.com/sentinez/sentinez/pkg/std/stdflag"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/spf13/pflag"
)

var onceEdge sync.Once

// Parse flag args for grpc service
func Parse() *common.Flag {
	onceEdge.Do(func() {
		stdflag.Get().RulePath = "./resources/waf/data/v4-16-0"
		stdflag.Get().ProxyConfig = "./cmd/edge/v1/proxy.yaml"
		stdflag.Get().EnvFile = "./cmd/edge/v1/.env"
		stdflag.Get().CertificateFile = "./localhost.pem"
		stdflag.Get().CertKeyFile = "./localhost-key.pem"

		pflag.StringVar(&stdflag.Get().EnvFile, "env-file",
			stdflag.Get().GetEnvFile(), "environment variables config file")

		pflag.StringVar(&stdflag.Get().RulePath, "rule-path",
			stdflag.Get().GetRulePath(), "core rulesets root path for rules")

		pflag.StringVar(&stdflag.Get().ProxyConfig, "proxy-config",
			stdflag.Get().GetProxyConfig(), "origin config yaml configuration")

		pflag.StringVar(&stdflag.Get().CertificateFile, "cert-file",
			stdflag.Get().GetCertificateFile(), "TLS certificate file .pem")

		pflag.StringVar(&stdflag.Get().CertKeyFile, "cert-key",
			stdflag.Get().GetCertKeyFile(), "TLS certificate key .pem")

		stdflag.Parse(edge.GetMetaEdge())
	})

	if err := stdflag.Validate(stdflag.Get()); err != nil {
		zlog.Fatal(err)
	}

	return stdflag.Get()
}
