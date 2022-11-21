package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
)

func HandleRequest(lb loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostHeader := r.Host
		lbServices := lb.GetServices()
		svc, ok := lbServices[hostHeader]
		if !ok {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Please provice a service domain in the Host header, current Host header: %s", hostHeader))
			return
		}
		sp := svc.GetServerPool()
		peer, err := lb.NextPeer(sp)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if peer != nil {
			log.Println(peer.URL)
			peer.ReverseProxy.ServeHTTP(w, r)
			return
		}
		respondWithError(w, http.StatusServiceUnavailable, fmt.Sprintf("No server available for the service domain %s", hostHeader))
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
