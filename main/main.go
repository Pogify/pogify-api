package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	v1 "github.com/pogify/pogify-api"
)

func main() {
	s := startServer()
	s.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func startServer() *gin.Engine {
	s := gin.Default()

	if os.Getenv("V1") != "" {
		v1.ServerV1(s.Group("/v1"))
	}

	return s
}
