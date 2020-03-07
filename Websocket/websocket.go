package Websocket

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"strconv"
)

type Websocket struct {
	e        *echo.Echo
	WriteUrl string
	OnMsg    func(msg string, chanel chan interface{}) error
	Port     int
}

func New(onMessage func(msg string, chanel chan interface{}) error, ctx context.Context, port int) *Websocket {
	w := &Websocket{
		e:        echo.New(),
		WriteUrl: "/role",
		OnMsg:    onMessage,
		Port:     port,
	}

	w.e.GET("/role", func(c echo.Context) error {
		return w.RoleWorker(ctx, c, w.Port)
	})
	return w
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
			fmt.Println(fmt.Sprintf("i'am RoleWorker port%d", port))
			msg := ""
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error %s", err))
			}

			error := w.OnMsg
			if error != nil {
				log.Errorf("message not send to chanel", err)
			}

			fmt.Println(fmt.Sprintf("i'am port %d read message %s", port, msg))
			message, err := strconv.Atoi(msg)
			if port == message {
				fmt.Println(fmt.Sprintf(" serverport %d == %d ", port, message))
				fmt.Println("i am not watcher")
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
