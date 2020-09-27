package pogifyapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func Test_server_startSession(t *testing.T) {
	mr, err := miniredis.Run()
	defer mr.Close()
	if err != nil {
		t.Fatalf("MiniRedis error: %s", err)
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	router := gin.New()

	Server(router.Group("/"))
	var key1 string
	t.Run("Test /session/start returns 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/start", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatal("/session/start didn't return 200 on normal operation")
		}

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
			return
		}

		var v map[string]interface{}
		json.Unmarshal(body, &v)

		key1 = v["session"].(string)

		if v["token"].(string) == "" {
			t.Error("Returned empty token")
		}

		if v["expiresIn"].(float64) != 3600 {
			t.Errorf("Returned expiresIn != 3600. Got: %v", v["expiresIn"].(int))
		}

		if v["refreshToken"].(string) == "" {
			t.Error("Returned empty token")
		}

		rt, err := mr.Get("session:" + v["session"].(string))

		if v["refreshToken"] != rt {
			t.Errorf("RefreshToken mismatch. Redis: %v, Response: %v", rt, v["refreshToken"])
		}

	})

	t.Run("Test /session/start returns 200 the other key", func(t *testing.T) {
		var otherKey string
		if key1 == "test1" {
			otherKey = "test2"
		} else {
			otherKey = "test1"
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/start", nil)
		router.ServeHTTP(w, req)
		body, _ := ioutil.ReadAll(w.Body)

		var v map[string]interface{}
		json.Unmarshal(body, &v)
		if v["session"] != otherKey {
			t.Fatalf("Incorrect key receieved. Got: %v. Expected: %v", v["session"], otherKey)
		}

	})

}
