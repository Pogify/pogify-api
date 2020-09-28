package pogifyapi

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *server) GenerateProblem(c *gin.Context) {
	nonce, nErr := c.Get("nonce")
	if !nErr {
		c.AbortWithError(500, errors.New("no nonce"))
		return
	}

	checksum, ncErr := c.Get("checksum")
	if !ncErr {
		c.AbortWithError(500, errors.New("no nonce checksum"))
		return
	}

	splitNonce := strings.Split(nonce.(string), ".")
	issued, _ := strconv.Atoi(splitNonce[1])

	d := gin.H{
		"sessionId":  splitNonce[0],
		"checksum":   checksum,
		"issued":     issued,
		"difficulty": s.pow.Difficulty,
	}

	c.Negotiate(200, gin.Negotiate{
		Offered: []string{gin.MIMEJSON, gin.MIMEXML},
		Data:    d,
	})
}
