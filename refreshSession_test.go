package pogifyapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

func Test_server_refreshSession(t *testing.T) {
	sessionCode, _ := gonanoid.ID(10)
	claims := sessionJwtClaims{
		sessionCode,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	mockToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		t.Fatalf("Error generating token: %v", err)
	}

	mr, err := miniredis.Run()
	defer mr.Close()
	if err != nil {
		t.Fatalf("MiniRedis error: %s", err)
		return
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	router := gin.New()

	Server(router.Group("/"))

	t.Run("empty call", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/refresh", nil)
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("empty call to refreshSession didn't return 400, instead: %v", w.Code)
		}
	})
	t.Run("with sessionToken without refreshToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/refresh", nil)
		q := req.URL.Query()
		q.Add("sessionToken", mockToken)

		req.URL.RawQuery = q.Encode()
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("call without refreshToken to refreshSession didn't return 400, instead: %v", w.Code)
		}
	})
	t.Run("with invalid sessionToken, with refreshToken", func(t *testing.T) {
		invalidToken, _ := token.SignedString([]byte("aaaa"))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/refresh", nil)
		q := req.URL.Query()
		q.Add("sessionToken", invalidToken)
		q.Add("refreshToken", "aaaaaaa")

		req.URL.RawQuery = q.Encode()
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("call with invalid session to refreshSession didn't return 400, instead: %v", w.Code)
		}
	})
	t.Run("with sessionToken, with expired refreshToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/refresh", nil)
		q := req.URL.Query()
		q.Add("sessionToken", mockToken)
		q.Add("refreshToken", "aaaaa")

		req.URL.RawQuery = q.Encode()
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("call with expired refreshSession didn't return 400, instead: %v", w.Code)
		}
	})
	t.Run("with sessionToken, with invalid refreshToken", func(t *testing.T) {
		mr.Set("session:"+sessionCode, "abc")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/refresh", nil)
		q := req.URL.Query()
		q.Add("sessionToken", mockToken)
		q.Add("refreshToken", "aaaaa")

		req.URL.RawQuery = q.Encode()
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("call with invalid refreshToken to refreshSession didn't return 400, instead: %v", w.Code)
		}
	})
	t.Run("with sessionToken, with good refreshToken", func(t *testing.T) {
		mr.Set("session:"+sessionCode, "abc")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/refresh", nil)
		q := req.URL.Query()
		q.Add("sessionToken", mockToken)
		q.Add("refreshToken", "abc")

		req.URL.RawQuery = q.Encode()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("proper call to refreshSession didn't return 200, instead: %v", w.Code)
		}

		body, _ := ioutil.ReadAll(w.Body)

		var j map[string]interface{}

		json.Unmarshal(body, &j)

		if j["session"].(string) != sessionCode {
			t.Errorf("refreshSession returned unequal session. Got: %v, Expected: %v", j["session"], sessionCode)
		}

		if j["expiresIn"].(float64) != time.Hour.Seconds() {
			t.Errorf("refreshSession returned different expiresIn. Got: %v, Expected: %v", j["expiresIn"].(int), time.Hour.Seconds())
		}

		newRefreshToken, _ := mr.Get("session:" + sessionCode)

		if j["refreshToken"].(string) != newRefreshToken {
			t.Errorf("refreshToken returned different refeshTokens. Got: %v, Expected: %v", j["refreshToken"].(string), newRefreshToken)
		}
	})

}
