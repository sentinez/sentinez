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

package engine

import (
	"github.com/sentinez/contrib/httphz"
	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/transport"
	quichttpx "github.com/sentinez/sentinez/pkg/network/httpx/quic"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
)

func Hertz(conf *settingpb.Config) (corehttp.Server, corehttp.ServerOption) {
	return httphz.NewServer(conf),
		corehttp.WithOnStdConnect(transport.OnHertzConnect)
}

func Standard(conf *settingpb.Config) (corehttp.Server, corehttp.ServerOption) {
	return stdhttpx.NewServer(conf),
		corehttp.WithOnStdConnect(transport.OnStandardConnect)
}

func Quic(conf *settingpb.Config) (corehttp.Server, corehttp.ServerOption) {
	return quichttpx.NewServer(conf),
		corehttp.WithOnStdConnect(transport.OnStandardConnect)
}
