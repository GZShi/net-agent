module ws

go 1.15

require (
	github.com/gorilla/websocket v1.4.2
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	logger v0.0.0-00010101000000-000000000000
)

replace logger => ../logger
