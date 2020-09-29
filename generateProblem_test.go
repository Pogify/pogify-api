package pogifyapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	ginpow "github.com/jeongy-cho/gin-pow"
)

func Test_server_GenerateProblem(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		s := new(server)
		s.pow, _ = ginpow.New(&ginpow.Middleware{
			ExtractData: func(c *gin.Context) (string, error) { return "", nil },
		})

		m, err := ginpow.New(&ginpow.Middleware{
			Check:                   true,
			Secret:                  "secret",
			NonceChecksumContextKey: "checksum",
			ExtractData:             func(c *gin.Context) (string, error) { return "", nil },
			NonceGenerator:          func(i int) ([]byte, error) { return []byte("test1.123"), nil },
		})

		if err != nil {
			t.Error("error when making middleware", err.Error())
		}
		c.Accepted = []string{gin.MIMEJSON}
		m.GenerateNonceMiddleware(c)
		s.GenerateProblem(c)

		ret := struct {
			SessionID string `json:"sessionId"`
			Checksum  string `json:"checksum"`
			Issued    int    `json:"issued"`
		}{}
		expect := struct {
			SessionID string `json:"sessionId"`
			Checksum  string `json:"checksum"`
			Issued    int    `json:"issued"`
		}{
			SessionID: "test1",
			Checksum:  "9778626ec2d8370d4dde0e1d080d12c419ae4224f7e7cb94a1da2f606159aa98",
			Issued:    123,
		}

		json.Unmarshal(w.Body.Bytes(), &ret)

		if !reflect.DeepEqual(ret, expect) {
			t.Errorf("not equal; Got: %+v, Expected: %+v", ret, expect)
		}
	})
	t.Run("no nonce", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		s := new(server)

		c.Accepted = []string{gin.MIMEJSON}
		s.GenerateProblem(c)

		if expect := http.StatusInternalServerError; w.Code != expect {
			t.Errorf("server returned status %v; Expected: %v", w.Code, expect)
		}

	})
	t.Run("no checksum", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		s := new(server)
		m, err := ginpow.New(&ginpow.Middleware{
			NonceChecksumContextKey: "checksum",
			ExtractData:             func(c *gin.Context) (string, error) { return "", nil },
			NonceGenerator:          func(i int) ([]byte, error) { return []byte("test1.123"), nil },
		})

		if err != nil {
			t.Error("error when making middleware", err.Error())
		}
		c.Accepted = []string{gin.MIMEJSON}
		m.GenerateNonceMiddleware(c)
		s.GenerateProblem(c)

		if expect := http.StatusInternalServerError; w.Code != expect {
			t.Errorf("server returned status %v; Expected: %v", w.Code, expect)
		}
	})

}

func Test_server_GenerateProblem_integration(t *testing.T) {
	mr, err := miniredis.Run()
	defer mr.Close()
	if err != nil {
		t.Fatalf("MiniRedis error: %s", err)
	}
	os.Setenv("REDIS_URI", "redis://"+mr.Addr())

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	// r := gin.Default()
	Server(r.Group("/"))

	req := httptest.NewRequest("GET", "/session/issue", strings.NewReader(""))
	// req.Header.Add("Accept", "*/*")

	r.ServeHTTP(w, req)

	p := struct {
		SessionID  string `json:"sessionId" binding:"required"`
		Issued     int    `json:"issued" binding:"required"`
		Checksum   string `json:"checksum" binding:"required"`
		Difficulty int    `json:"difficulty" binding:"required"`
	}{}
	if e := json.Unmarshal(w.Body.Bytes(), &p); e != nil {
		t.Error(e)
		return
	}

	if p.SessionID != "test1" && p.SessionID != "test2" {
		t.Errorf("returned session id is not test1 or test2: %v", p.SessionID)
	}

	if p.Issued != 123 {
		t.Errorf("returned unexpected %v, expected: 123", p.Issued)
	}

	if expected := "9778626ec2d8370d4dde0e1d080d12c419ae4224f7e7cb94a1da2f606159aa98"; p.Checksum != expected {
		t.Errorf("returned unexpected checksum; got: %v, expected: %v", p.Checksum, expected)
	}

	if expected := 1; p.Difficulty != expected {
		t.Errorf("returned unexpected difficulty; got: %v, expected: %v", p.Difficulty, expected)
	}
}
