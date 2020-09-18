package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type r struct {
	conn            *redis.Client
	refreshTokenTTL string
}

var newSessionScript = `local c = redis.call("ttl", KEYS[1])
												if (c < 0) then 
													redis.call("set", KEYS[1], ARGV[1])
													redis.call("expire", KEYS[1], ARGV[2])
													return 1
													end
												return 0`

func (r *r) newSession(sessionToken string, refreshToken string) (int64, error) {

	key := []string{
		"session:" + sessionToken,
	}

	val, err := r.conn.Eval(ctx, newSessionScript, key, refreshToken, r.refreshTokenTTL).Result()
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

func (r *r) verifyAndSetNewSession(sessionID string, token string, newToken string) (int64, error) {
	val, err := r.conn.Eval(ctx, verifyAndSetScript, []string{"session:" + sessionID}, token, newToken, r.refreshTokenTTL).Result()
	return val.(int64), err
}

var requestLimitScript = `
	local c = redis.call('incr',KEYS[1]) 
	if (c == 1) then 
		redis.call('expire', KEYS[1], ARGV[1]) 
	end 	
	return {c, redis.call('ttl', KEYS[1])}`

func (r *r) rateLimitRequest(sessionID string, id string) (int64, error) {
	key := fmt.Sprintf("requestLimit:%v:%v", sessionID, id)
	// TODO: this should be host configured
	val, err := r.conn.Eval(ctx, requestLimitScript, []string{key}, os.Getenv("REQUEST_INTERVAL")).Result()
	return val.(int64), err
}

func (r *r) reverseRateLimit(sessionID string, id string) (int64, error) {
	return r.conn.Decr(ctx, fmt.Sprintf("requestLimit:%v:%v", sessionID, id)).Result()
}
