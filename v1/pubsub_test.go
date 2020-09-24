package pogifyapi

import (
	"net/http"
	"os"
	"testing"

	gonanoid "github.com/matoous/go-nanoid"
)

func Test_pubsub_pub(t *testing.T) {
	p := new(pubsub)
	p.url = os.Getenv("PUBSUB_URL")

	ch := make(chan *http.Response)
	errCh := make(chan error)

	testChannel, _ := gonanoid.ID(10)
	testData, _ := gonanoid.ID(20)

	t.Run("missing auth header", func(t *testing.T) {
		go p.pub(ch, errCh, testChannel, []byte(testData))

		select {
		case res := <-ch:
			if res.StatusCode != http.StatusUnauthorized {
				t.Errorf("pub method returned %v not %v", res.StatusCode, http.StatusUnauthorized)
			}

		case err := <-errCh:
			t.Errorf("return error: %s", err)
		}
	})

	t.Run("missing auth header", func(t *testing.T) {
		p.secret = _pubsubsecret
		go p.pub(ch, errCh, testChannel, []byte(testData))

		select {
		case res := <-ch:
			if res.StatusCode != http.StatusOK {
				t.Errorf("pub method returned %v not %v", res.StatusCode, http.StatusOK)
			}
		case err := <-errCh:
			t.Errorf("return error: %s", err)
		}
	})
}
