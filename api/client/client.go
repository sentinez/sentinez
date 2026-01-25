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

package client

import (
	"github.com/sentinez/sentinez/api/client/connection"
	"github.com/sentinez/sentinez/api/client/discovery"
	"github.com/sentinez/sentinez/api/client/local"
	"github.com/sentinez/sentinez/api/client/options"
	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

func NewIAM(opt *options.Options,
) (iampb.IdentityAccessManagementServiceClient, error) {

	srv, err := discovery.GetDiscovery(opt).Discover(iampb.GetMetaIamServiceKey())
	if err != nil {
		return nil, err
	}

	conn, err := connection.Conn(srv.Address)
	if err != nil {
		return nil, err
	}

	return iampb.NewIdentityAccessManagementServiceClient(conn), nil
}

func NewLocalIAM(hdl iampb.IdentityAccessManagementServiceServer,
) (iampb.IdentityAccessManagementServiceClient, error) {

	bufLis := local.RegisterServiceServer(
		iampb.GetMetaIamServiceKey(), hdl,
		iampb.RegisterIdentityAccessManagementServiceServer)

	conn, err := connection.BufConn(bufLis)
	if err != nil {
		return nil, err
	}

	return iampb.NewIdentityAccessManagementServiceClient(conn), nil
}

func NewLocalEdgeEngine(hdl edgepb.EdgeEngineServiceServer,
) (edgepb.EdgeEngineServiceClient, error) {
	bufLis := local.RegisterServiceServer(
		edgepb.GetMetaEdgeServiceKey(), hdl,
		edgepb.RegisterEdgeEngineServiceServer)

	conn, err := connection.BufConn(bufLis)
	if err != nil {
		return nil, err
	}

	return edgepb.NewEdgeEngineServiceClient(conn), nil
}

func NewLocalGreeter(hdl greeterpb.GreeterServiceServer,
) (greeterpb.GreeterServiceClient, error) {

	bufLis := local.RegisterServiceServer(
		greeterpb.GetMetaGreeterServiceKey(), hdl,
		greeterpb.RegisterGreeterServiceServer)

	conn, err := connection.BufConn(bufLis)
	if err != nil {
		return nil, err
	}

	return greeterpb.NewGreeterServiceClient(conn), nil
}
