package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexander231/reverse-proxy/pkg/handlers"
	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
)

func Start(lb loadbalancer.LoadBalancer, address string, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleRequest(lb))

	log.Printf("Serving requests at %s:%d", address, port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
