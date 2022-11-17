package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
)

const (
	roundRobinPolicy = "ROUND_ROBIN"
	randomPolicy     = "RANDOM"
)

func HandleRequest(lb *loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostHeader := r.Host
		if _, ok := lb.GetServices()[hostHeader]; !ok {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("The service domain %s is not present in the current configuration", hostHeader))
			return
		}
		switch lb.GetLbPolicy() {
		case roundRobinPolicy:
			{

				respondWithJSON(w, http.StatusOK, "ROUND_ROBIN")
				return
			}
		case randomPolicy:
			{
				respondWithJSON(w, http.StatusOK, "RANDOM")
				return
			}
		}

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
