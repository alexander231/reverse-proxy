package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/alexander231/reverse-proxy/pkg/handlers"
	"github.com/alexander231/reverse-proxy/pkg/parsing"
	"github.com/pkg/errors"
)

func Start(cfg *parsing.Config) error {
	log.Println(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleRequest)

	PORT := cfg.GetProxyPort()
	log.Printf("Listening on port %d", PORT)
	if err := http.ListenAndServe(":"+strconv.Itoa(PORT), mux); err != nil {
		return errors.Wrap(err, "Listening and serving")
	}
	return nil
}
