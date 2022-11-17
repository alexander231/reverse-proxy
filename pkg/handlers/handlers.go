package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
)

func HandleRequest(lb *loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(lb.GetLbPolicy())
		switch lb.GetLbPolicy() {
		case "ROUND_ROBIN":
			{
				respondWithJSON(w, http.StatusOK, "ROUND_ROBIN")
				return
			}
		case "RANDOM":
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
