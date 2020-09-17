package main

import (
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	gonanoid "github.com/matoous/go-nanoid"
)

type sessionJwtClaims struct {
	Session string `json:"session"`
	jwt.StandardClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// StartSession ...
func StartSession(c *gin.Context) {

	sessionCode, err := generateSessionCode()
	if err != nil {
		fmt.Println("err generating Sessioncode")
		return
	}

	refreshToken, err := gonanoid.ID(64)

	claims := sessionJwtClaims{
		sessionCode,
		jwt.StandardClaims{
			ExpiresAt: 60 * 60,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSign, err := token.SignedString(jwtSecret)

	fmt.Println(claims)

	if err != nil {
		fmt.Println(err)
	}

	c.JSON(200, gin.H{
		"message":      sessionCode,
		"refreshToken": refreshToken,
		"token":        tokenSign,
	})
}

func generateSessionCode() (string, error) {
	return gonanoid.Generate("abcdefghijklmnopqrstuwxyz0123456789-", 5)
}
