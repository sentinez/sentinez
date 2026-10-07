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

// Package greeter implements the Greeter service server.
package greeter

import (
	greeterhdl "github.com/sentinez/controlplane/greeter/v1/handler"
	coregrpc "github.com/sentinez/core/grpc"
	greeterpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

func NewService(conf *settingpb.Config) *Greeter {
	return &Greeter{
		Server:  coregrpc.New(),
		handler: greeterhdl.New(),
		conf:    conf,
	}
}

// Greeter implements GreeterServiceServer.
type Greeter struct {
	*coregrpc.Server
	handler greeterpb.GreeterServiceServer
	conf    *settingpb.Config
}

func (g *Greeter) Start() error {
	greeterpb.RegisterGreeterServiceServer(g.AsServer(), g.handler)

	return g.Serve(g.conf, coregrpc.WithDiscorvery(true))
}
