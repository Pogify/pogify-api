package v1

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	if os.Getenv("TWITCH_CLIENT_ID") == "" {
		log.Print("missing TWITCH_CLIENT_ID in .env")
	}
	if os.Getenv("TWITCH_CLIENT_SECRET") == "" {
		log.Print("missing TWITCH_CLIENT_SECRET in .env")

	}
}

func (s *server) twitchAuth(c *gin.Context) {
	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", nil)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	q := req.URL.Query()
	q.Add("client_id", os.Getenv("TWITCH_CLIENT_ID"))
	q.Add("client_secret", os.Getenv("TWITCH_CLIENT_SECRET"))
	q.Add("code", c.Query("code"))
	q.Add("grant_type", "authorization_code")
	q.Add("redirect_uri", "http://localhost:3006/auth/twitch")
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// body, _ := ioutil.ReadAll(res.Body)
	c.Header("Access-Control-Allow-Origin", "*")
	c.DataFromReader(200, res.ContentLength, "application/json", res.Body, *new(map[string]string))

}
