package pogifyapi

import (
	"errors"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

type sessionJwtClaims struct {
	Session string `json:"session"`
	jwt.StandardClaims
}

// StartSession ...
func (s *server) claimSession(c *gin.Context) {

	var sessionCode string
	sid, exists := c.Get("sessionID")
	if !exists {
		c.AbortWithError(500, errors.New("sessionID not set in context"))
		return
	}
	sessionCode = sid.(string)

	refreshToken, err := gonanoid.ID(64)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	val, err := s.redis.newSession(sessionCode, refreshToken)

	if err != nil {
		log.Print(err)
		c.AbortWithError(500, err)
		return
	}

	if val != 1 {
		c.String(410, "code taken")
		return
	}

	claims := sessionJwtClaims{
		sessionCode,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 60*60,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSign, err := token.SignedString(s.jwt.secret)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{
		"session":      sessionCode,
		"refreshToken": refreshToken,
		"expiresIn":    60 * 60,
		"token":        tokenSign,
	})
}
