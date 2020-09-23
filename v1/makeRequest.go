package v1

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type request struct {
	Session  string `json:"session" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	Token    string `json:"token" binding:"required"`
	Request  string `json:"request" binding:"required"`
}

// var rx = /^.*(?:(?:youtu\.be\/|v\/|vi\/|u\/\w\/|embed\/)|(?:(?:watch)?\?v(?:i)?=|\&v(?:i)?=))([^#\&\?]*).*/;

func (r *request) validateRequest() {
	// r.Request
}

func (s *server) makeRequest(c *gin.Context) {
	var r request

	err := c.ShouldBindJSON(&r)

	if err != nil {
		c.Error(err)
		c.String(400, fmt.Sprint(err))
		return
	}

	// var for identifier
	var id string
	// validate token against provider
	switch r.Provider {
	case "twitch":
		// validate token with twitch
		// s.auth.getTwitchKeys()
		token, err := s.auth.ValidateTwitchToken(r.Token)
		if err != nil {
			c.Error(err)
			c.String(401, fmt.Sprint(err))
			return
		}
		id = token.Claims.(jwt.MapClaims)["sub"].(string)
	case "google":
		// validate token with google
		s.auth.getGooglePEM()
		token, err := s.auth.ValidateGoogleToken(r.Token)
		if err != nil {
			c.Error(err)
			c.String(401, fmt.Sprint(err))
			return
		}
		id = token.Claims.(jwt.MapClaims)["sub"].(string)
	default:
		c.String(400, "invalid provider")
		return
	}

	t0 := time.Now()
	// eager increment rate limit
	rateLimit, err := s.redis.rateLimitRequest(r.Session, id)
	log.Print(time.Now().Sub(t0))

	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if rateLimit[0] > 1 {
		c.Header("retry-after", fmt.Sprint(rateLimit[1]))
		c.Status(429)
		return
	}

	// check active session
	res, _ := http.Get(fmt.Sprintf("%v/channels-stats?id=%v", s.pubsub.url, r.Session))
	if err != nil {
		go s.redis.reverseRateLimit(r.Session, id)
		c.AbortWithError(500, err)
		return
	}
	if res.StatusCode == 404 {
		c.String(400, "inactive session")
		return
	}

	tA := time.Now()
	ch := make(chan *http.Response)
	errCh := make(chan error)

	go s.pubsub.pub(ch, errCh, "host_"+r.Session, []byte(r.Request))

	select {
	case <-ch:
	case err = <-errCh:
		c.AbortWithError(500, err)

	}
	log.Print(time.Now().Sub(tA))
}
