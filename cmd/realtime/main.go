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

package main

import (
	"github.com/sentinez/core/runner"
	"github.com/sentinez/sentinez"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/cmd/realtime/apps/config"
	"github.com/sentinez/sentinez/internal/realtime"
	wscore "github.com/sentinez/sentinez/pkg/network/wsz"
)

func main() {
	app := runner.NewApp(config.Config(), sentinez.Code)
	app.Run(func(conf *confpb.Config) error {
		var (
			wsSrv = wscore.NewServer(conf.GetMeta())
			rt    = realtime.New(wsSrv)
		)

		runner.Register(rt.Start, rt.Shutdown)

		return nil
	})
}
