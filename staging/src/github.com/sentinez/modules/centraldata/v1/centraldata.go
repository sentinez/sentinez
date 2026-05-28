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

package centraldata

import (
	"context"

	coregrpc "github.com/sentinez/core/grpc"
	centraldatafac "github.com/sentinez/modules/centraldata/v1/factory"
	"github.com/sentinez/sentinez/api/client/local"
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/centraldata/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"google.golang.org/grpc/test/bufconn"
)

var bufLis *bufconn.Listener

func GetListener() *bufconn.Listener {
	return bufLis
}

func NewService(ctx context.Context, appConf *confpb.Config) *CentralData {
	return &CentralData{
		Server: coregrpc.NewDefault(appConf),
		hdl:    centraldatafac.NewDefaultHandler(ctx, appConf),
	}
}

type CentralData struct {
	*coregrpc.Server
	hdl pb.CentralDataServiceServer
}

func (mod *CentralData) Start(_ context.Context) error {
	pb.RegisterCentralDataServiceServer(mod.AsServer(), mod.hdl)

	bufLis = bufconn.Listen(local.BufSize)
	return mod.BufServe(bufLis)
}
