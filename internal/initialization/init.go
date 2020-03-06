package initialization

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"golang.org/x/net/websocket"
)

func NewKafkaReader(port int, group bool) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{fmt.Sprintf("localhost:%d", port)},
		Topic:     "time6",
		GroupID:   fmt.Sprintf("%v", group),
		Partition: 0,
	})
}

func NewWebsocket(port int) (*websocket.Conn, error) {
	origin := fmt.Sprintf("http://localhost:%d/", port)
	url := fmt.Sprintf("ws://localhost:%d/role", port)
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	return ws, nil
}
