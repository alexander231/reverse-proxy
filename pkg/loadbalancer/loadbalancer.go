package loadbalancer

import (
	"fmt"

	"github.com/alexander231/reverse-proxy/pkg/parsing"
)

type LoadBalancer struct {
	Services []parsing.Service
	LbPolicy string
}

type Service struct {
	name    string
	domain  string
	servers []Server
}

type Server struct {
	address string
}

func NewLoadBalancer(services []parsing.Service) *LoadBalancer {
	return &LoadBalancer{LbPolicy: "RANDOM"}
}

func buildServices(services []parsing.Service) []string {
	addresses := make([]string, 0)
	for _, service := range services {
		for _, host := range service.Hosts {
			addresses = append(addresses, fmt.Sprintf("%s:%d", host.Address, host.Port))
		}
	}
	return addresses
}
