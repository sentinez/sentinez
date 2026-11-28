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
	"github.com/sentinez/core/szgrpc/client/connection"
	"github.com/sentinez/core/szgrpc/client/discovery"
	"github.com/sentinez/core/szgrpc/client/identity"
	"github.com/sentinez/core/szgrpc/grpcgateway"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	greeterpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	"google.golang.org/grpc"
)

func NewIAM(opts ...discovery.Option,
) (iampb.IdentityAccessManagementServiceClient, error) {

	addr, err := discover(iampb.GetMetaIamServiceKey(), opts)
	if err != nil {
		return nil, err
	}

	conn, err := connection.Conn(addr)
	if err != nil {
		return nil, err
	}

	return iampb.NewIdentityAccessManagementServiceClient(conn), nil
}

// NewCentralData returns the client of centraldata found in consul, and the
// connection to close. tls and token are those of the internal mTLS and of
// the service token; without TLS the call is plaintext.
func NewCentralData(tls connection.TLS, token string,
	opts ...discovery.Option,
) (centraldatapb.CentralDataServiceClient, *grpc.ClientConn, error) {

	addr, err := discover(centraldatapb.GetMetaCentraldataServiceKey(), opts)
	if err != nil {
		return nil, nil, err
	}

	return DialCentralData(addr, tls, token)
}

// DialCentralData is NewCentralData for a known address. The server
// certificate is verified against identity.CentralData.
func DialCentralData(addr string, tls connection.TLS, token string,
) (centraldatapb.CentralDataServiceClient, *grpc.ClientConn, error) {

	opts, err := tls.DialOptions(identity.CentralData)
	if err != nil {
		return nil, nil, err
	}

	opts = append(opts, connection.WithBearer(token, tls.Enabled()))

	conn, err := connection.ConnWith(addr, opts...)
	if err != nil {
		return nil, nil, err
	}

	return centraldatapb.NewCentralDataServiceClient(conn), conn, nil
}

// NewLocalIAM returns a client of hdl over a grpcgateway.LocalConn.
// NewLocalGreeter returns a client of hdl over a grpcgateway.LocalConn.
func NewLocalGreeter(hdl greeterpb.GreeterServiceServer,
) greeterpb.GreeterServiceClient {
	conn := grpcgateway.NewLocalConn()
	greeterpb.RegisterGreeterServiceServer(conn, hdl)

	return greeterpb.NewGreeterServiceClient(conn)
}

// NewCommunication returns the client of the communication service found
// in consul.
func NewCommunication(opts ...discovery.Option,
) (communicationpb.CommunicationServiceClient, error) {
	addr, err := discover(
		communicationpb.GetMetaCommunicationServiceKey(), opts)
	if err != nil {
		return nil, err
	}

	conn, err := connection.Conn(addr)
	if err != nil {
		return nil, err
	}

	return communicationpb.NewCommunicationServiceClient(conn), nil
}

// NewLocalCommunication returns a client of hdl over a
// grpcgateway.LocalConn, and the conn to close: it ends the open streams.
// opts are those of the conn, e.g. the interceptors that authorize the
// calls as on the network.
func NewLocalCommunication(hdl communicationpb.CommunicationServiceServer,
	opts ...grpcgateway.LocalOption,
) (communicationpb.CommunicationServiceClient, *grpcgateway.LocalConn) {
	conn := grpcgateway.NewLocalConn(opts...)
	communicationpb.RegisterCommunicationServiceServer(conn, hdl)

	return communicationpb.NewCommunicationServiceClient(conn), conn
}

// discover returns the address of an instance of serviceKey in the consul
// of opts.
func discover(serviceKey string, opts []discovery.Option) (string, error) {
	dcvr, err := discovery.GetDiscovery(opts...)
	if err != nil {
		return "", err
	}

	srv, err := dcvr.Discover(serviceKey)
	if err != nil {
		return "", err
	}

	return srv.Address, nil
}
