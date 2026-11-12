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

package connection

import (
	"context"
	"net"

	"github.com/sentinez/sentinez/api/interceptor/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func Conn(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(validate.UnaryClientInterceptor()),
	)

	return conn, err
}

func BufConn(bufLis *bufconn.Listener) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(validate.UnaryClientInterceptor()),
		grpc.WithContextDialer(
			func(ctx context.Context, _ string) (net.Conn, error) {
				return bufLis.DialContext(ctx)
			},
		),
	)

	return conn, err
}

// ConnWith is Conn with the dial options opts instead of the plaintext
// credentials, ex: TLS.DialOptions and WithBearer.
func ConnWith(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append([]grpc.DialOption{
		grpc.WithChainUnaryInterceptor(validate.UnaryClientInterceptor()),
	}, opts...)

	return grpc.NewClient(addr, opts...)
}

// WithBearer sends token as the Authorization header of every call. The
// token travels only over a secure transport unless secure is false, which
// is for the local plaintext development.
func WithBearer(token string, secure bool /* requireTLS */) grpc.DialOption {
	return grpc.WithPerRPCCredentials(bearer{token: token, secure: secure})
}

type bearer struct {
	token  string
	secure bool
}

func (b bearer) GetRequestMetadata(context.Context, ...string,
) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + b.token}, nil
}

func (b bearer) RequireTransportSecurity() bool { return b.secure }
