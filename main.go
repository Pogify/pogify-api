package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET missing in .env. Server will use empty string as secret")
	}

	if os.Getenv("REDIS_HOST") == "" {
		log.Println("REDIS_HOST missing in .env. Server will use localhost:6379 instead")
	}

	if os.Getenv("REFRESH_TOKEN_TTL") == "" {
		log.Println("REFRESH_TOKEN_TTL missing in .env. Server will use 1 hour.")
	}

	if os.Getenv("PUBSUB_SECRET") == "" {
		log.Println("PUBSUB_SECRET missing in .env. Server will use empty string as secret")
	}

	if os.Getenv("PUBSUB_URL") == "" {
		panic("PUBSUB_URL missing in .env. Add it and restart the server.")
	}
}

func main() {
	s := startServer()
	s.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type server struct {
	redis  *r
	pubsub *pubsub
	jwt    *j
}

func (s *server) cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "X-Session-Token,Content-Type")
	c.Header("Access-Control-Max-Age", "7200")
}

type j struct {
	secret string
}

type pubsub struct {
	secret string
	url    string
}

func (p *pubsub) pub(ch chan<- *http.Response, errCh chan<- error, channel string, data []byte) {
	pub, err := http.NewRequest("POST", p.url, bytes.NewReader(data))
	pub.Header.Add("authorization", p.secret)
	pubQ := pub.URL.Query()
	pubQ.Add("id", channel)
	pub.URL.RawQuery = pubQ.Encode()

	if err != nil {
		ch <- nil
		errCh <- err
		return
	}

	res, err := http.DefaultClient.Do(pub)
	if err != nil {
		ch <- nil
		errCh <- err
		return
	}

	ch <- res
	errCh <- nil

}

func startServer() *gin.Engine {
	var s = new(server)
	var r = new(r)

	r.conn = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if ttl := os.Getenv("REFRESH_TOKEN_TTL"); ttl != "" {
		r.refreshTokenTTL = os.Getenv("REFRESH_TOKEN_TTL")
	} else {
		r.refreshTokenTTL = fmt.Sprint(60 * 60)
	}

	s.redis = r

	var p = new(pubsub)
	p.secret = os.Getenv("PUBSUB_SECRET")
	p.url = os.Getenv("PUBSUB_URL")
	s.pubsub = p

	var j = new(j)
	j.secret = os.Getenv("JWT_SECRET")
	s.jwt = j

	rr := gin.Default()
	rr.POST("/startSession", s.startSession)
	rr.POST("/refreshSession", s.refreshSession)
	rr.OPTIONS("/postUpdate", s.cors)
	rr.POST("/postUpdate", s.postUpdate)
	rr.POST("/makeRequest", s.makeRequest)
	rr.OPTIONS("/makeRequest", s.cors)

	return rr
}
