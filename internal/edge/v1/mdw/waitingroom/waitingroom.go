package waitingroom

import (
	"github.com/sentinez/sentinez/internal/edge/v1/chains"
	"github.com/sentinez/sentinez/internal/edge/v1/queue"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
)

func New() *WaitingRoom {
	return &WaitingRoom{
		Base: &chains.Base{},
	}
}

type WaitingRoom struct {
	*chains.Base
	_ *queue.Queue
}

func (wr *WaitingRoom) Handle(ctx *httpxhz.Context) error {
	return wr.HandleNext(ctx)
}
