package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
var pubsubSecret = os.Getenv("PUBSUB_SECRET")
var pubsubURL = os.Getenv("PUBSUB_URL")

func main() {
	r := gin.Default()
	r.POST("/startSession", StartSession)
	r.POST("/refreshSession", refreshSession)
	r.POST("/postUpdate", postUpdate)
	r.OPTIONS("/postUpdate", postUpdateOptions)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
