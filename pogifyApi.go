package pogifyapi

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis/v8"
	ginpow "github.com/jeongy-cho/gin-pow"
	gonanoid "github.com/matoous/go-nanoid"
)

var _testing = false

func init() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET missing in .env. Server will use empty string as secret")
	}

	if os.Getenv("REDIS_URI") == "" {
		log.Println("REDIS_URI missing in .env. Server will use localhost:6379 instead")
	}

	if os.Getenv("REFRESH_TOKEN_TTL") == "" {
		log.Println("REFRESH_TOKEN_TTL missing in .env. Server will use 1 hour.")
	}

	if os.Getenv("PUBSUB_SECRET") == "" {
		log.Println("PUBSUB_SECRET missing in .env. Server will use empty string as secret")
	}

	if os.Getenv("PUBSUB_URL") == "" {
		if !_testing {
			panic("PUBSUB_URL missing in .env. Add it and restart the server.")
		}
	}

	if os.Getenv("POW_SECRET") == "" {
		log.Println("POW_SECRET missing in .env. Server will use random string as secret")
	}

	if os.Getenv("POW_DIFFICULTY") == "" {
		log.Println("POW_DIFFICULTY missing in .env. Server will use difficulty 0")

	} else {
		if _, err := strconv.Atoi(os.Getenv("POW_DIFFICULTY")); err != nil {
			log.Println("Can't parse POW_DIFFICULTY to int, server will use difficulty 0")
		}
	}
}

type server struct {
	redis  *r
	pubsub *pubsub
	jwt    *j
	auth   *auth
	pow    *ginpow.Middleware
}

func (s *server) cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "X-Session-Token,Content-Type")
	c.Header("Access-Control-Max-Age", "7200")
}

type j struct {
	secret []byte
}

// Server sets routes for api
func Server(rr *gin.RouterGroup) {
	var s = new(server)
	var r = new(r)

	redisURI := os.Getenv("REDIS_URI")
	if len(redisURI) == 0 {
		redisURI = "redis://localhost:6379"
	}

	opts, err := redis.ParseURL(redisURI)
	if err != nil {
		panic(err)
	}

	r.conn = redis.NewClient(opts)

	if ttl := os.Getenv("REFRESH_TOKEN_TTL"); ttl != "" {
		r.refreshTokenTTL = os.Getenv("REFRESH_TOKEN_TTL")
	} else {
		r.refreshTokenTTL = fmt.Sprint(60 * 60)
	}

	s.redis = r

	var p = new(pubsub)
	p.secret = os.Getenv("PUBSUB_SECRET")
	p.url = os.Getenv("PUBSUB_URL")
	s.pubsub = p

	var j = new(j)
	j.secret = []byte(os.Getenv("JWT_SECRET"))
	s.jwt = j

	var a = new(auth)
	go a.getGooglePEM()
	go a.getTwitchKeys()
	s.auth = a

	powDiff, _ := strconv.Atoi(os.Getenv("POW_DIFFICULTY"))

	s.pow, err = ginpow.New(&ginpow.Middleware{
		ExtractAll: extractAll,
		Check:      true,
		Secret:     os.Getenv("POW_SECRET"),
		Difficulty: powDiff,

		NonceGenerator: generateSessionCode,

		NonceContextKey:          "nonce",
		NonceChecksumContextKey:  "checksum",
		HashDifficultyContextKey: "hashDifficulty",
	})
	if err != nil {
		panic(err)
	}

	sessionEndpoints := rr.Group("/session")
	{
		sessionEndpoints.GET("/issue", s.pow.GenerateNonceMiddleware, s.GenerateProblem)

		sessionEndpoints.OPTIONS("/claim", s.cors)
		sessionEndpoints.POST("/claim", s.pow.VerifyNonceMiddleware, s.claimSession)

		sessionEndpoints.OPTIONS("/refresh", s.cors)
		sessionEndpoints.POST("/refresh", s.refreshSession)

		sessionEndpoints.OPTIONS("/update", s.cors)
		sessionEndpoints.POST("/update", s.postUpdate)

		sessionEndpoints.OPTIONS("/request", s.cors)
		sessionEndpoints.POST("/request", s.makeRequest)

		sessionEndpoints.OPTIONS("/config", s.cors)
		sessionEndpoints.GET("/config", s.getConfig)
		sessionEndpoints.POST("/config", s.setConfig)
	}
	rr.POST("/auth/twitch", s.twitchAuth)
}

func generateSessionCode(_ int) (string, error) {
	// testing flag for predictable keys
	if _testing {
		return "test1.123", nil
	}

	nonce, err := gonanoid.Generate("abcdefghijklmnopqrstuwxyz0123456789-", 5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v.%v", nonce, time.Now().Unix()), nil
}

// Time is a JSON un/marshallable type of time.Time
type Time time.Time

// MarshalJSON is used to convert the timestamp to JSON
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *Time) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

func extractAll(c *gin.Context) (nonce string, nonceChecksum string, data string, hash string, err error) {
	b := new(SessionClaim)

	if err = c.ShouldBindBodyWith(b, binding.JSON); err != nil {
		c.Error(err)
		c.String(400, err.Error())
		c.Abort()
		return
	}
	c.Set("sessionID", b.SessionID)

	if dif := time.Since(time.Time(b.Issued)); dif > time.Minute {
		err = fmt.Errorf("problem expired by %v", dif)
		c.Error(err)
		c.String(400, err.Error())
		c.Abort()
		return
	}

	nonce = fmt.Sprintf("%v.%v", b.SessionID, strconv.FormatInt(time.Time(b.Issued).Unix(), 10))
	nonceChecksum = b.Checksum
	data = b.Solution
	hash = b.Hash
	return
}
