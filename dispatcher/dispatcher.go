package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	Kafka "github.com/leshachaplin/communicationUtils/kafka"
	log "github.com/sirupsen/logrus"
	"time"
)

type Dispatcher struct {
	Operation   chan Operation
	CancelRead  chan Operation
	CancelWrite chan Operation
	kafkaClient *Kafka.Client
}

func New(k *Kafka.Client) *Dispatcher {
	return &Dispatcher{
		Operation:   make(chan Operation),
		kafkaClient: k,
	}
}

func (d *Dispatcher) Do(done context.Context) {
	go func(d *Dispatcher) {
		for {
			select {
			case m := <-d.Operation:
				switch m.(type) {
				case ReadRoleOperation:
					{
						d.ReadRoleContainer(m)
					}
				case ReadMessageOperation:
					{
						d.ReadMessageContainer()
					}
				case WriteMessageOperation:
					{
						d.WriteMessageContainer()
					}
				}
			case <-done.Done():
				return
			}
		}
	}(d)
}

func (d *Dispatcher) WriteMessageContainer() {
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
					err = d.kafkaClient.WriteMessage(msg)
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

func (d *Dispatcher) ReadMessageContainer() {
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
					m, err := d.kafkaClient.ReadMessage()
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

func (d *Dispatcher) ReadRoleContainer(operation Operation) {
	go func() {
		log.Info("READ ROLE")

		role := operation.(ReadRoleOperation)
		if role.Role == "read" {
			log.Info("send read operation")
			d.Operation <- ReadMessageOperation{}
			if d.CancelWrite != nil {
				d.CancelWrite <- CancelWrite{}
				log.Info("send cancel write")
			} else {
				d.CancelWrite = make(chan Operation)
				d.CancelWrite <- CancelWrite{}
			}
		} else {
			log.Info("send write operation")
			d.Operation <- WriteMessageOperation{}

			if d.CancelRead != nil {
				d.CancelRead <- CancelRead{}
				log.Info("send cancel read")
			} else {
				d.CancelRead = make(chan Operation)
				d.CancelRead <- CancelRead{}
			}
		}
	}()
}
