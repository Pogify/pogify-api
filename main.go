package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	v1 "github.com/pogify/pogify-api/v1"
)

func main() {
	s := startServer()
	s.Run(":8081") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func startServer() *gin.Engine {
	s := gin.Default()

	v1.ServerV1(s.Group("/v1"))

	return s
}
