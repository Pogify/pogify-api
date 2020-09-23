package v1

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func (s *server) postUpdate(c *gin.Context) {
	sessionToken := c.GetHeader("X-Session-Token")
	if sessionToken == "" {
		c.String(400, "missing X-Session-Token header")
		return
	}

	token, err := jwt.Parse(sessionToken, func(token *jwt.Token) (interface{}, error) {
		return s.jwt.secret, nil
	})

	if token.Valid {
		sessionID := token.Claims.(jwt.MapClaims)["session"].(string)
		data, _ := c.GetRawData()

		ch := make(chan *http.Response)
		errCh := make(chan error)

		go s.pubsub.pub(ch, errCh, sessionID, data)

		var res *http.Response
		select {
		case res = <-ch:
		case err := <-errCh:
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
		c.Error(ve)
		c.String(400, fmt.Sprint(ve))
	} else {
		c.AbortWithError(403, err)
	}

}
