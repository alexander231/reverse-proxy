package main

import (
	"log"

	"github.com/pkg/errors"

	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
	"github.com/alexander231/reverse-proxy/pkg/parsing"
	"github.com/alexander231/reverse-proxy/pkg/server"
)

// edit this after finished to run project from root dir
// const filepath = "config/config.yaml"
const (
	filepath = "config/config.yaml"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := parsing.NewConfig(filepath)
	if err != nil {
		return errors.Wrap(err, "Getting config")
	}

	lb := loadbalancer.NewLoadBalancer(cfg.GetLbPolicy(), cfg.GetServices())
	if lb.CountServices() == 0 {
		return errors.Wrap(err, "Please provide one or more services")
	}

	go loadbalancer.HealthCheck(lb)

	port := cfg.GetProxyPort()
	address := cfg.GetProxyAddress()
	server.Start(lb, address, port)

	return nil
}
