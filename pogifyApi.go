package pogifyapi

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var _testing = false

func init() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET missing in .env. Server will use empty string as secret")
	}

	if os.Getenv("REDIS_URI") == "" {
		log.Println("REDIS_URI missing in .env. Server will use localhost:6379 instead")
	}

	if os.Getenv("REFRESH_TOKEN_TTL") == "" {
		log.Println("REFRESH_TOKEN_TTL missing in .env. Server will use 1 hour.")
	}

	if os.Getenv("PUBSUB_SECRET") == "" {
		log.Println("PUBSUB_SECRET missing in .env. Server will use empty string as secret")
	}

	if os.Getenv("PUBSUB_URL") == "" {
		if !_testing {
			panic("PUBSUB_URL missing in .env. Add it and restart the server.")
		}
	}
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

// Server sets routes for  version 1 of api
func Server(rr *gin.RouterGroup) {
	var s = new(server)
	var r = new(r)

	redisURI := os.Getenv("REDIS_URI")
	if len(redisURI) == 0 {
		redisURI = "redis://localhost:6379"
	}

	opts, err := redis.ParseURL(redisURI)
	if err != nil {
		panic(err)
	}

	r.conn = redis.NewClient(opts)

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
	go a.getGooglePEM()
	go a.getTwitchKeys()
	s.auth = a

	sessionEndpoints := rr.Group("/session")
	{
		sessionEndpoints.POST("/start", s.startSession)

		sessionEndpoints.POST("/refresh", s.refreshSession)

		sessionEndpoints.OPTIONS("/update", s.cors)
		sessionEndpoints.POST("/update", s.postUpdate)

		sessionEndpoints.OPTIONS("/request", s.cors)
		sessionEndpoints.POST("/request", s.makeRequest)

		sessionEndpoints.POST("/config", s.setConfig)
		sessionEndpoints.OPTIONS("/config", s.cors)

		sessionEndpoints.GET("/config", s.getConfig)
	}
	rr.POST("/auth/twitch", s.twitchAuth)
}
