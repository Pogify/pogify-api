package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/fatih/structs"
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
		local c = redis.call()
    return 1
  end
  return 0 
  `

func (r *r) verifyAndSetNewRefreshToken(sessionID string, token string, newToken string) (int64, error) {
	val, err := r.conn.Eval(ctx, verifyAndSetScript, []string{"session:" + sessionID}, token, newToken, r.refreshTokenTTL).Result()
	return val.(int64), err
}

var requestLimitScript = `
	local c = redis.call('incr',KEYS[1]) 
	local r = redis.call('hget', KEYS[1]..":config", "RefreshInterval")
	if (c == 1) then 
		if (r == 0) then 
			redis.call('expire', KEYS[1], r) 
		else 
			redis.call('expire', KEYS[1], 60)
		end
	end 	
	return {c, redis.call('ttl', KEYS[1])}`

func (r *r) rateLimitRequest(sessionID string, id string) ([]int64, error) {
	hash := sha1.New()
	hash.Write([]byte(id))
	bs := hash.Sum(nil)
	key := fmt.Sprintf("requestLimit:%v:%x", sessionID, bs)
	val, err := r.conn.Eval(ctx, requestLimitScript, []string{key}).Result()

	valS := make([]int64, 0, 2)
	for _, v := range val.([]interface{}) {
		valS = append(valS, v.(int64))
	}

	return valS, err
}

func (r *r) reverseRateLimit(sessionID string, id string) (int64, error) {
	hash := sha1.New()
	hash.Write([]byte(id))
	bs := hash.Sum(nil)
	return r.conn.Decr(ctx, fmt.Sprintf("requestLimit:%v:%v", sessionID, bs)).Result()
}

func (r *r) setSessionConfig(sessionID string, config config) error {
	parsedStr, _ := strconv.ParseInt(r.refreshTokenTTL, 10, 64)

	key := fmt.Sprintf("session:%v:config", sessionID)
	pipe := r.conn.TxPipeline()
	pipe.HSet(ctx, key, structs.Map(config)).Result()
	pipe.Expire(ctx, key, time.Duration(parsedStr)*time.Second)
	_, err := pipe.Exec(ctx)
	return err
}
