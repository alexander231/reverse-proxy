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
	URL          *url.URL
	alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewLoadBalancer(lbPolicy string, services []parsing.Service) *LoadBalancer {
	return &LoadBalancer{lbPolicy: lbPolicy, services: buildServices(services)}
}

func (lb *LoadBalancer) GetServices() map[string]service {
	return lb.services
}

func (lb *LoadBalancer) GetLbPolicy() string {
	return lb.lbPolicy
}

func (s *server) SetAlive(alive bool) {
	s.mu.Lock()
	s.alive = alive
	s.mu.Unlock()
}

func (s *server) IsAlive() bool {
	s.mu.RLock()
	alive := s.alive
	s.mu.RUnlock()

	return alive
}

func (s *server) GetURL() *url.URL {
	return s.URL
}

func (s *serverPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.servers)))
}

// GetNextPeer returns next active peer to take a connection
func (s *serverPool) GetNextPeerRR() *server {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.servers) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.servers) // take an index by modding with length
		// if we have an alive backend, use it and store if its not the original one
		if s.servers[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx)) // mark the current one
			}
			return s.servers[idx]
		}
	}
	return nil
}

func (s *serverPool) GetNextPeerRandom() *server {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	l := len(s.servers)
	randomIdx := random.Intn(l)
	curr := int(atomic.LoadUint64(&s.current))
	log.Printf("curr: %d randomIndex: %d", curr, randomIdx)
	if s.servers[randomIdx].IsAlive() {
		if randomIdx != curr {
			atomic.StoreUint64(&s.current, uint64(randomIdx)) // mark the current one
		}
		return s.servers[randomIdx]
	}
	return nil
}

func buildServices(services []parsing.Service) map[string]service {
	lbServices := make(map[string]service)
	for _, svc := range services {
		servers := make([]*server, 0)
		for _, host := range svc.Hosts {
			serverURL, _ := url.Parse(fmt.Sprintf("%s:%d", host.Address, host.Port))
			alive := isServerAlive(serverURL)
			servers = append(servers, &server{URL: serverURL, alive: alive})
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

// func (lb *LoadBalancer) CountServers() int {
// 	return len(lb.ser)
// }
