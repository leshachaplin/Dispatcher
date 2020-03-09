package websocket

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"net/http"
)

type Websocket struct {
	OnMsg func(msg string)
	port  int
}

func New(onMessage func(msg string), ctx context.Context, port int, e *echo.Echo) (*Websocket, error) {

	w := &Websocket{
		OnMsg: onMessage,
		port:  port,
	}

	e.GET("/role", func(c echo.Context) error {
		return w.RoleWorker(ctx, c.Response(), c.Request())
	})
	return w, nil
}

func (w *Websocket) NewWebsocketConnection() (*websocket.Conn, error) {
	origin := fmt.Sprintf("http://localhost:%d/", w.port)
	url := fmt.Sprintf("ws://localhost:%d/role", w.port)
	webSocket, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	return webSocket, nil
}

func (w *Websocket) RoleWorker(ctx context.Context, wr http.ResponseWriter, req *http.Request) error {

	websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println("handle server connection")
		defer ws.Close()
		go func() {
			<-ctx.Done()
			ws.Close()
		}()
		for {
			msg := ""
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error %s", err))
			}

			w.OnMsg(msg)
		}
	}).ServeHTTP(wr, req)
	return nil
}
