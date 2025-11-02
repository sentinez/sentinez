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

package httpxdmz

import (
	"context"
	"reflect"

	"github.com/cloudwego/hertz/pkg/network"
	"github.com/sentinez/sentinez/pkg/zlog"
)

type TransporterKey string

const TransCtxKey TransporterKey = "trans_ctx_key"

type Transporter struct {
	network.Transporter
}

func (t *Transporter) ListenAndServe(onData network.OnData) error {
	return t.Transporter.ListenAndServe(
		func(ctx context.Context, conn any) error {
			// if tlsConn, ok := conn.(*standard.TLSConn); ok {
			// 	state := tlsConn.ConnectionState()
			// }

			t := reflect.TypeOf(conn)
			zlog.Debugf("[httpxdmz][transp] type of conn: %s", t.String())

			return onData(ctx, conn)
		},
	)
}
