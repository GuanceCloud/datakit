package prom_remote_write

import (
	"net"
	"net/http"
	"testing"
	"time"
)

func newHTTPListener() *Input {
	parser := Parser{}

	listener := &Input{
		Path:        "/receive",
		Methods:     []string{"POST"},
		Parser:      parser,
		MaxBodySize: 70000,
		DataSource:  "body",
		Tags:        map[string]string{"a": "b", "c": "d"},
	}
	// var filter = []string{"gc"}
	// listener.MetricNameFilter = filter
	listener.MeasurementPrefix = "hello_"
	listener.MeasurementName = "world"
	listener.TagsIgnore = []string{"a"}
	return listener
}

// start an HTTP server and then let prometheus writes data to the api
func TestWriteHTTP(t *testing.T) {
	listener := newHTTPListener()
	start(listener)
	time.Sleep(time.Hour)
}

func start(h *Input) error {
	if h.MaxBodySize == 0 {
		h.MaxBodySize = defaultMaxBodySize
	}

	server := &http.Server{
		Addr:    ":1234",
		Handler: h,
	}

	var listener net.Listener
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		return err
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			l.Errorf("Serve failed: %v", err)
		}
	}()

	l.Infof("Listening on %s", listener.Addr().String())

	return nil
}
