package main

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

type sessionJwtClaims struct {
	Session string `json:"session"`
	jwt.StandardClaims
}

// StartSession ...
func StartSession(c *gin.Context) {
	sessionCode, err := generateSessionCode()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	refreshToken, err := gonanoid.ID(64)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	for true {
		val, err := NewSession(sessionCode, refreshToken)

		if err != nil {
			fmt.Print(err)
			c.AbortWithError(500, err)
			return
		}

		if val == 1 {
			break
		} else {
			sessionCode, err = generateSessionCode()
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
		}
	}

	claims := sessionJwtClaims{
		sessionCode,
		jwt.StandardClaims{
			ExpiresAt: 60 * 60,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSign, err := token.SignedString(jwtSecret)

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

func generateSessionCode() (string, error) {
	return gonanoid.Generate("abcdefghijklmnopqrstuwxyz0123456789-", 5)
}
