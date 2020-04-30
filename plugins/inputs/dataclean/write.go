package dataclean

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/rotate"
)

const (
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = `text/plain; charset=utf-8`
	defaultMethod        = http.MethodPost
	defaultQueueCap      = 10000
)

type writer interface {
	write([]byte, *reqinfo) error
	close() error
}

type reqinfo struct {
	metrics []telegraf.Metric
	origUrl *url.URL
	headers http.Header
	queries url.Values
}

type httpWriter struct {
	client *http.Client
	schema string
	host   string //host:port
	path   string
	query  url.Values
	bgzip  bool
}

type fileWriter struct {
	Files               []string
	RotationInterval    time.Duration
	RotationMaxSize     int64
	RotationMaxArchives int

	writer  io.Writer
	closers []io.Closer
}

type writerMgr struct {
	serializer serializers.Serializer
	writers    []writer
	queues     []chan *reqinfo

	total     int64
	sendTotal int64

	running int32
	done    chan struct{}

	ctx       context.Context
	cancelFun context.CancelFunc
}

func newWritMgr() *writerMgr {
	w := &writerMgr{}
	w.serializer, _ = serializers.NewInfluxSerializer()
	w.done = make(chan struct{})
	w.ctx, w.cancelFun = context.WithCancel(context.Background())
	nc := 10
	w.running = int32(nc)
	for i := 0; i < nc; i++ {
		q := make(chan *reqinfo, defaultQueueCap)
		w.queues = append(w.queues, q)
	}

	return w
}

func (wm *writerMgr) addHttpWriter(remote string) {
	u, err := url.Parse(remote)
	if err != nil {
		log.Printf("E! [dataclean] invaid url=%s, err:%s", remote, err)
		return
	}
	w := newhttpWriter(u.Scheme, u.Host, u.Path, u.Query())
	wm.writers = append(wm.writers, w)
}

func (wm *writerMgr) addFileWriter(file string) {
	w, err := newFileWritfer([]string{file})
	if err == nil {
		wm.writers = append(wm.writers, w)
	}
}

func newhttpWriter(schema, host, path string, query url.Values) writer {
	w := &httpWriter{
		schema: schema,
		host:   host,
		path:   path,
		bgzip:  true,
		query:  query,
	}
	w.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: defaultClientTimeout,
	}
	return w
}

func newFileWritfer(files []string) (writer, error) {
	f := &fileWriter{
		Files: files,
	}

	writers := []io.Writer{}

	if len(f.Files) == 0 {
		f.Files = []string{"stdout"}
	}

	for _, file := range f.Files {
		if file == "stdout" {
			writers = append(writers, os.Stdout)
		} else {
			of, err := rotate.NewFileWriter(file, f.RotationInterval, f.RotationMaxSize, f.RotationMaxArchives)
			if err != nil {
				log.Printf("E! [dataclean] open file %s failed, %s", file, err)
				return nil, err
			}

			writers = append(writers, of)
			f.closers = append(f.closers, of)
		}
	}
	f.writer = io.MultiWriter(writers...)

	return f, nil
}

func (wm *writerMgr) add(req *reqinfo) {
	if len(wm.writers) == 0 {
		return
	}
	idx := uint64(wm.total) % uint64(len(wm.queues))
	select {
	case <-wm.ctx.Done():
		return
	case wm.queues[idx] <- req:
		wm.total++
	case <-time.After(time.Second * 5):
		log.Printf("W! [dataclean] queue too busy, total=%v", wm.total)
	}
}

func (wm *writerMgr) run() {

	log.Printf("[dataclean] write queue run")
	for i := 0; i < len(wm.queues); i++ {
		go wm.runQueue(wm.ctx, i, wm.queues[i])
	}
}

func (wm *writerMgr) stop() {
	wm.cancelFun()

	for _, q := range wm.queues {
		close(q)
	}

	select {
	case <-wm.done:
	case <-time.After(time.Second * 5):
		log.Printf("E! [dataclean] close queue time out")
	}

	for _, w := range wm.writers {
		w.close()
	}

	log.Printf("[dataclean] write done")
}

func (wm *writerMgr) runQueue(ctx context.Context, index int, queue chan *reqinfo) {

	defer func() {

		//log.Printf("D! [dataclean] write queue %d quit", index)

		if e := recover(); e != nil {
			log.Printf("E! [dataclean] write queue panic, %v", e)
		}

		if atomic.AddInt32(&wm.running, -1) == 0 {
			close(wm.done)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-queue:
			if !ok {
				return
			}

			log.Printf("D! [dataclean.write] Buffer fullness: %d / %d metrics", len(req.metrics), defaultQueueCap)

			body, err := wm.serializer.SerializeBatch(req.metrics)
			if err != nil {
				log.Printf("E! [dataclean] serialize metrics failed, %s", err)
				continue
			}
			for _, w := range wm.writers {
				if err := w.write(body, req); err != nil {
					log.Printf("E! [dataclean] write failed, %s", err)
				}
			}

		}
	}
}

func (w *httpWriter) close() error {
	return nil
}

func (w *httpWriter) write(data []byte, ri *reqinfo) error {

	var reqBodyBuffer io.Reader = bytes.NewBuffer(data)

	var err error
	if w.bgzip {
		rc, err := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

	u := ri.origUrl
	u.Scheme = w.schema
	u.Host = w.host
	u.Path = w.path

	//保留ftdataway的参数，同时带上url如果有特定参数(如果和ftdataway的参数重复，则ftdataway优先)
	query := make(url.Values)
	for k, v := range w.query {
		query[k] = v
	}

	for k, v := range ri.queries {
		if query.Get(k) == "" {
			query[k] = v
		}
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequest("POST", u.String(), reqBodyBuffer)
	if err != nil {
		return err
	}

	// if h.Username != "" || h.Password != "" {
	// 	req.SetBasicAuth(h.Username, h.Password)
	// }

	//req.Header.Set("User-Agent", "Telegraf/"+internal.Version())
	if ri.headers != nil {
		req.Header = ri.headers
	}
	//req.Header.Set("Content-Type", defaultContentType)
	if w.bgzip {
		req.Header.Set("Content-Encoding", "gzip")
	}
	// for k, v := range h.Headers {
	// 	if strings.ToLower(k) == "host" {
	// 		req.Host = v
	// 	}
	// 	req.Header.Set(k, v)
	// }

	log.Printf("D! [dataclean] final url: %s", u)
	log.Printf("D! [dataclean] final header: %s", req.Header)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to %s received status code: %d", u.String(), resp.StatusCode)
	}

	return nil
}

func (f *fileWriter) write(data []byte, _ *reqinfo) error {
	_, err := f.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileWriter) close() error {
	var err error
	for _, c := range f.closers {
		errClose := c.Close()
		if errClose != nil {
			err = errClose
		}
	}
	return err
}
