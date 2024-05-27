package utils

import (
	// "fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var wsDialer = websocket.Dialer{
	Proxy:            http.ProxyFromEnvironment,
	HandshakeTimeout: 45 * time.Second,
}

func NewWebSocketClient(url string) (*websocket.Conn, *http.Response, error) {
	return wsDialer.Dial(url, nil)
}

func Subscribe(ws *websocket.Conn, symbol string) error {
	color.Green(`{"method":"SUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)
	// fmt.Println(`{"method":"SUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)
	subscribeMsg := []byte(`{"method":"SUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)

	return ws.WriteMessage(websocket.TextMessage, subscribeMsg)

}

func Unsubscribe(ws *websocket.Conn, symbol string) error {
	color.Red(`{"method":"UNSUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)
	// fmt.Println(`{"method":"UNSUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)
	subscribeMsg := []byte(`{"method":"UNSUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)

	return ws.WriteMessage(websocket.TextMessage, subscribeMsg)
}
