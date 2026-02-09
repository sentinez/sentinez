package httphz

import (
	"context"

	"github.com/cloudwego/hertz/pkg/network"
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

			return onData(ctx, conn)
		},
	)
}
