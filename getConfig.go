package main

import "github.com/gin-gonic/gin"

func (s *server) getConfig(c *gin.Context) {
	id := c.Query("session")

	if id == "" {
		c.String(400, "no session query")
		return
	}

	config, err := s.redis.getSessionConfig(id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, config)

}
