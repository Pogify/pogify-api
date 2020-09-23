package v1

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"reflect"
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
		redis.call("expire", KEYS[1]..":config", ARGV[3])
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
	local r = redis.call('hget', "session:" .. ARGV[1] .. ":config", "RequestInterval")
	if (c <= 1) then 
		if (r == false) then 
			redis.call('expire', KEYS[1], 60)
		else 
			redis.call('expire', KEYS[1], r) 
		end
	end 	
	return {c, redis.call('ttl', KEYS[1])}`

func (r *r) rateLimitRequest(sessionID string, id string) ([]int64, error) {
	hash := sha1.New()
	hash.Write([]byte(id))
	bs := hash.Sum(nil)
	key := fmt.Sprintf("requestLimit:%v:%x", sessionID, bs)
	log.Print(key)
	val, err := r.conn.Eval(ctx, requestLimitScript, []string{key}, sessionID).Result()
	log.Print(err)
	valS := make([]int64, 0, 2)
	log.Print(val)
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

func (r *r) getSessionConfig(sessionID string) (*config, error) {
	key := fmt.Sprintf("session:%v:config", sessionID)

	conf, err := r.conn.HGetAll(ctx, key).Result()

	if err != nil {
		return nil, err
	}

	if len(conf) == 0 {
		return nil, fmt.Errorf("No config for %s", sessionID)
	}
	fmt.Printf("%+v", conf)

	c := cast(&conf)

	return c, nil

}

func cast(conf *map[string]string) *config {
	var c config
	s := reflect.ValueOf(&c).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		switch f.Type().String() {
		case "int":
			i, _ := strconv.ParseInt((*conf)[typeOfT.Field(i).Name], 10, 64)
			f.SetInt(i)
		case "string":
			f.SetString((*conf)[typeOfT.Field(i).Name])
		}
	}

	return &c
}
