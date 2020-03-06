package Kafka

import (
	"context"
	"fmt"
	"github.com/leshachaplin/Dispatcher/internal/dispatcher"
	"github.com/segmentio/kafka-go"
	"time"
)

type Kafka struct {
	connection *kafka.Conn
	reader     *kafka.Reader
}

func New(topic string, port int, group string) (*Kafka, error) {

	senderMsg, err := kafka.DialLeader(context.Background(), "tcp", fmt.Sprintf("localhost:%d", 9092), topic, 0)
	if err != nil {
		return nil, err
	}

	readerMsg := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{fmt.Sprintf("localhost:%d", port)},
		Topic:     topic,
		GroupID:   fmt.Sprintf("%v", group),
		Partition: 0,
	})
	return &Kafka{
		connection: senderMsg,
		reader:     readerMsg,
	}, nil
}

func (k *Kafka) WriteMessage(msg []byte) error {
	//lo
	_, err := k.connection.WriteMessages(kafka.Message{
		Value: msg,
	})
	return err
}

func (k *Kafka) ReadMessage(done context.Context, d *dispatcher.Dispatcher) ([]byte, error) {
	//log

	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	m, err := k.reader.ReadMessage(ctx)
	return m.Value, err
}
