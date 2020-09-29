package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	v1 "github.com/pogify/pogify-api"
	v2 "github.com/pogify/pogify-api/v2"
)

func main() {
	s := startServer()
	s.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func startServer() *gin.Engine {
	s := gin.Default()

	if os.Getenv("V1") != "" {
		v1.Server(s.Group("/v1"))
	}

	if os.Getenv("V2") != "" {
		v2.Server(s.Group("/v2"))
	}

	return s
}
