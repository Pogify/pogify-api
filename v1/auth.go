package v1

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
)

type auth struct {
	googlePEM       map[string]*rsa.PublicKey
	googlePEMExpiry time.Time
	twitchKeys      map[string]*rsa.PublicKey
}

func (a *auth) getGooglePEM() map[string]*rsa.PublicKey {
	if len(a.googlePEM) > 0 && a.googlePEMExpiry.After(time.Now()) {
		return a.googlePEM
	}
	res, _ := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	body, _ := ioutil.ReadAll(res.Body)

	exp, _ := time.Parse(time.RFC1123, res.Header["Expires"][0])
	a.googlePEMExpiry = exp

	var r map[string]string
	json.Unmarshal(body, &r)

	rb := make(map[string]*rsa.PublicKey)

	for k, v := range r {
		pem, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(v))
		rb[k] = pem
	}
	a.googlePEM = rb
	return rb
}

func (a *auth) getTwitchKeys() map[string]*rsa.PublicKey {
	if len(a.twitchKeys) > 0 {
		return a.twitchKeys
	}
	keySet, _ := jwk.FetchHTTP("https://id.twitch.tv/oauth2/keys")

	rb := make(map[string]*rsa.PublicKey)

	for _, v := range keySet.Keys {
		n, _ := v.Materialize()

		rb[v.KeyID()] = n.(*rsa.PublicKey)
	}

	a.twitchKeys = rb
	return rb
}

func (a *auth) ValidateGoogleToken(t string) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {

		kid := a.getGooglePEM()[t.Header["kid"].(string)]
		if kid != nil {
			return a.getGooglePEM()[t.Header["kid"].(string)], nil
		}
		return nil, errors.New("token: kid does not exist")

	})
}

func (a *auth) ValidateTwitchToken(t string) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if !strings.Contains(t.Claims.(jwt.MapClaims)["iss"].(string), "id.twitch.tv") {
			return nil, errors.New("token: invalid iss")
		}

		kid := a.getTwitchKeys()[t.Header["kid"].(string)]
		if kid != nil {
			return a.getTwitchKeys()[t.Header["kid"].(string)], nil
		}
		return nil, errors.New("token: kid does not exist")

	})
}
