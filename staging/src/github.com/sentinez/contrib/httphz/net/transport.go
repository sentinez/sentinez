package httphznet

import (
	"context"

	"github.com/cloudwego/hertz/pkg/network"
)

type Transporter struct {
	network.Transporter
}

func (t *Transporter) ListenAndServe(onData network.OnData) error {
	return t.Transporter.ListenAndServe(
		func(ctx context.Context, conn any) error {
			return onData(ctx, conn)
		},
	)
}
