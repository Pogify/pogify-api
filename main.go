package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	s.Run(":8081") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type server struct {
	redis  *r
	pubsub *pubsub
	jwt    *j
	auth   *auth
}

func (s *server) cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "X-Session-Token,Content-Type")
	c.Header("Access-Control-Max-Age", "7200")
}

type j struct {
	secret []byte
}

type pubsub struct {
	secret string
	url    string
}

func (p *pubsub) pub(ch chan<- *http.Response, errCh chan<- error, channel string, data []byte) {
	pub, err := http.NewRequest("POST", p.url+"/pub", bytes.NewReader(data))
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
		errCh <- err
		return
	}

	ch <- res

}

type auth struct {
	googlePEM       map[string]*rsa.PublicKey
	googlePEMExpiry time.Time
}

func (a *auth) getGooglePEM() map[string]*rsa.PublicKey {
	if len(a.googlePEM) > 0 && a.googlePEMExpiry.After(time.Now()) {
		return a.googlePEM
	}
	res, _ := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	body, _ := ioutil.ReadAll(res.Body)

	exp, _ := time.Parse(time.RFC1123, res.Header["Expires"][0])
	a.googlePEMExpiry = exp

	var r map[string]string
	json.Unmarshal(body, &r)

	rb := make(map[string]*rsa.PublicKey)

	for k, v := range r {
		pem, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(v))
		rb[k] = pem
	}
	a.googlePEM = rb
	return rb
}

func (a *auth) ValidateGoogleToken(t string) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		return a.getGooglePEM()[t.Header["kid"].(string)], nil
	})
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
	j.secret = []byte(os.Getenv("JWT_SECRET"))
	s.jwt = j

	var a = new(auth)
	a.getGooglePEM()
	s.auth = a

	rr := gin.Default()
	rr.POST("/startSession", s.startSession)
	rr.POST("/refreshSession", s.refreshSession)
	rr.OPTIONS("/postUpdate", s.cors)
	rr.POST("/postUpdate", s.postUpdate)
	rr.POST("/makeRequest", s.makeRequest)
	rr.OPTIONS("/makeRequest", s.cors)
	rr.POST("/setConfig", s.setConfig)
	rr.OPTIONS("/setConfig", s.cors)
	rr.GET("/getConfig", s.getConfig)

	return rr
}
