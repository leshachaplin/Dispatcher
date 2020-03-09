package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/leshachaplin/Dispatcher/communicationUtils/Kafka"
	"github.com/leshachaplin/Dispatcher/communicationUtils/dispatcher"
	"github.com/leshachaplin/Dispatcher/communicationUtils/operations"
	"github.com/leshachaplin/Dispatcher/communicationUtils/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

var (
	isWatcher  bool
	serverPort int
	clientPort int
	mu         sync.Mutex
)

type Message struct {
	Time string `json:"time"`
}

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(isWatcher)
		s := make(chan os.Signal)
		signal.Notify(s, os.Interrupt)
		done, cnsl := context.WithCancel(context.Background())
		e := echo.New()

		d := &dispatcher.Dispatcher{
			Operation: make(chan dispatcher.Operation),
		}

		w, err := websocket.New(OnMessage(d), done, serverPort, e)
		if err != nil {
			log.Errorf("websocket not dial", err)
		}

		go func(e *echo.Echo) {
			e.Start(fmt.Sprintf(":%d", serverPort))
		}(e)

		time.Sleep(time.Second * 10)
		ws, err := w.NewWebsocketConnection()
		if err != nil {
			log.Errorf("websocket not dial", err)
		}
		defer ws.Close()

		k, err := Kafka.New("time115", 9092, strconv.FormatBool(isWatcher))
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
								//d.CancelWrite <- CancelWrite{}
								//fmt.Println("SEND CANCEL WRITE")
							}
						} else {
							log.Info("send write operation")
							d.Operation <- operations.WriteOperationMessage{}

							if d.CancelRead != nil {
								d.CancelRead <- operations.CancelRead{}
								log.Info("send cancel read")
							} else {
								d.CancelRead = make(chan dispatcher.Operation)
								//d.CancelRead <- CancelRead{}
								//fmt.Println("SEND CANCEL READ")
							}
						}
					}()
				}
			case operations.WriteRole:
				{
					log.Info("WRITE ROLE")
					go func() {
						log.Info(fmt.Sprintf("i'am watcher %v change role %d", isWatcher, clientPort))
						mes := "read"
						if isWatcher {
							mes = "write"
							isWatcher = false
						} else {
							isWatcher = true
						}
						if _, err := ws.Write([]byte(mes)); err != nil {
							log.Errorf("message not send", err)
						}

						if !isWatcher { //????????????????????????????????????????????????????????????????
							if d.CancelRead != nil {
								d.CancelRead <- operations.CancelRead{}
								log.Info("send cancel READ")
							} else {
								d.CancelRead = make(chan dispatcher.Operation)
								d.CancelRead <- operations.CancelRead{}
								log.Info("CANCEL READ")
							}

							d.Operation <- operations.WriteOperationMessage{}
							log.Info("send WRITE operation")
						} else {
							if d.CancelWrite != nil {
								d.CancelWrite <- operations.CancelWrite{}
								log.Info("send cancel WRITE")
							} else {
								d.CancelWrite = make(chan dispatcher.Operation)
								d.CancelWrite <- operations.CancelWrite{}
								log.Info("CANCEL WRITE")
							}

							d.Operation <- operations.ReadOperationMessage{}
							log.Info("send READ operation")

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
									}
									err = k.WriteMessage(msg)
									if err != nil {
										log.Errorf("message not Send", err)
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

		ManagerOfMessages(done, d)

		if isWatcher {
			d.Operation <- operations.ReadOperationMessage{}
		} else {
			d.Operation <- operations.WriteOperationMessage{}
		}

		<-s
		close(s)
		cnsl()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&isWatcher, "watcher", false, "User role for app")
	rootCmd.PersistentFlags().IntVar(&serverPort, "serverPort", 6774, "User role for app")
	rootCmd.PersistentFlags().IntVar(&clientPort, "clientPort", 8668, "User role for app")
}

func Execute() error {
	return rootCmd.Execute()
}

func ManagerOfMessages(ctx context.Context, d *dispatcher.Dispatcher) {
	go func(ctx context.Context) {
		roleTicker := time.NewTicker(time.Second * 30)
		//messageTicker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-roleTicker.C:
				{
					log.Info("tick to change role")
					d.Operation <- operations.WriteRole{}
				}
			case <-ctx.Done():
				{
					return
				}
			}
		}
	}(ctx)
}

func OnMessage(dis *dispatcher.Dispatcher) func(msg string) {
	return func(msg string) {
		dis.Operation <- &operations.ReadRole{Role: msg}
	}
}
