package config

// if using go modules

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerPort int    `env:"Port" envDefault:"6774"`
	ClientPort int    `env:"Port" envDefault:"8668"`
	Group      string `env:"Group" envDefault:"Sender"`
	Topic      string `env:"Topic" envDefault:"time1000"`
	KafkaUrl   string `env:"Port" envDefault:"localhost:9092"`
	Origin     string `env:"Origin" envDefault:"http://localhost:6774/"`
	Url        string `env:"Url" envDefault:"ws://localhost:6774/role"`
}

func NewConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return &cfg
}
