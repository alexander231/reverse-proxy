package main

import (
	"log"

	"github.com/alexander231/reverse-proxy/pkg/parsing"
	"github.com/pkg/errors"
)

const filepath = "config/config.yaml"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := parsing.GetConfig(filepath)
	if err != nil {
		return errors.Wrap(err, "Getting config")
	}
	log.Println(cfg)
	return nil
}
