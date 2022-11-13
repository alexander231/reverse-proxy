package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexander231/reverse-proxy/pkg/handlers"
	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
	"github.com/alexander231/reverse-proxy/pkg/parsing"
	"github.com/pkg/errors"
)

func Start(cfg *parsing.Config) error {
	log.Println(cfg)
	lb := loadbalancer.NewLoadBalancer(cfg.GetServices())
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleRequest(lb))

	PORT := cfg.GetProxyPort()
	ADDRESS := cfg.GetProxyAddress()
	log.Printf("Serving requests at %s:%d", ADDRESS, PORT)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), mux); err != nil {
		return errors.Wrap(err, "Listening and serving")
	}
	return nil
}
