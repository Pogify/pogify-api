package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func main() {
	r := gin.Default()
	r.POST("/startSession", StartSession)
	r.POST("/refreshSession", refreshSession)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
