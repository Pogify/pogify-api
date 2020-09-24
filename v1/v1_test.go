package v1

import (
	"fmt"
	"io/ioutil"
	"log"
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

var _pubsubsecret = "secret"

func TestMain(m *testing.M) {
	log.Print("main")
	mockPubSubHandler := gin.New()

	mockPubSubHandler.POST("/pub", func(c *gin.Context) {
		if c.GetHeader("authorization") != _pubsubsecret {
			c.Status(401)
			return
		}

		body, _ := ioutil.ReadAll(c.Request.Body)
		log.Printf("PubSub got: %x", body)
		c.String(200, "ok")
	})

	mockPubSubServer := httptest.NewServer(mockPubSubHandler)
	defer mockPubSubServer.Close()

	os.Setenv("PUBSUB_URL", mockPubSubServer.URL)
	code := m.Run()
	os.Exit(code)
}

func TestServerV1(t *testing.T) {
	os.Setenv("REFRESH_TOKEN_TTL", "100")

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
		{"/auth/twitch", "GET"},
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
