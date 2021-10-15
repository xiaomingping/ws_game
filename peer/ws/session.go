package ws

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/util"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Socket会话
type WsSession struct {
	peer.CoreContextSet
	peer.CoreSessionIdentify
	*peer.CoreProcBundle

	pInterface cellnet.Peer

	conn *websocket.Conn

	// 退出同步器
	exitSync sync.WaitGroup

	// 发送队列
	sendQueue *cellnet.Pipe

	cleanupGuard sync.Mutex

	HeartbeatTime time.Time

	endNotify func()
}

func (self *WsSession) Peer() cellnet.Peer {
	return self.pInterface
}

// 取原始连接
func (self *WsSession) Raw() interface{} {
	if self.conn == nil {
		return nil
	}

	return self.conn
}

func (self *WsSession) Close() {
	self.sendQueue.Add(nil)
}

// 发送封包
func (self *WsSession) Send(msg interface{}) {
	self.sendQueue.Add(msg)
}

// 接收循环
func (self *WsSession) recvLoop() {

	for self.conn != nil {
		msg, err := self.ReadMessage(self)
		if err != nil {
			if !util.IsEOFOrNetReadError(err) {
				zap.S().Info("session closed:", err)
			}
			self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: &cellnet.SessionClosed{}})
			break
		}
		self.ProcEvent(&cellnet.RecvMsgEvent{Ses: self, Msg: msg})
	}

	self.Close()

	// 通知完成
	self.exitSync.Done()
}

// 发送循环
func (self *WsSession) sendLoop() {
	var writeList []interface{}
	for {
		writeList = writeList[0:0]
		exit := self.sendQueue.Pick(&writeList)
		// 遍历要发送的数据
		for _, msg := range writeList {
			self.SendMessage(&cellnet.SendMsgEvent{Ses: self, Msg: msg})
		}
		if exit {
			break
		}
	}
	// 关闭连接
	if self.conn != nil {
		self.conn.Close()
		self.conn = nil
	}
	// 通知完成
	self.exitSync.Done()
}

// 启动会话的各种资源
func (self *WsSession) Start() {
	// 将会话添加到管理器
	self.Peer().(peer.SessionManager).Add(self)
	// 需要接收和发送线程同时完成时才算真正的完成
	self.exitSync.Add(2)
	go func() {
		// 等待2个任务结束
		self.exitSync.Wait()

		// 将会话从管理器移除
		self.Peer().(peer.SessionManager).Remove(self)

		if self.endNotify != nil {
			self.endNotify()
		}

	}()
	// 启动并发接收goroutine
	go self.recvLoop()
	// 启动并发发送goroutine
	go self.sendLoop()
}

// 设置心跳时间
func (self *WsSession) SetPingTime() {
	self.cleanupGuard.Lock()
	defer self.cleanupGuard.Unlock()
	self.HeartbeatTime = time.Now().Add(time.Second * time.Duration(60))
}

/**
心跳超时
*/
func (self *WsSession) IsHeartbeatTimeout() (timeout bool) {
	self.cleanupGuard.Lock()
	defer self.cleanupGuard.Unlock()
	if time.Now().After(self.HeartbeatTime) {
		timeout = true
	}
	return
}

func newSession(conn *websocket.Conn, p cellnet.Peer, endNotify func()) *WsSession {
	self := &WsSession{
		conn:          conn,
		endNotify:     endNotify,
		sendQueue:     cellnet.NewPipe(),
		pInterface:    p,
		HeartbeatTime: time.Now(),
		CoreProcBundle: p.(interface {
			GetBundle() *peer.CoreProcBundle
		}).GetBundle(),
	}
	return self
}
