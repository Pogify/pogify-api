package v1

import (
	"encoding/hex"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func Test_r_newSession(t *testing.T) {
	m, err := miniredis.Run()
	defer m.Close()
	if err != nil {
		t.Fatal(err)
	}

	var r = new(r)
	r.conn = redis.NewClient(&redis.Options{
		Addr: m.Addr(),
	})
	r.refreshTokenTTL = "100"

	t.Run("Set a new session", func(t *testing.T) {
		key := "test1"
		testValue := "test1"

		ret, err := r.newSession(key, testValue)
		if err != nil {
			t.Fatalf("newSession errored with: %v", err)
			return
		}

		if ret != 1 {
			t.Error("newSession didn't return `1` on first call")
		}

		val, err := m.Get("session:test1")
		if err != nil {
			t.Fatalf("miniRedis errored with: %s", err)
		}

		if val != testValue {
			t.Fatalf("newSession set %s instead of %s", val, testValue)
			return
		}

		ttl := m.TTL("session:test1")

		if ttl != 100*time.Second {
			t.Error("newSession didn't set ttl")
		}
	})

	t.Run("test newSession on an existing session", func(t *testing.T) {
		key := "test1"
		testValue := "test2"
		m.FastForward(10 * time.Second)
		ret, err := r.newSession(key, testValue)
		if err != nil {
			t.Fatalf("newSession errored with: %v", err)
		}

		if ret != 0 {
			t.Error("newSession didn't return `0` on second call")
		}

		val, err := m.Get("session:test1")
		if err != nil {
			t.Errorf("miniRedis errored with: %s", err)
			return
		}
		if val == testValue {
			t.Errorf("newSession overwrote the value of %s to %s, but should be %s", key, testValue, "test1")
		}

		ttl := m.TTL("session:test1")

		if ttl != 90*time.Second {
			t.Error("newSession reset ttl")
		}
	})
}

func Test_r_verifyAndSetNewRefreshToken(t *testing.T) {
	m, err := miniredis.Run()
	defer m.Close()
	if err != nil {
		t.Fatalf("miniRedis errored with: %s", err)
		return
	}

	// new redis object
	var r = new(r)
	r.conn = redis.NewClient(&redis.Options{
		Addr: m.Addr(),
	})
	r.refreshTokenTTL = "10"

	session := "test1"
	token1 := "token1"
	token2 := "token2"

	val, err := r.verifyAndSetNewRefreshToken(session, token1, token2)
	if err != nil {
		t.Fatalf("verifyAndSetNewRefershToken errored with: %v", err)
	}
	if val != -1 {
		t.Fatal("verifyAndSetNewRefreshToken didn't return `-1` on a call to an expired key")
		return
	}

	// db setup
	r.newSession(session, token1)
	r.setSessionConfig(session, config{100})
	m.FastForward(5 * time.Second)

	res, err := r.verifyAndSetNewRefreshToken(session, token1, token2)

	if err != nil {
		t.Fatalf("verifyAndSetNewRefreshToken errored with: %s", err)
		return
	}

	if res != 1 {
		t.Error("verifyAndSetNewRefreshToken didn't return `1` on correct refresh token")
	}

	res, err = r.verifyAndSetNewRefreshToken(session, token1, token2)
	if err != nil {
		t.Fatalf("verifyAndSetNewRefreshToken errored with: %s", err)
		return
	}

	if res != 0 {
		t.Error("verifyAndSetNewRefreshToken didn't return `0` on incorrect refresh token")
	}

	newToken, err := m.Get("session:" + session)
	if err != nil {
		t.Fatalf("miniRedis errored: %v", err)
		return
	}

	if newToken != token2 {
		t.Errorf("verifyAndSetNewRefreshToken didn't return new refreshToken: %v but returned %v", token2, newToken)
	}

	sessionTTL := m.TTL("session:" + session)
	configTTL := m.TTL("session:" + session + ":config")
	ttl, _ := strconv.Atoi(r.refreshTokenTTL)

	if sessionTTL != time.Duration(ttl)*time.Second {
		t.Errorf("verifyAndSetNewRefreshToken didn't reset ttl for refresh token")
	}
	if configTTL != time.Duration(ttl)*time.Second {
		t.Errorf("verifyAndSetNewRefreshToken didn't reset ttl for config")
	}
}

func Test_r_rateLimitRequest(t *testing.T) {
	m, err := miniredis.Run()
	defer m.Close()
	if err != nil {
		t.Fatalf("miniRedis errored: %v", err)
		return
	}

	var r = new(r)
	r.conn = redis.NewClient(&redis.Options{
		Addr: m.Addr(),
	})
	r.refreshTokenTTL = "10"

	session := "test1"
	id := "id1"

	valS, err := r.rateLimitRequest(session, id)

	expectArr := [2]int64{1, 60}
	if valS != expectArr {
		t.Errorf("rateLimitRequest returned incorrect array. Got: %v Expected %v", valS, expectArr)
	}

	m.FlushAll()
	r.setSessionConfig(session, config{10})

	valS, err = r.rateLimitRequest(session, id)

	expectArr = [2]int64{1, 10}
	if valS != expectArr {
		t.Errorf("rateLimitRequest returned incorrect array. Got: %v Expected %v", valS, expectArr)
	}

	m.FastForward(5 * time.Second)
	valS, err = r.rateLimitRequest(session, id)

	expectArr = [2]int64{2, 5}
	if valS != expectArr {
		t.Errorf("rateLimitRequest returned incorrect array. Got: %v Expected %v", valS, expectArr)
	}

}

func Test_r_reverseRateLimit(t *testing.T) {
	m, err := miniredis.Run()
	defer m.Close()
	if err != nil {
		t.Fatalf("miniRedis errored: %v", err)
		return
	}

	var r = new(r)
	r.conn = redis.NewClient(&redis.Options{
		Addr: m.Addr(),
	})
	r.refreshTokenTTL = "10"

	dec, err := r.reverseRateLimit("test", "test")

	if dec != -1 {

	}
}

func Test_hashID(t *testing.T) {
	hash1, _ := hex.DecodeString("A9993E364706816ABA3E25717850C26C9CD0D89D")
	hash2, _ := hex.DecodeString("33C3E1456F82369F35BA05F71CBECC86A1BDD84C")
	hash3, _ := hex.DecodeString("3F0F29F9494B0446E5ECD67F08E995F4C5197079")

	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{"test hash 1", args{"abc"}, []byte(hash1)},
		{"test hash 1", args{"aasdfjlkklselknf"}, []byte(hash2)},
		{"test hash 1", args{"dfasdfeff2iu4yf98h9bf23hf-2134-_323"}, []byte(hash3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hashID(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hashID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_r_getSessionConfig(t *testing.T) {
	m, err := miniredis.Run()
	defer m.Close()
	if err != nil {
		t.Fatalf("miniRedis errored: %v", err)
		return
	}

	var r = new(r)
	r.conn = redis.NewClient(&redis.Options{
		Addr: m.Addr(),
	})
	r.refreshTokenTTL = "10"

	session := "test"

	nilConf, err := r.getSessionConfig(session)
	if err == nil {
		t.Error("getSessionConfig on no config didn't Error")
	}

	if nilConf != nil {
		t.Errorf("getSessionConfig didn't return `nil` on no conf; instead returned: %#v", nilConf)
	}

	setConf := config{100}
	r.setSessionConfig(session, setConf)

	gotConf, err := r.getSessionConfig(session)

	if !reflect.DeepEqual(setConf, *gotConf) {
		t.Errorf("getSessionConfig didn't return conf as was set; instead returned: %#v, expected: %#v", gotConf, setConf)
	}

}
