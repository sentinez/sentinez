package waitingroom

import (
	corehttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	"github.com/sentinez/sentinez/pkg/dmz/queue"
)

var _ chains.Handler = (*WaitingRoom)(nil)

func New() chains.Handler {
	return &WaitingRoom{
		BaseHandler: chains.New(),
	}
}

type WaitingRoom struct {
	*chains.BaseHandler
	_ *queue.Queue
}

func (wr *WaitingRoom) Handle(ctx corehttp.Context) error {

	return wr.HandleNext(ctx)
}
