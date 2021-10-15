package ws

import (
	"github.com/davyxu/cellnet"
	"sync"
	"sync/atomic"
	"time"
)

type CoreSessionManager struct {
	sesById sync.Map // 使用Id关联会话

	sesIDGen int64 // 记录已经生成的会话ID流水号

	count int64 // 记录当前在使用的会话数量
}

func (self *CoreSessionManager) SetIDBase(base int64) {

	atomic.StoreInt64(&self.sesIDGen, base)
}

func (self *CoreSessionManager) Count() int {
	return int(atomic.LoadInt64(&self.count))
}

func (self *CoreSessionManager) Add(ses cellnet.Session) {

	id := atomic.AddInt64(&self.sesIDGen, 1)

	atomic.AddInt64(&self.count, 1)

	ses.(interface {
		SetID(int64)
	}).SetID(id)

	self.sesById.Store(id, ses)
}

func (self *CoreSessionManager) Remove(ses cellnet.Session) {

	self.sesById.Delete(ses.ID())

	atomic.AddInt64(&self.count, -1)
}

// 获得一个连接
func (self *CoreSessionManager) GetSession(id int64) cellnet.Session {
	if v, ok := self.sesById.Load(id); ok {
		return v.(cellnet.Session)
	}

	return nil
}

func (self *CoreSessionManager) VisitSession(callback func(cellnet.Session) bool) {

	self.sesById.Range(func(key, value interface{}) bool {

		return callback(value.(cellnet.Session))

	})
}

func (self *CoreSessionManager) CloseAllSession() {

	self.VisitSession(func(ses cellnet.Session) bool {

		ses.Close()

		return true
	})
}

// 活跃的会话数量
func (self *CoreSessionManager) SessionCount() int {

	v := atomic.LoadInt64(&self.count)

	return int(v)
}

// 心跳检测
func (self *CoreSessionManager) PingAuth() {
	ticker := time.NewTicker(time.Second * time.Duration(60))
	for {
		select {
		case <-ticker.C:
			self.sesById.Range(func(key, value interface{}) bool {
				s := value.(*WsSession)
				if s.IsHeartbeatTimeout() {
					s.Close()
				}
				return true
			})
		}
	}

}
