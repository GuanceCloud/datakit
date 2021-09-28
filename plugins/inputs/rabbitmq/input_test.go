package rabbitmq

import (

	//"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/nodes":
			nodeHandle(w, r)

		case "/api/exchanges":
			exchangeHandle(w, r)
		case "/api/overview":
			overviewHandle(w, r)
		case "/api/queues":
			queueHandle(w, r)
		case "/api/queues/{vhost}/{name}/bindings":
			bindingHandle(w, r)
		default:
			t.Errorf("unexpected url: %s", r.URL.Path)
		}
	}))

	defer ts.Close()

	n := &Input{
		Url: ts.URL,
	}
	cli, err := n.createHttpClient()
	if err != nil {
		l.Fatal(err)
	}
	n.client = cli

	n.getMetric()
	for _, v := range collectCache {
		s, err := v.LineProto()
		if err != nil {
			t.Error(err)
		}
		t.Logf(s.String())
	}
}

func nodeHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(nodeHandleData))
}

func queueHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(queueHandleData))
}

func overviewHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(overviewHandleData))
}

func exchangeHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(exchangeHandleData))
}

func bindingHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(bindingHandleData))
}
