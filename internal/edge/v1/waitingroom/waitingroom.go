package waitingroom

import (
	"github.com/sentinez/sentinez/internal/common/chains"
	"github.com/sentinez/sentinez/internal/common/queue"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
)

var _ chains.Handler = (*WaitingRoom)(nil)

func New() *WaitingRoom {
	return &WaitingRoom{
		BaseHandler: chains.New(),
	}
}

type WaitingRoom struct {
	*chains.BaseHandler
	_ *queue.Queue
}

func (wr *WaitingRoom) Handle(ctx *httpxhz.Context) error {

	return wr.HandleNext(ctx)
}
