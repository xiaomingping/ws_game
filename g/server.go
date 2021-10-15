package g

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/xiaomingping/ws_game/peer/ws"
	_ "github.com/xiaomingping/ws_game/proc/ws"
)

type opt func(s Session, Message interface{})

type serve struct {
	address     string
	queue       cellnet.EventQueue
	genericPeer cellnet.GenericPeer
	Dispatcher  *proc.MessageDispatcher
}

func NewServe(address string) *serve {
	queue := cellnet.NewEventQueue()
	s := &serve{
		address:     address,
		queue:       queue,
		genericPeer: nil,
		Dispatcher:  nil,
	}
	s.queue.EnableCapturePanic(true)
	s.genericPeer = peer.NewGenericPeer("ws.Acceptor", "server", s.address, s.queue)
	s.Dispatcher = proc.NewMessageDispatcherBindPeer(s.genericPeer, "ws.ltv")
	return s
}

func (s *serve) AddCallback(msgName string, userCallback opt) {
	s.Dispatcher.RegisterMessage(msgName, func(ev cellnet.Event) {
		userCallback(ev.Session().(Session), ev.Message())
	})
}
func (s *serve) Start() {
	// 开始侦听
	s.genericPeer.Start()

	// 事件队列开始循环
	s.queue.StartLoop()

	// 阻塞等待事件队列结束退出( 在另外的goroutine调用queue.StopLoop() )
	s.queue.Wait()
}
