package Websocket

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"strconv"
)

type Websocket struct {
	e *echo.Echo
	WriteUrl string
}

func New(onMessage func(msg string)) (*Websocket, error) {
	e.Get
}

func RoleWorker(ctx context.Context, c echo.Context, ) error {

	websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println("handle server connection")
		defer ws.Close()
		go func() {
			<-ctx.Done()
			ws.Close()
		}()
		for {
			select {
			case <-:
				{
					fmt.Println(fmt.Sprintf("i'am RoleWorker port%d", serverPort))
					msg := ""
					err := websocket.Message.Receive(ws, &msg)
					if err != nil {
						fmt.Println(fmt.Sprintf("Error %s", err))
						w.OnMes(msg)
					}

					fmt.Println(fmt.Sprintf("i'am port %d read message %s", serverPort, msg))
					message, err := strconv.Atoi(msg)
					if serverPort == message {
						fmt.Println(fmt.Sprintf("wathcher %v , serverport %d == %d ", isWatcher, serverPort, message))
						fmt.Println("i am not watcher")
					}
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}