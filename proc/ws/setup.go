package ws

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
)

func init() {

	proc.RegisterProcessor("ws.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(WSMessageTransmitter))
		bundle.SetHooker(new(MsgHooker))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}
