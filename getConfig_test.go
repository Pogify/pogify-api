package pogifyapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func Test_server_getConfig(t *testing.T) {
	mr, err := miniredis.Run()
	defer mr.Close()
	if err != nil {
		t.Fatalf("MiniRedis error: %s", err)
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	router := gin.New()

	Server(router.Group("/"))

	t.Run("test get on missing query", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/session/config", nil)
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("getConfig didn't return 400 on empty query, instead: %#v", w.Code)
		}
	})

	t.Run("test get on missing config", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/session/config?session=test", nil)
		router.ServeHTTP(w, req)

		if w.Code != 404 {
			t.Errorf("getConfig didn't return 404 on empty query, instead: %#v", w.Code)
		}
	})

	t.Run("test get on missing config", func(t *testing.T) {
		mr.HSet("session:test:config", "RefreshInterval", "100")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/session/config?session=test", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("getConfig didn't return 404 on empty query, instead: %#v", w.Code)
		}
	})

}
