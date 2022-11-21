package loadbalancer

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexander231/reverse-proxy/pkg/parsing"
	"github.com/pkg/errors"
)

const (
	roundRobinPolicy = "ROUND_ROBIN"
	randomPolicy     = "RANDOM"
)

type LoadBalancer interface {
	CountServices() int
	GetServices() map[string]Service
	GetLbPolicy() string
	NextPeer(*ServerPool) (*Server, error)
}

type loadBalancer struct {
	services map[string]Service
	lbPolicy string
}
type Service struct {
	Name       string
	Domain     string
	ServerPool *ServerPool
}

type ServerPool struct {
	Servers []*Server
	Current uint64
}
type Server struct {
	URL            *url.URL
	tcpHealthcheck bool
	mu             sync.RWMutex
	ReverseProxy   *httputil.ReverseProxy
}

func NewLoadBalancer(cfg parsing.Config) *loadBalancer {
	return &loadBalancer{lbPolicy: cfg.GetLbPolicy(), services: buildServices(cfg.GetServices())}
}

func (lb *loadBalancer) CountServices() int {
	return len(lb.services)
}

func (lb *loadBalancer) GetServices() map[string]Service {
	return lb.services
}

func (lb *loadBalancer) GetLbPolicy() string {
	return lb.lbPolicy
}

func (lb *loadBalancer) NextPeer(sp *ServerPool) (*Server, error) {
	switch lb.lbPolicy {
	case roundRobinPolicy:
		{
			return sp.GetNextPeerRR(), nil
		}
	case randomPolicy:
		{
			return sp.GetNextPeerRandom(), nil
		}
	}
	return nil, errors.New("Not a valid loadbalacing policy in config")
}

func (s *Server) SetAlive(alive bool) {
	s.mu.Lock()
	s.tcpHealthcheck = alive
	s.mu.Unlock()
}

func (s *Server) IsAlive() bool {
	s.mu.RLock()
	alive := s.tcpHealthcheck
	s.mu.RUnlock()

	return alive
}

func (s *Server) GetURL() *url.URL {
	return s.URL
}

func (sp *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&sp.Current, uint64(1)) % uint64(len(sp.Servers)))
}

// GetNextPeer returns next active peer to take a connection
func (sp *ServerPool) GetNextPeerRR() *Server {
	// loop the servers to find a server that is alive
	next := sp.NextIndex()
	l := len(sp.Servers) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(sp.Servers) // take an index by modding with length
		// if we have an alive server, use it and store if it is not the original one
		if sp.Servers[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&sp.Current, uint64(idx)) // mark the current one
			}
			return sp.Servers[idx]
		}
	}
	return nil
}

func (sp *ServerPool) GetNextPeerRandom() *Server {
	// generate random source
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	l := len(sp.Servers)

	// generate random index
	randomIdx := random.Intn(l)

	// get the current index
	curr := int(atomic.LoadUint64(&sp.Current))
	// if we have and alive server, use it and store if it is not the current one
	if sp.Servers[randomIdx].IsAlive() {
		if randomIdx != curr {
			atomic.StoreUint64(&sp.Current, uint64(randomIdx)) // mark the current one
		}
		return sp.Servers[randomIdx]
	}
	return nil
}

func (sp *ServerPool) healthCheck() {
	for _, srv := range sp.Servers {
		status := "up"
		alive := isServerAlive(srv.URL)
		srv.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", srv.URL, status)
	}
}

func (svc *Service) CountServers() int {
	return len(svc.ServerPool.Servers)
}

func (svc *Service) GetServerPool() *ServerPool {
	return svc.ServerPool
}

func HealthCheck(services map[string]Service) {
	t := time.NewTicker(time.Second * 20)
	for {
		for _, svc := range services {
			select {
			case <-t.C:
				log.Printf("Starting health check on domain %s...\n", svc.Domain)
				svc.ServerPool.healthCheck()
				log.Printf("Health check completed on domain %s\n", svc.Domain)
			}
		}
	}
}

func buildServices(services []parsing.Service) map[string]Service {
	lbServices := make(map[string]Service)
	for _, svc := range services {
		servers := make([]*Server, 0)
		for _, host := range svc.Hosts {
			// get the URL for each server
			serverURL, _ := url.Parse(fmt.Sprintf("%s:%d", host.Address, host.Port))
			// set if that server is alive
			alive := isServerAlive(serverURL)
			// set proxy for each server
			proxy := httputil.NewSingleHostReverseProxy(serverURL)
			servers = append(servers, &Server{URL: serverURL, tcpHealthcheck: alive, ReverseProxy: proxy})
		}
		srvPool := &ServerPool{Servers: servers, Current: 0}
		lbServices[svc.Domain] = Service{Name: svc.Name, Domain: svc.Domain, ServerPool: srvPool}
	}
	return lbServices
}

func isServerAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Service unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}
