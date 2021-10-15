# ws_game
websocket 轻量级游戏框架

#demo
```
const (
	TestAddress = "http://127.0.0.1:8080/ws"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	s := g.NewServe(TestAddress)
	s.AddCallback("cellnet.SessionAccepted", Open)
	s.AddCallback("cellnet.SessionClosed", Closed)
	s.Start()
}

func Open(s g.Session, Message interface{}) {
	s.(cellnet.ContextSet).SetContext("ping","222")
	zap.S().Info("open",s.ID())
}

func Closed(s g.Session,message interface{})  {
	zap.S().Info("Closed",s.ID())
}
```