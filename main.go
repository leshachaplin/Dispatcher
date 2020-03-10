package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	"github.com/leshachaplin/Dispatcher/config"
	"github.com/leshachaplin/Dispatcher/dispatcher"
	Kafka "github.com/leshachaplin/communicationUtils/kafka"
	Websocket "github.com/leshachaplin/communicationUtils/websocket"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var (
	isWatcher bool
)

type Message struct {
	Time string `json:"time"`
}

func main() {
	log.Info(isWatcher)
	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	done, cnsl := context.WithCancel(context.Background())
	e := echo.New()
	cfg := config.NewConfig()
	k, err := Kafka.New(cfg.Topic, cfg.KafkaUrl, cfg.Group)
	if err != nil {
		log.Errorf("kafka client not connected ", err)
	}
	defer k.Close()

	dispatch := dispatcher.New(k)

	_, err = Websocket.NewServer(func(msg string) {
		dispatch.Operation <- dispatcher.ReadRoleOperation{Role: msg}
	}, done, e)
	if err != nil {
		log.Errorf("websocketUtils not dial", err)
	}

	go func(e *echo.Echo) {
		e.Start(fmt.Sprintf(":%d", cfg.ServerPort))
	}(e)

	dispatch.Do(done)

	dispatch.Operation <- dispatcher.WriteMessageOperation{}

	<-s
	close(s)
	cnsl()
}
