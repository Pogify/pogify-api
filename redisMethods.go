package main

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_HOST"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
})

var newSessionScript = `local c = redis.call("ttl", KEYS[1])
												if (c < 0) then 
													redis.call("set", KEYS[1], ARGV[1])
													redis.call("expire", KEYS[1], ARGV[2])
													return 1
													end
												return 0`

// NewSession ...
func NewSession(sessionToken string, refreshToken string) (int64, error) {

	key := []string{
		"session:" + sessionToken,
	}

	val, err := rdb.Eval(ctx, newSessionScript, key, refreshToken, os.Getenv("REFRESH_TOKEN_TTL")).Result()
	if err != nil {
		log.Print(err)
	}

	return val.(int64), err
}

var verifyAndSetScript = `
  local t = redis.call("get", KEYS[1])
  if (t == false) then
    return -1
  end
  if (t == ARGV[1]) then
    redis.call("set", KEYS[1], ARGV[2])
    redis.call("expire", KEYS[1], ARGV[3])
    return 1
  end
  return 0 
  `

func verifyAndSetNewSession(sessionID string, token string, newToken string) (int64, error) {
	val, err := rdb.Eval(ctx, verifyAndSetScript, []string{"session:" + sessionID}, token, newToken, os.Getenv("REFRESH_TOKEN_TTL")).Result()
	return val.(int64), err
}
