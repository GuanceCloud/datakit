// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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

	defer resp.Body.Close() //nolint:errcheck
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
}
