package Websocket

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

type Websocket struct {
	Echo     *echo.Echo
	WriteUrl string
	OnMsg    func(msg string)
	Port     int
}

func New(onMessage func(msg string), ctx context.Context, port int) (*Websocket, error) {

	w := &Websocket{
		Echo:     echo.New(),
		WriteUrl: "/role",
		OnMsg:    onMessage,
		Port:     port,
	}

	w.Echo.GET("/role", func(c echo.Context) error {
		return w.RoleWorker(ctx, c, w.Port)
	})
	return w, nil
}

func (w *Websocket) NewWebsocketConnection() (*websocket.Conn, error) {
	origin := fmt.Sprintf("http://localhost:%d/", w.Port)
	url := fmt.Sprintf("ws://localhost:%d/role", w.Port)
	webSocket, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	return webSocket, nil
}

func (w *Websocket) RoleWorker(ctx context.Context, c echo.Context, port int) error {

	websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println("handle server connection")
		defer ws.Close()
		go func() {
			<-ctx.Done()
			ws.Close()
		}()
		for {
			fmt.Println(fmt.Sprintf("i'am RoleWorker port %d", port))
			msg := ""
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error %s", err))
			}

			w.OnMsg(msg)
			fmt.Println("i am not watcher")
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
