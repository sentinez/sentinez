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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/client/discovery"
	"github.com/sentinez/sentinez/pkg/client/local"
	"github.com/sentinez/sentinez/pkg/client/options"
)

func NewIAMService(opt *options.Options,
) (iam.IdentityAccessManagementServiceClient, error) {

	srv, err := discovery.GetDiscovery(opt).Discover(iam.GetMetaIamServiceKey())
	if err != nil {
		return nil, err
	}

	conn, err := Conn(srv.Address)
	if err != nil {
		return nil, err
	}

	return iam.NewIdentityAccessManagementServiceClient(conn), nil
}

func NewLocalIAMService(hdl iam.IdentityAccessManagementServiceServer,
) (iam.IdentityAccessManagementServiceClient, error) {

	bufLis := local.RegisterServiceServer(
		iam.GetMetaIamServiceKey(), hdl,
		iam.RegisterIdentityAccessManagementServiceServer)

	conn, err := BufConn(bufLis)
	if err != nil {
		return nil, err
	}

	return iam.NewIdentityAccessManagementServiceClient(conn), nil
}

func NewLocalGreeterService(hdl greeter.GreeterServiceServer,
) (greeter.GreeterServiceClient, error) {

	bufLis := local.RegisterServiceServer(
		greeter.GetMetaGreeterServiceKey(), hdl,
		greeter.RegisterGreeterServiceServer)

	conn, err := BufConn(bufLis)
	if err != nil {
		return nil, err
	}

	return greeter.NewGreeterServiceClient(conn), nil
}
