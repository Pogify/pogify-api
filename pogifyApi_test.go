package pogifyapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

var _ = (func() interface{} {
	_testing = true
	return nil
})()

var _pubsubsecret = "secret"

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	log.Print("main")
	mockPubSubHandler := gin.New()

	os.Setenv("POW_SECRET", "secret")
	os.Setenv("POW_DIFFICULTY", "1")

	mockPubSubHandler.POST("/pub", func(c *gin.Context) {
		if c.GetHeader("authorization") != _pubsubsecret {
			c.Status(401)
			return
		}

		body, _ := ioutil.ReadAll(c.Request.Body)
		log.Printf("PubSub got: %x", body)
		c.String(200, "ok")
	})

	mockPubSubHandler.GET("/channels-stats", func(c *gin.Context) {
		switch id := c.Query("id"); id {
		case "exist":
			c.Status(200)
		case "notexist":
			c.Status(404)
		}
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

	Server(router.Group("/"))

	cases := []struct {
		endpoint string
		method   string
	}{
		{"/session/issue", "GET"},
		{"/session/claim", "OPTIONS"},
		{"/session/claim", "POST"},
		{"/session/refresh", "OPTIONS"},
		{"/session/refresh", "POST"},
		{"/session/update", "OPTIONS"},
		{"/session/update", "POST"},
		{"/session/request", "OPTIONS"},
		{"/session/request", "POST"},
		{"/session/config", "OPTIONS"},
		{"/session/config", "GET"},
		{"/session/config", "POST"},
		{"/auth/twitch", "POST"},
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

func TestServer_pow_Middleware_ExtractAll(t *testing.T) {

	id, _ := gonanoid.Nanoid()
	iss := time.Now()
	cs, _ := gonanoid.Nanoid()
	sol, _ := gonanoid.Nanoid()
	h, _ := gonanoid.Nanoid()

	t.Run("clean", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		reqBody := struct {
			SessionID string `json:"sessionId"`
			Issued    Time   `json:"issued"`
			Checksum  string `json:"checksum"`
			Solution  string `json:"solution"`
			Hash      string `json:"hash"`
		}{
			SessionID: id,
			Issued:    Time(iss),
			Checksum:  cs,
			Solution:  sol,
			Hash:      h,
		}

		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("", "/", bytes.NewReader(body))
		nonce, nonceChecksum, data, hash, err := extractAll(c)

		idIss := fmt.Sprintf("%v.%v", id, strconv.FormatInt(iss.Unix(), 10))
		if idIss != nonce {
			t.Errorf("nonce doesn't match, Got: %v, Expected %v", nonce, idIss)
		}
		if nonceChecksum != cs {
			t.Errorf("nonceChecksome doesn't match, Got: %v, Expected %v", nonceChecksum, cs)
		}
		if sol != data {
			t.Errorf("solution doesn't match, Got: %v, Expected %v", data, sol)
		}
		if hash != h {
			t.Errorf("hash doesn't match, Got: %v, Expected %v", h, hash)
		}
		if err != nil {
			t.Errorf("got an error when shouldn't: %v", err)
		}

	})
	t.Run("bind error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		reqBody := struct {
			Issued   Time   `json:"issued"`
			Checksum string `json:"checksum"`
			Solution string `json:"solution"`
			Hash     string `json:"hash"`
		}{
			Issued:   Time(iss),
			Solution: sol,
			Checksum: cs,
			Hash:     h,
		}

		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("", "/", bytes.NewReader(body))
		extractAll(c)
		if expect := http.StatusBadRequest; w.Code != expect {
			t.Errorf("got status %v expected %v", w.Code, expect)
		}
	})
	t.Run("old", func(t *testing.T) {
		oldIss := time.Unix(0, 0)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		reqBody := struct {
			SessionID string `json:"sessionId"`
			Issued    Time   `json:"issued"`
			Checksum  string `json:"checksum"`
			Solution  string `json:"solution"`
			Hash      string `json:"hash"`
		}{
			SessionID: id,
			Issued:    Time(oldIss),
			Checksum:  cs,
			Solution:  sol,
			Hash:      h,
		}

		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("", "/", bytes.NewReader(body))
		extractAll(c)

		if expect := http.StatusBadRequest; w.Code != expect {
			t.Errorf("got status %v expected %v", w.Code, expect)
		}
	})
}

func Test_generateSessionCode(t *testing.T) {
	_testing = false
	defer func() {
		_testing = true
	}()

	nonce, err := generateSessionCode(1)
	if err != nil {
		panic(err)
	}

	split := strings.Split(string(nonce), ".")

	if expect := 5; len(split[0]) != expect {
		t.Errorf("different session id length: %v, expected %v", len(split[0]), expect)
	}

	if split[1] == "" {
		t.Errorf("no second part of nonce: %v", nonce)
	} else {
		if secondPartInt, err := strconv.Atoi(split[1]); err == nil {
			if secondPart := time.Unix(int64(secondPartInt), 0); time.Now().Sub(secondPart) > time.Second {
				t.Errorf("returned unexpected time; Got: %v, Expected: %v", secondPart, time.Now())
			}

		} else {
			t.Errorf("error parsing second part of %v; %v", nonce, err)
		}
	}
}
