package pogifyapi

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"	gonanoid "github.com/matoous/go-nanoid"
)


func (s *server) generateProblem(c *gin.Context) {
	sessionId, _ := c.Get("sessionId")
	checksum, _ := c.Get("checksum")
	
	d := gin.H{
		"sessionId": sessionId,
		"checksum": checksum,
		"issued": time.Now().Unix(),

	}
}


func generateSessionCode(_ int) (string, error) {
	// testing flag for predictable keys
	if _testing {
		if rand.Float32() < 0.1 {
			return "test1", nil
		}
		return "test2", nil
	}

	if nonce, err :=	gonanoid.Generate("abcdefghijklmnopqrstuwxyz0123456789-", 5); err != nonce {}
}
