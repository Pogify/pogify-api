package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

var _ = (func() interface{} {
	_testing = true
	return nil
})()

func TestServerV1(t *testing.T) {
	mr, err := miniredis.Run()
	defer mr.Close()
	if err != nil {
		t.Fatalf("MiniRedis error: %s", err)
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	router := gin.Default()

	ServerV1(router.Group("/"))

	cases := []struct {
		endpoint string
		method   string
	}{
		{"/startSession", "POST"},
		{"/refreshSession", "POST"},
		{"/postUpdate", "OPTIONS"},
		{"/postUpdate", "POST"},
		{"/makeRequest", "POST"},
		{"/makeRequest", "OPTIONS"},
		{"/setConfig", "OPTIONS"},
		{"/getConfig", "GET"},
		{"/twitchauth", "GET"},
	}

	for _, testCase := range cases {
		t.Run(fmt.Sprintf("Endpoint exists: %s %s", testCase.method, testCase.endpoint), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(testCase.method, testCase.endpoint, nil)
			router.ServeHTTP(w, req)

			if w.Code == 404 {
				t.Errorf("%s %s returned 404, ", testCase.method, testCase.endpoint)
			}

		})

	}
}
