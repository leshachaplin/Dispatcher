package main

import (
	"github.com/leshachaplin/Dispatcher/internal/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
