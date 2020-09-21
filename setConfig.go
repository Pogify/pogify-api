package main

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type config struct {
	RequestInterval int `json:"requestInterval" binding:"required"`
}

func (s *server) setConfig(c *gin.Context) {
	sessionToken := c.GetHeader("X-Session-Token")
	if sessionToken == "" {
		c.String(400, "missing X-Session-Token header")
		return
	}

	token, err := jwt.Parse(sessionToken, func(t *jwt.Token) (interface{}, error) {
		return s.jwt.secret, nil
	})

	if err != nil {
		c.Error(err)
		c.String(401, fmt.Sprint(err))
		return
	}

	sessionID := token.Claims.(jwt.MapClaims)["session"].(string)

	var conf config
	err = c.ShouldBindJSON(&conf)
	if err != nil {
		c.Error(err)
		c.String(400, fmt.Sprint(err))
		return
	}

	err = s.redis.setSessionConfig(sessionID, conf)

	if err != nil {
		c.AbortWithError(500, err)
	} else {
		c.String(200, "ok")
	}

}
