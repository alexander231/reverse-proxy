package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/alexander231/reverse-proxy/pkg/loadbalancer"
)

type lbMockNoServices struct{}

func (lbm *lbMockNoServices) CountServices() int {
	return 0
}
func (lbm *lbMockNoServices) GetServices() map[string]loadbalancer.Service {
	return map[string]loadbalancer.Service{}
}
func (lbm *lbMockNoServices) GetLbPolicy() string {
	return "ROUND_ROBIN"
}

func (lbm *lbMockNoServices) NextPeer(sp *loadbalancer.ServerPool) (*loadbalancer.Server, error) {
	return nil, nil
}

type lbMockNoPeer struct{}

func (lb *lbMockNoPeer) CountServices() int {
	return 1
}
func (lb *lbMockNoPeer) GetServices() map[string]loadbalancer.Service {
	return map[string]loadbalancer.Service{"service1.com": {}}
}
func (lb *lbMockNoPeer) GetLbPolicy() string {
	return "RANDOM"
}
func (lb *lbMockNoPeer) NextPeer(sp *loadbalancer.ServerPool) (*loadbalancer.Server, error) {
	return nil, nil
}

type lbMockPeerOff struct{}

func (lb *lbMockPeerOff) CountServices() int {
	return 1
}
func (lb *lbMockPeerOff) GetServices() map[string]loadbalancer.Service {
	return map[string]loadbalancer.Service{"service1.com": {}}
}
func (lb *lbMockPeerOff) GetLbPolicy() string {
	return "RANDOM"
}
func (lb *lbMockPeerOff) NextPeer(sp *loadbalancer.ServerPool) (*loadbalancer.Server, error) {
	mockURL, _ := url.Parse("http://mock")
	return &loadbalancer.Server{URL: mockURL, ReverseProxy: httputil.NewSingleHostReverseProxy(mockURL)}, nil
}
func TestHandleRequest(t *testing.T) {
	t.Run("returns response for request with empty string at Host header", func(t *testing.T) {
		var resMsg map[string]string
		emptyHostHeader := ""
		req := newHandlerRequest(emptyHostHeader)
		res := httptest.NewRecorder()
		lb := &lbMockNoServices{}

		HandleRequest(lb).ServeHTTP(res, req)
		_ = json.Unmarshal(res.Body.Bytes(), &resMsg)

		assertStatus(t, res.Code, http.StatusBadRequest)
		assertResponseBody(t, resMsg["error"], fmt.Sprintf("Please provice a service domain in the Host header, current Host header: %s", emptyHostHeader))
	})
	t.Run("returns response for request valid Host Header no Peer", func(t *testing.T) {
		var resMsg map[string]string
		hostHeader := "service1.com"
		req := newHandlerRequest(hostHeader)
		res := httptest.NewRecorder()
		lb := &lbMockNoPeer{}

		HandleRequest(lb).ServeHTTP(res, req)
		_ = json.Unmarshal(res.Body.Bytes(), &resMsg)

		assertStatus(t, res.Code, http.StatusServiceUnavailable)
		assertResponseBody(t, resMsg["error"], "No server available for the service domain service1.com")
	})
	t.Run("returns response for request valid Host Header Has Peer Offline", func(t *testing.T) {
		hostHeader := "service1.com"
		req := newHandlerRequest(hostHeader)
		res := httptest.NewRecorder()
		lb := &lbMockPeerOff{}

		HandleRequest(lb).ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusBadGateway)
		assertResponseBody(t, res.Body.String(), "")
	})
}

func newHandlerRequest(hostHeader string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Host = hostHeader
	return req
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}
