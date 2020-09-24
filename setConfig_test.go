package pogifyapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func Test_server_setConfig(t *testing.T) {
	sessionCode := "test"
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
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	router := gin.New()

	ServerV1(router.Group("/"))

	t.Run("test without token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/setConfig", nil)
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("setConfig didn't return 400 without access token, instead: %v", w.Code)
		}

	})

	t.Run("test with token, no body", func(t *testing.T) {

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/setConfig", nil)
		req.Header.Add("X-Session-Token", mockToken)
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("setConfig didn't return 400, instead: %v", w.Code)
		}
	})

	t.Run("test with invalid token", func(t *testing.T) {
		invalidToken, _ := token.SignedString([]byte("aaaa"))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/setConfig", nil)
		req.Header.Add("X-Session-Token", invalidToken)
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Errorf("setConfig didn't return 401 on invalid token, instead: %v", w.Code)
		}
	})

	t.Run("test with body", func(t *testing.T) {

		conf := []byte("{\"requestInterval\":100}")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/setConfig", bytes.NewReader(conf))
		req.Header.Add("X-Session-Token", mockToken)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("setConfig didn't return 401 on invalid token, instead: %v", w.Code)
		}

		ri := mr.HGet("session:test:config", "RequestInterval")

		if ri != "100" {
			t.Error("setConfig didn't set config in redis")
		}

	})

}
