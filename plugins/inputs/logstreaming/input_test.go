package logstreaming

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const body = `body`

func TestLogstreamingHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))

	var buf bytes.Buffer
	buf.WriteString(body)

	resp, err := http.Post(ts.URL, "test/plain", &buf)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
}
