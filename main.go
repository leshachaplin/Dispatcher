package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/leshachaplin/Dispatcher/config"
	"github.com/leshachaplin/Dispatcher/dispatcher"
	"github.com/leshachaplin/Dispatcher/operations"
	Kafka "github.com/leshachaplin/communicationUtils/kafka"
	Websocket "github.com/leshachaplin/communicationUtils/websocket"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"time"
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

	d := &dispatcher.Dispatcher{
		Operation: make(chan dispatcher.Operation),
	}

	_, err := Websocket.NewServer(func(msg string) {
		d.Operation <- operations.ReadRole{Role: msg}
	}, done, e)
	if err != nil {
		log.Errorf("websocketUtils not dial", err)
	}

	go func(e *echo.Echo) {
		e.Start(fmt.Sprintf(":%d", cfg.ServerPort))
	}(e)

	k, err := Kafka.New("time1152", cfg.KafkaUrl, strconv.FormatBool(isWatcher))
	if err != nil {
		log.Errorf("kafka client not connected ", err)
	}
	defer k.Close()

	d.OnOperation = func(operation dispatcher.Operation) {
		switch operation.(type) {
		case operations.ReadRole:
			{
				log.Info("READ ROLE")
				go func() {
					log.Info("read role work")
					role := operation.(operations.ReadRole)
					if role.Role == "read" {
						log.Info("send read operation")
						d.Operation <- operations.ReadOperationMessage{}
						if d.CancelWrite != nil {
							d.CancelWrite <- operations.CancelWrite{}
							log.Info("send cancel write")
						} else {
							d.CancelWrite = make(chan dispatcher.Operation)
							d.CancelWrite <- operations.CancelWrite{}
						}
					} else {
						log.Info("send write operation")
						d.Operation <- operations.WriteOperationMessage{}

						if d.CancelRead != nil {
							d.CancelRead <- operations.CancelRead{}
							log.Info("send cancel read")
						} else {
							d.CancelRead = make(chan dispatcher.Operation)
							d.CancelRead <- operations.CancelRead{}
						}
					}
				}()
			}
		case operations.ReadOperationMessage:
			{
				log.Info("read Message")
				go func() {
					for {
						select {
						case <-d.CancelRead:
							{
								log.Info("cancel read")
								return
							}
						default:
							{
								m, err := k.ReadMessage()
								if err != nil {
									log.Errorf("message not read", err)
									continue
								}
								fmt.Println(string(m))

								time.Sleep(time.Second)
							}
						}
					}
				}()
			}
		case operations.WriteOperationMessage:
			{
				log.Info("write Message")
				go func() {
					for {
						select {
						case <-d.CancelWrite:
							{
								log.Info("Cancel write")
								return
							}
						default:
							{
								msg, err := json.Marshal(map[string]string{
									"time": time.Now().String(),
								})
								if err != nil {
									log.Errorf("message not Send", err)
									continue
								}
								err = k.WriteMessage(msg)
								if err != nil {
									log.Errorf("message not Send", err)
									continue
								}
								fmt.Println("SEND MESSAGE")
								time.Sleep(time.Second)
							}
						}
					}
				}()
			}
		}
	}

	dispatcher.Do(d, done)

	if isWatcher {
		d.Operation <- operations.ReadOperationMessage{}
	} else {
		d.Operation <- operations.WriteOperationMessage{}
	}

	<-s
	close(s)
	cnsl()
}

func OnMessage(dis *dispatcher.Dispatcher) func(msg string) {
	return func(msg string) {
		dis.Operation <- operations.ReadRole{Role: msg}
	}
}
