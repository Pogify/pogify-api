package pogifyapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

func Test_server_postUpdate(t *testing.T) {
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
	os.Setenv("PUBSUB_SECRET", _pubsubsecret)

	router := gin.Default()

	ServerV1(router.Group("/"))

	t.Run("empty call", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/postUpdate", nil)
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("returned code %v, expected %v", w.Code, 400)
		}
	})
	t.Run("invalid token", func(t *testing.T) {
		invalidToken, _ := token.SignedString([]byte("aaaa"))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/postUpdate", nil)
		req.Header.Add("X-Session-Token", invalidToken)
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Errorf("returned code %v, expected %v", w.Code, 401)
		}
	})
	t.Run("proper token, empty body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/postUpdate", nil)
		req.Header.Add("x-session-token", mockToken)
		router.ServeHTTP(w, req)

		if w.Code != 500 {
			t.Errorf("returned code %v, expected %v", w.Code, 500)
		}
	})

	t.Run("proper token, body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/postUpdate", strings.NewReader("{\"nothing\":\"nothing\"}"))
		req.Header.Add("x-session-token", mockToken)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("returned code %v, expected %v", w.Code, 200)
		}
	})

}
