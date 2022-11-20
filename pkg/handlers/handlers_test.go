package handlers

import (
	"net/http"
	"testing"

	"github.com/alexander231/reverse-proxy/pkg/parsing"
)

// func TestHandleRequest(t *testing.T) {
// 	t.Run("returns response for request with empty string", func(t *testing.T) {
// 		cfg := newConfig()
// 		request := newHandlerRequest("/")
// 		response := httptest.NewRecorder()
// 		lb := loadbalancer.NewLoadBalancer("RANDOM")

// 		assertStatus(t, response.Code, http.StatusOK)
// 		assertResponseBody(t, response.Body.String(), "20")
// 	})
// }

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func newConfig(filepath string) *parsing.Config {
	cfg, _ := parsing.NewConfig(filepath)
	return cfg
}

func newLoadBalancer() {
	return
}
func newHandlerRequest(hostname string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Host = hostname
	return req
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}
