package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func postUpdate(c *gin.Context) {
	sessionToken := c.GetHeader("X-Session-Token")
	if sessionToken == "" {
		c.String(400, "missing X-Session-Token header")
		return
	}

	token, err := jwt.Parse(sessionToken, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if token.Valid {
		sessionID := token.Claims.(jwt.MapClaims)["session"].(string)
		data, _ := c.GetRawData()
		fmt.Println(string(data))
		pub, err := http.NewRequest("POST", pubsubURL, bytes.NewReader(data))
		pub.Header.Add("authorization", pubsubSecret)
		pubQ := pub.URL.Query()
		pubQ.Add("id", sessionID)
		pub.URL.RawQuery = pubQ.Encode()

		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		res, err := http.DefaultClient.Do(pub)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		if res.StatusCode > 399 {
			log.Printf("Pubsub error with: %v", res.StatusCode)
			c.AbortWithStatus(500)
			return
		}

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Data(200, "application/json", body)
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors == jwt.ValidationErrorExpired {
			fmt.Print(ve.Errors, (jwt.ValidationErrorExpired))
			c.String(400, "sessionToken expired")
		} else {
			c.String(400, "invalid sessionToken")
		}
	} else {
		c.AbortWithError(403, err)
	}

}

func postUpdateOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "X-Session-Token")
	c.Header("Access-Control-Max-Age", "7200")
}
