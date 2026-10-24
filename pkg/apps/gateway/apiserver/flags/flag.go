// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package flags provides the app setting for apiserver service
package flags

import (
	"fmt"
	"sync"

	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/sentinez/api/proto/sentinez/gateway/apiserver/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szflag"
	"github.com/spf13/pflag"
)

var _onceAPIServer sync.Once

// Parse flag args for apiserver service
func Parse() (*settingpb.Flag, error) {
	_onceAPIServer.Do(func() {
		szflag.Get().EnvFile = "./cmd/szapiserver/.env"

		pflag.StringVar(&szflag.Get().EnvFile, settingpb.XFlag_EnvFile,
			szflag.Get().GetEnvFile(), "environment variable config file")

		// with a certificate the apiserver serves HTTPS and HTTP/3,
		// otherwise plain HTTP
		pflag.StringVar(&szflag.Get().CertFile, settingpb.XFlag_CertFile,
			szflag.Get().GetCertFile(), "TLS certificate file")

		pflag.StringVar(&szflag.Get().CertKeyFile, settingpb.XFlag_CertKeyFile,
			szflag.Get().GetCertKeyFile(), "TLS certificate key")

		szflag.Parse(szrunner.Banner(apiserver.GetMetaApiserver()))
	})

	if err := szflag.Validate(szflag.Get()); err != nil {
		return nil, fmt.Errorf("validate flags: %w", err)
	}

	return szflag.Get(), nil
}
