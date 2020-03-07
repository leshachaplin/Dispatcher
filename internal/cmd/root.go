package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/leshachaplin/Dispatcher/Websocket"
	"github.com/leshachaplin/Dispatcher/internal/dispatcher"
	"github.com/leshachaplin/Dispatcher/internal/initialization"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	isWatcher  bool
	serverPort int
	clientPort int
	mu         sync.Mutex
)

type ReadRole struct {
	Role string
}

type WriteRole struct {
}

type ReadOperationMessage struct {
}

type WriteOperationMessage struct {
}

type CancelWrite struct {
}

type CancelRead struct {
}

type Message struct {
	Time string `json:"time"`
}

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(isWatcher)
		s := make(chan os.Signal)
		signal.Notify(s, os.Interrupt)
		done, cnsl := context.WithCancel(context.Background())
		d := &dispatcher.Dispatcher{
			Operation: make(chan dispatcher.Operation),
		}
		//
		//e := echo.New()
		//
		//e.GET("/role", func(c echo.Context) error {
		//	websocket.Handler(func(ws *websocket.Conn) {
		//		fmt.Println("handle server connection")
		//		defer ws.Close()
		//		//fmt.Println(fmt.Sprintf("i'am RoleWorker port%d", port))
		//		for {
		//
		//			fmt.Println("Try to read from websocket")
		//			msg := ""
		//			err := websocket.Message.Receive(ws, &msg)
		//			fmt.Println(fmt.Sprintf("Read from websoket %s", msg))
		//			if err != nil {
		//				fmt.Println(fmt.Sprintf("Error %s", err))
		//			}
		//
		//			//fmt.Println(fmt.Sprintf("i'am port %d read message %s", port, msg))
		//			d.Operation <- &ReadRole{Role: msg}
		//			//if port == message {
		//			fmt.Println(fmt.Sprintf("message %s", msg))
		//		}
		//		//}
		//	}).ServeHTTP(c.Response(), c.Request())
		//
		//	return nil
		//})

		w := Websocket.New(OB(*d), done, serverPort)

		go func(e *echo.Echo) {
			e.Start(fmt.Sprintf(":%d", serverPort))
		}(w.Echo)

		time.Sleep(time.Second * 10)
		ws, err := initialization.NewWebsocket(clientPort)
		if err != nil {
			log.Fatalf("websocket not connect", err)
		}

		defer ws.Close()

		readerMsg := initialization.NewKafkaReader(9092, isWatcher)
		defer readerMsg.Close()

		senderMsg, err := kafka.DialLeader(context.Background(), "tcp", fmt.Sprintf("localhost:%d", 9092), "time6", 0)
		if err != nil {
			log.Fatalf("not dialling", err)
		}
		defer senderMsg.Close()

		d.OnOperation = func(operation dispatcher.Operation) {
			switch operation.(type) {
			case ReadRole:
				{
					go func() {
						fmt.Println("read role work")
						role := operation.(ReadRole)
						if role.Role == "read" {
							fmt.Println("send read operation")
							d.Operation <- ReadOperationMessage{}
							if d.CancelWrite != nil {
								d.CancelWrite <- CancelWrite{}
								fmt.Println("send cancel write")
							} else {
								d.CancelWrite = make(chan dispatcher.Operation)
							}
						} else {
							fmt.Println("send write operation")
							d.Operation <- WriteOperationMessage{}

							if d.CancelRead != nil {
								d.CancelRead <- CancelRead{}
								fmt.Println("send cancel write")
							} else {
								d.CancelRead = make(chan dispatcher.Operation)
							}
						}
					}()
				}
			case WriteRole:
				{
					go func() {
						fmt.Println(fmt.Sprintf("i'am watcher %v change role %d", isWatcher, clientPort))
						mes := "write"
						if isWatcher {
							mes = "read"
							isWatcher = false
						} else {
							isWatcher = true
						}
						_, err := ws.Write([]byte(mes))

						if err != nil {
							log.Errorf("message not read", err)
						}

						if isWatcher {
							if d.CancelRead != nil {
								d.CancelRead <- CancelRead{}
								fmt.Println("send cancel write")
							} else {
								d.CancelRead = make(chan dispatcher.Operation)
							}

							d.Operation <- WriteOperationMessage{}
							fmt.Println("send write operation")
						} else {
							if d.CancelWrite != nil {
								d.CancelWrite <- CancelWrite{}
								fmt.Println("send cancel write")
							} else {
								d.CancelWrite = make(chan dispatcher.Operation)
							}

							d.Operation <- ReadOperationMessage{}
							fmt.Println("send read operation")

						}
					}()

				}
			case ReadOperationMessage:
				{
					fmt.Println("read Message")
					go ReadMessageFromKafka(done, d, readerMsg)
				}
			case WriteOperationMessage:
				{
					fmt.Println("write Message")
					go WriteMwssageToKafka(d, senderMsg)
				}
			}
		}

		dispatcher.Do(d, done)
		if isWatcher {
			d.Operation <- ReadOperationMessage{}
		} else {
			d.Operation <- WriteOperationMessage{}
		}
		ManagerOfMessages(done, d)
		//SendOnAnother(done, dispatcher)

		<-s
		close(s)
		cnsl()
	},
}

//func WriteMwssageToKafka(d *dispatcher.Dispatcher, kClient *kafka.Conn) {
//	fmt.Println("write work")
//	for {
//		select {
//		case <-d.CancelWrite:
//			{
//				fmt.Println("Cancel write")
//				return
//			}
//		default:
//			{
//
//				msg, err := json.Marshal(map[string]string{
//					"time": time.Now().String(),
//				})
//				if err != nil {
//					log.Errorf("message not Send", err)
//				}
//				_, err = kClient.WriteMessages(kafka.Message{
//					Value: msg,
//				})
//				if err != nil {
//					log.Errorf("message not Send", err)
//				}
//				fmt.Println("SEND MESSAGE")
//				time.Sleep(time.Second)
//			}
//		}
//	}
//}
//
//func ReadMessageFromKafka(done context.Context, d *dispatcher.Dispatcher, kClient *kafka.Reader) {
//	fmt.Println("read work")
//	for {
//		select {
//		case <-d.CancelRead:
//			{
//				fmt.Println("cancel read")
//				return
//			}
//		default:
//			{
//				ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
//				m, err := kClient.ReadMessage(ctx)
//				if err != nil {
//					log.Errorf("message not reead", err)
//				}
//				fmt.Println(string(m.Value))
//				time.Sleep(time.Second)
//			}
//		}
//	}
//}

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
		roleTicker := time.NewTicker(time.Second * 10)
		//messageTicker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-roleTicker.C:
				{
					fmt.Println("tick to change role")
					d.Operation <- WriteRole{}
				}
			case <-ctx.Done():
				{
					return
				}
			}
		}
	}(ctx)
}

func OB(dis dispatcher.Dispatcher) func(msg string) {
	return func(msg string) {
		dis.Operation <- &ReadRole{Role: msg}
	}
}
