package prom_remote_write

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func newHTTPListener() *Input {
	parser := Parser{}

	listener := &Input{
		Path:        "/prom_remote_write",
		Methods:     []string{"POST"},
		Parser:      parser,
		MaxBodySize: 70000,
		DataSource:  "body",
	}
	filter := []string{"gc"}
	listener.MetricNameFilter = filter
	listener.MeasurementPrefix = "hello_"
	listener.MeasurementName = "world"
	listener.Tags = map[string]string{"a": "b", "c": "d"}
	listener.TagsIgnore = []string{"a"}
	return listener
}

// Start an HTTP server and then let prometheus writes data to the api.
func TestWriteHTTP(t *testing.T) {
	listener := newHTTPListener()
	start(listener)
	time.Sleep(time.Hour)
}

// Start prometheus remote write first before running this test.
func TestCollect(t *testing.T) {
	input := newHTTPListener()
	input.Output = "/prom_data"
	start(input)
	time.Sleep(5 * time.Second)
	for i := 0; i < 10000; i++ {
		fp := input.Output
		if !path.IsAbs(fp) {
			dir := datakit.InstallDir
			fp = filepath.Join(dir, fp)
		}
		file, err := os.Open(fp)
		if err != nil {
			t.Errorf("fail to open file")
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			t.Errorf("fail to read file")
		}
		measurements, err := input.Parse(data)
		if err != nil {
			t.Errorf("fail to parse data")
		}
		var points []*io.Point
		for _, m := range measurements {
			mm := m.(*Measurement)
			input.AddAndIgnoreTags(mm)
			p, err := mm.LineProto()
			if err != nil {
				t.Errorf("fail to get point")
			}
			fmt.Println(p)
			points = append(points, p)
		}
		time.Sleep(time.Millisecond * 500)
	}
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
