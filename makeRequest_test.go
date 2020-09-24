package pogifyapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func Test_server_makeRequest(t *testing.T) {
	m, _ := miniredis.Run()
	defer m.Close()
	os.Setenv("REDIS_URI", "redis://"+m.Addr())

	router := gin.Default()

	Server(router.Group("/"))

	t.Run("empty call", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/makeRequest", nil)
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("empty call to makeRequest didn't return 400, but %v", w.Code)
		}
	})

	t.Run("invalid body, inactive session", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/makeRequest", strings.NewReader("{\"invalid\":\"invalid\"}"))
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("invalid body to makeRequest didn't return 400, but %v", w.Code)
		}
	})
	t.Run("valid request, inactive session", func(t *testing.T) {
		validBody := request{
			Session:  "notexist",
			Provider: "notaprovider",
			Token:    "not.a.token",
			Request:  "not a request",
		}
		validBodyBytes, _ := json.Marshal(validBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/makeRequest", bytes.NewReader(validBodyBytes))
		router.ServeHTTP(w, req)

		if expect := 404; w.Code != expect {
			body, _ := ioutil.ReadAll(w.Body)
			t.Errorf("%s", string(body))
			t.Errorf("invalid body to makeRequest didn't return %v, but %v", expect, w.Code)
		}
	})
	t.Run("valid request, active session", func(t *testing.T) {
		validBody := request{
			Session:  "exist",
			Provider: "notaprovider",
			Token:    "not.a.token",
			Request:  "not a request",
		}
		validBodyBytes, _ := json.Marshal(validBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/makeRequest", bytes.NewReader(validBodyBytes))
		router.ServeHTTP(w, req)

		if expect := 200; w.Code != expect {
			body, _ := ioutil.ReadAll(w.Body)
			t.Errorf("%s", string(body))
			t.Errorf("invalid body to makeRequest didn't return %v, but %v", expect, w.Code)
		}
	})
	t.Run("repeated call should 429", func(t *testing.T) {
		validBody := request{
			Session:  "exist",
			Provider: "notaprovider",
			Token:    "not.a.token",
			Request:  "not a request",
		}
		validBodyBytes, _ := json.Marshal(validBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/makeRequest", bytes.NewReader(validBodyBytes))
		router.ServeHTTP(w, req)

		if expect := 429; w.Code != expect {
			body, _ := ioutil.ReadAll(w.Body)
			t.Errorf("%s", string(body))
			t.Errorf("invalid body to makeRequest didn't return %v, but %v", expect, w.Code)
		}
	})
}
