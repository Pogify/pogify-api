package v1

import (
	"reflect"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func Test_auth_getGooglePEM(t *testing.T) {
	a := new(auth)

	if len(a.googlePEM) > 0 {
		t.Fatal("googlePEM has items when it shouldn't")
	}

	pem := a.getGooglePEM()

	if len(pem) == 0 {
		t.Fatal("Empty PEM return")
	}

	for k, v := range pem {
		if k == "" {
			t.Fatalf("Empty key in PEM: %+v", pem)
		}
		if v == nil {
			t.Fatalf("<nil> value in PEM: %+v", pem)
		}
	}

	if !reflect.DeepEqual(pem, a.googlePEM) {
		t.Fatalf("Returned pem not cached")
	}

	// check that repeated call returns cache

	// add a key directly
	randKey := func() string {
		for k := range pem {
			return k
		}
		return ""
	}()

	a.googlePEM["test"] = a.googlePEM[randKey]

	pem2 := a.getGooglePEM()

	if pem2["test"] == nil {
		t.Fatal("getGooglePEM did not return cached on second call")
	}

}

func Test_auth_getTwitchKeys(t *testing.T) {
	a := new(auth)

	if len(a.twitchKeys) > 0 {
		t.Fatal("twitchKeys has items when it shouldn't")
	}

	keys := a.getTwitchKeys()

	if len(keys) == 0 {
		t.Fatal("Empty Keys return")
	}

	for k, v := range keys {
		if k == "" {
			t.Fatalf("Empty key in PEM: %+v", keys)
		}
		if v == nil {
			t.Fatalf("<nil> value in PEM: %+v", keys)
		}
	}

	if !reflect.DeepEqual(keys, a.twitchKeys) {
		t.Fatalf("Returned keys not cached")
	}

	// check that repeated call returns cache

	// add a key directly
	randKey := func() string {
		for k := range keys {
			return k
		}
		return ""
	}()

	a.twitchKeys["test"] = a.twitchKeys[randKey]

	keys2 := a.getTwitchKeys()

	if keys2["test"] == nil {
		t.Fatal("getTwitchKeys did not return cached on second call")
	}

}

func Test_auth_ValidateGoogleToken(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		a       *auth
		args    args
		want    *jwt.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.ValidateGoogleToken(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("auth.ValidateGoogleToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("auth.ValidateGoogleToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
