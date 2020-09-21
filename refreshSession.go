package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

// RefreshSession ...
func (s *server) refreshSession(c *gin.Context) {
	sessionJWT := c.Query("sessionToken")
	if sessionJWT == "" {
		c.String(400, "missing query: sessionToken")
		return
	}

	refreshToken := c.Query("refreshToken")
	if refreshToken == "" {
		c.String(400, "missing query: refreshToken")
	}

	token, err := jwt.Parse(sessionJWT, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorExpired == 0 {
			c.AbortWithError(400, err)
			return
		}
	}

	sessionID := token.Claims.(jwt.MapClaims)["session"].(string)

	newRefreshToken, err := gonanoid.ID(64)

	val, err := s.redis.verifyAndSetNewRefreshToken(sessionID, refreshToken, newRefreshToken)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	switch val {
	case -1:
		c.String(400, "refresh token expired")
	case 0:
		c.String(400, "invalid refreshToken")
	case 1:
		claims := sessionJwtClaims{
			sessionID,
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
			"session":      sessionID,
			"refreshToken": newRefreshToken,
			"token":        tokenSign,
			"expiresIn":    3000,
		})
	}

}
