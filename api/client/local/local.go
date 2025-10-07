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

package local

import (
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var serverOnce = map[string]*ServerOnce{}

const BufSize = 1024 * 1024

type ServerOnce struct {
	once sync.Once
	conn *bufconn.Listener
}

type R[T any] func(s grpc.ServiceRegistrar, srv T)

func RegisterServiceServer[T any](
	serviceKey string, hdl T, register R[T]) *bufconn.Listener {

	srvOnce, ok := serverOnce[serviceKey]
	if !ok {
		serverOnce[serviceKey] = &ServerOnce{}
		srvOnce = serverOnce[serviceKey]
	}

	srvOnce.once.Do(func() {
		s := grpc.NewServer()
		srvOnce.conn = bufconn.Listen(BufSize)

		register(s, hdl)

		go func() {
			_ = s.Serve(srvOnce.conn)
		}()
	})

	return srvOnce.conn
}
