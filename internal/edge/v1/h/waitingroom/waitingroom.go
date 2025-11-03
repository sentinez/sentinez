package waitingroom

import (
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	"github.com/sentinez/sentinez/pkg/dmz/queue"
	httpxdmz "github.com/sentinez/sentinez/pkg/network/httpx/dmz"
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

func (wr *WaitingRoom) Handle(ctx *httpxdmz.Context) error {

	return wr.HandleNext(ctx)
}
