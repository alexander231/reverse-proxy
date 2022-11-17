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
)

type LoadBalancer struct {
	services map[string]service
	lbPolicy string
}

type service struct {
	name       string
	domain     string
	serverPool *serverPool
}

type serverPool struct {
	servers []*server
	current uint64
}
type server struct {
	URL            *url.URL
	tcpHealthcheck bool
	mu             sync.RWMutex
	ReverseProxy   *httputil.ReverseProxy
}

func NewLoadBalancer(lbPolicy string, services []parsing.Service) *LoadBalancer {
	return &LoadBalancer{lbPolicy: lbPolicy, services: buildServices(services)}
}

func (lb *LoadBalancer) CountServices() int {
	return len(lb.GetServices())
}

func (lb *LoadBalancer) GetServices() map[string]service {
	return lb.services
}

func (lb *LoadBalancer) GetLbPolicy() string {
	return lb.lbPolicy
}

func (s *server) SetAlive(alive bool) {
	s.mu.Lock()
	s.tcpHealthcheck = alive
	s.mu.Unlock()
}

func (s *server) IsAlive() bool {
	s.mu.RLock()
	alive := s.tcpHealthcheck
	s.mu.RUnlock()

	return alive
}

func (s *server) GetURL() *url.URL {
	return s.URL
}

func (sp *serverPool) NextIndex() int {
	return int(atomic.AddUint64(&sp.current, uint64(1)) % uint64(len(sp.servers)))
}

// GetNextPeer returns next active peer to take a connection
func (sp *serverPool) GetNextPeerRR() *server {
	// loop the servers to find a server that is alive
	next := sp.NextIndex()
	l := len(sp.servers) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(sp.servers) // take an index by modding with length
		// if we have an alive server, use it and store if it is not the original one
		if sp.servers[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&sp.current, uint64(idx)) // mark the current one
			}
			return sp.servers[idx]
		}
	}
	return nil
}

func (sp *serverPool) GetNextPeerRandom() *server {
	// generate random source
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	l := len(sp.servers)

	// generate random index
	randomIdx := random.Intn(l)

	// get the current index
	curr := int(atomic.LoadUint64(&sp.current))
	log.Printf("curr: %d randomIndex: %d", curr, randomIdx)
	// if we have and alive server, use it and store if it is not the current one
	if sp.servers[randomIdx].IsAlive() {
		if randomIdx != curr {
			atomic.StoreUint64(&sp.current, uint64(randomIdx)) // mark the current one
		}
		return sp.servers[randomIdx]
	}
	return nil
}

func (sp *serverPool) healthCheck() {
	for _, srv := range sp.servers {
		status := "up"
		alive := isServerAlive(srv.URL)
		srv.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", srv.URL, status)
	}
}

func (svc *service) CountServers() int {
	return len(svc.serverPool.servers)
}

func HealthCheck(lb *LoadBalancer) {
	t := time.NewTicker(time.Second * 20)
	for {
		for _, svc := range lb.GetServices() {
			select {
			case <-t.C:
				log.Printf("Starting health check on domain %s...\n", svc.domain)
				svc.serverPool.healthCheck()
				log.Printf("Health check completed on domain %s\n", svc.domain)
			}
		}
	}
}

func buildServices(services []parsing.Service) map[string]service {
	lbServices := make(map[string]service)
	for _, svc := range services {
		servers := make([]*server, 0)
		for _, host := range svc.Hosts {
			serverURL, _ := url.Parse(fmt.Sprintf("%s:%d", host.Address, host.Port))
			alive := isServerAlive(serverURL)
			servers = append(servers, &server{URL: serverURL, tcpHealthcheck: alive})
		}
		srvPool := &serverPool{servers: servers}
		lbServices[svc.Domain] = service{name: svc.Name, domain: svc.Domain, serverPool: srvPool}
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
