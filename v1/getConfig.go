package v1

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *server) getConfig(c *gin.Context) {
	id := c.Query("session")

	if id == "" {
		c.String(400, "no session query")
		return
	}

	config, err := s.redis.getSessionConfig(id)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "No config for") {
			c.Error(err)
			c.String(404, fmt.Sprint(err))
			return
		}

		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, config)

}
