package refertable

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckUrl(t *testing.T) {
	cases := []struct {
		url    string
		scheme string
		fail   bool
	}{
		{
			url:    "ht\tp//localss(",
			scheme: "http",
			fail:   true,
		},
		{
			url:    "httpS://localss(",
			scheme: "https",
			fail:   true,
		},
		{
			url:    "https://localhost/aa?a",
			scheme: "https",
		},
		{
			url:    "http://localhost/aa?a",
			scheme: "http",
		},
		{
			url:    "oss://localhost/aa?a",
			scheme: "oss",
			fail:   true,
		},
	}

	for _, v := range cases {
		scheme, err := checkURL(v.url)
		if err != nil {
			if !v.fail {
				t.Error(err)
			}
			continue
		}
		assert.Equal(t, v.scheme, scheme)
	}
}

func TestRunner(t *testing.T) {
	files := map[string]string{
		"a.json": testTableData,
	}
	server := newJsonDataServer(files)
	defer server.Close()

	url := server.URL

	plReferTable := &PlReferTables{}
	runner := &Runner{
		initFinished: make(chan struct{}),
	}

	ok := runner.InitFinished(time.Nanosecond)
	if ok {
		t.Errorf("func InitFinished exp: false, act %v", ok)
	}

	initReferTableRunner(runner, plReferTable, url+"?name=a.json", time.Second*5)

	ok = runner.InitFinished(time.Second)
	if !ok {
		t.Fatal("init not finishd")
	}

	v, ok := plReferTable.query("table1", []string{"key1"}, []any{"a"}, nil)
	if !ok || len(v) == 0 {
		t.Error("!ok")
	}

	stats := plReferTable.stats()
	assert.Equal(t, stats.Name[0], "table1")
	assert.Equal(t, stats.Name[1], "table2")
	assert.Equal(t, stats.Row[0], 3)
	assert.Equal(t, stats.Row[1], 2)
}

func newJsonDataServer(files map[string]string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
			default:
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			name := r.FormValue("name")
			data := files[name]
			w.Write([]byte(data))
			w.WriteHeader(http.StatusOK)
		},
	))
	return server
}
