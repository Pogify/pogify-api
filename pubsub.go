package pogifyapi

import (
	"bytes"
	"net/http"
)

type pubsub struct {
	secret string
	url    string
}

func (p *pubsub) pub(ch chan<- *http.Response, errCh chan<- error, channel string, data []byte) {
	pub, err := http.NewRequest("POST", p.url+"/pub", bytes.NewReader(data))
	pub.Header.Add("authorization", p.secret)
	pubQ := pub.URL.Query()
	pubQ.Add("id", channel)
	pub.URL.RawQuery = pubQ.Encode()

	if err != nil {
		errCh <- err
		return
	}

	res, err := http.DefaultClient.Do(pub)
	if err != nil {
		errCh <- err
		return
	}

	ch <- res

}
