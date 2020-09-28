package pogifyapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func Test_server_claimSession(t *testing.T) {
	mr, err := miniredis.Run()
	defer mr.Close()
	if err != nil {
		t.Fatalf("MiniRedis error: %s", err)
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	router := gin.Default()

	Server(router.Group("/"))
	// var key1 string
	t.Run("Test /session/claim returns 400 on empty request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/claim", strings.NewReader(""))
		router.ServeHTTP(w, req)

		if expect := http.StatusBadRequest; w.Code != expect {
			t.Errorf("expected error with: %v but got %v instead", expect, w.Code)
		}

	})
	t.Run("Test /session/claim returns 400 on bind error", func(t *testing.T) {

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/session/claim", strings.NewReader("{\"nothing\":\"nothing\"}"))
		router.ServeHTTP(w, req)

		if expect := http.StatusBadRequest; w.Code != expect {
			t.Errorf("expected error with: %v but got %v instead", expect, w.Code)
		}
	})

	t.Run("Test /session/claim", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/session/issue", strings.NewReader("{\"nothing\":\"nothing\"}"))
		router.ServeHTTP(w1, req1)

		p := struct {
			SessionID  string `json:"sessionId" binding:"required"`
			Issued     int64  `json:"issued" binding:"required"`
			Checksum   string `json:"checksum" binding:"required"`
			Difficulty int    `json:"difficulty" binding:"required"`
		}{}

		if e := json.Unmarshal(w1.Body.Bytes(), &p); e != nil {
			t.Error(e)
			return
		}
		p.Issued = time.Now().Unix()
		idIss := fmt.Sprintf("%v.%v", p.SessionID, strconv.FormatInt(p.Issued, 10))
		cs := sha256.Sum256([]byte(idIss + os.Getenv("POW_SECRET")))

		s, h := findSolution(idIss, p.Difficulty)
		sol := struct {
			SessionID string `json:"sessionId" binding:"required"`
			Issued    int64  `json:"issued" binding:"required"`
			Checksum  string `json:"checksum" binding:"required"`
			Solution  string `json:"solution" binding:"required"`
			Hash      string `json:"hash" binding:"required"`
		}{
			SessionID: p.SessionID,
			Issued:    p.Issued,
			Checksum:  hex.EncodeToString(cs[:]),
			Solution:  s,
			Hash:      h,
		}

		jsonSol, _ := json.Marshal(sol)
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/session/claim", bytes.NewReader(jsonSol))
		router.ServeHTTP(w2, req2)

		session := struct {
			ExpiresIn int    `json:"expiresIn" binding:"required"`
			RT        string `json:"refreshToken" binding:"required"`
			Session   string `json:"session" binding:"required"`
			Token     string `json:"token" binding:"required"`
		}{}
		if err := json.Unmarshal(w2.Body.Bytes(), &session); err != nil {
			t.Error(err)
		}
	})
	t.Run("Test already claimed session /session/claim", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/session/issue", strings.NewReader("{\"nothing\":\"nothing\"}"))
		router.ServeHTTP(w1, req1)

		p := struct {
			SessionID  string `json:"sessionId" binding:"required"`
			Issued     int64  `json:"issued" binding:"required"`
			Checksum   string `json:"checksum" binding:"required"`
			Difficulty int    `json:"difficulty" binding:"required"`
		}{}

		if e := json.Unmarshal(w1.Body.Bytes(), &p); e != nil {
			t.Error(e)
			return
		}
		p.Issued = time.Now().Unix()
		idIss := fmt.Sprintf("%v.%v", p.SessionID, strconv.FormatInt(p.Issued, 10))
		cs := sha256.Sum256([]byte(idIss + os.Getenv("POW_SECRET")))

		s, h := findSolution(idIss, p.Difficulty)

		sol := struct {
			SessionID string `json:"sessionId" binding:"required"`
			Issued    int64  `json:"issued" binding:"required"`
			Checksum  string `json:"checksum" binding:"required"`
			Solution  string `json:"solution" binding:"required"`
			Hash      string `json:"hash" binding:"required"`
		}{
			SessionID: p.SessionID,
			Issued:    p.Issued,
			Checksum:  hex.EncodeToString(cs[:]),
			Solution:  s,
			Hash:      h,
		}

		jsonSol, _ := json.Marshal(sol)
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/session/claim", bytes.NewReader(jsonSol))
		router.ServeHTTP(w2, req2)

		if expect := http.StatusGone; w2.Code != expect {
			t.Errorf("got unexpected response code %v, expected %v", w2.Code, expect)
		}
	})

}

func findSolution(nonce string, difficulty int) (s string, hash string) {
	s = "0"
	prefix := strings.Repeat("0", difficulty)

	for {
		h := sha256.Sum256([]byte(s + nonce))
		hash = hex.EncodeToString(h[:])

		if strings.HasPrefix(hash, prefix) {
			return
		}

		s += "0"
	}

}
