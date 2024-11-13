// Package exporter feed data to datakit
package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/stats"
)

var (
	defaultAPIServer        = "http://0.0.0.0:9529"
	defaultBPFTracingServer = "http://0.0.0.0:9529"
)

type target struct {
	t map[point.Category]string
}

func (t *target) URL(cat point.Category) string {
	return t.t[cat]
}

func buildBPFTracingTarge(bpftracingAPIServer string) *target {
	u, err := url.JoinPath(bpftracingAPIServer, "")
	if err != nil {
		log.Errorf("failed to join url `%s` and `%s`: %s",
			bpftracingAPIServer, "/v1/bpftracing", err)
	}
	return &target{
		t: map[point.Category]string{
			point.Tracing: u,
		},
	}
}

func buildTarget(apiServer string) *target {
	r := map[point.Category]string{}
	for _, cat := range point.AllCategories() {
		if u, err := url.JoinPath(apiServer, cat.URL()); err != nil {
			log.Errorf("failed to join url `%s` and `%s`: %s",
				apiServer, cat.URL(), err)
		} else {
			r[cat] = u
		}
	}
	return &target{
		t: r,
	}
}

const (
	dkLastErr = "/v1/lasterror"

	maxPtSendCount = 256
)

var log = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	log = nl
}

var (
	globalSender           *Sender
	globalBPFTracingSendor *Sender
)

type opt func(c *cfg)

type cfg struct {
	apiServer        string
	bpftracingServer string

	samplingRate          string
	samplingRatePtsPerMin string
}

func fixURL(u string) string {
	switch {
	case strings.HasPrefix(u, "http://"):
		return u
	case strings.HasPrefix(u, "https://"):
		return u
	default:
		return "http://" + u
	}
}

func WithSamplingRate(r string) opt {
	return func(c *cfg) {
		c.samplingRate = r
	}
}

func WithSamplingRatePtsPerMin(r string) opt {
	return func(c *cfg) {
		c.samplingRatePtsPerMin = r
	}
}

func WithBPFTracingServer(url string) opt {
	if url == "" {
		url = defaultBPFTracingServer
	}
	return func(c *cfg) {
		c.bpftracingServer = fixURL(url)
	}
}

func WithAPIServer(url string) opt {
	if url == "" {
		url = defaultAPIServer
	}
	return func(c *cfg) {
		c.apiServer = fixURL(url)
	}
}

var initOnce sync.Once

func Init(ctx context.Context, opts ...opt) {
	fn := func() {
		stats.MustRegister(ePtsVec)

		var c cfg
		for _, fn := range opts {
			fn(&c)
		}

		sampling := newSampling(ctx, &c)
		globalSender = NewSender(
			buildTarget(defaultAPIServer), sampling)
		globalBPFTracingSendor = NewSender(
			buildBPFTracingTarge(defaultBPFTracingServer), sampling)
	}
	initOnce.Do(fn)
}

type task struct {
	name string
	cat  point.Category

	data []*point.Point
}

type Sender struct {
	ch       chan *task
	httpCli  *http.Client
	target   *target
	sampling *sampling
}

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 16,
		IdleConnTimeout:     time.Second * 90,

		MaxConnsPerHost: 64,
	}
}

func NewSender(target *target, sampling *sampling) *Sender {
	sender := &Sender{
		ch: make(chan *task, 16),
		httpCli: &http.Client{
			Transport: newHTTPTransport(),
		},
		target:   target,
		sampling: sampling,
	}
	for i := 0; i < 4; i++ {
		go sender.runner(context.Background())
	}
	return sender
}

func (sender *Sender) runner(ctx context.Context) {
	for {
		select {
		case t := <-sender.ch:
			if err := sender.request(t); err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (sender *Sender) feed(name string, cat point.Category, data []*point.Point) error {
	catStr := cat.String()

	ePtsVec.WithLabelValues(name, catStr).Add(float64(len(data)))
	if globalSender == nil {
		return fmt.Errorf("sender not init")
	}

	// sampling
	if sender.sampling != nil {
		data = sender.sampling.sampling(catStr, data)
	}

	sender.ch <- &task{
		name: name,
		cat:  cat,
		data: data,
	}
	return nil
}

func (sender *Sender) request(data *task) error {
	if len(data.data) == 0 {
		return nil
	}

	if err := sender.doReq(data); err != nil {
		return fmt.Errorf("failed to send data: total %d pts: %w", len(data.data), err)
	}

	return nil
}

func (sender *Sender) doReq(task *task) error {
	if len(task.data) == 0 {
		return nil
	}

	if sender.httpCli == nil {
		return fmt.Errorf("no http client")
	}

	targetURL := sender.target.URL(task.cat)
	if targetURL == "" {
		return fmt.Errorf("unsupported category: %s", task.cat)
	}

	targetURL += "?input=" + url.QueryEscape(task.name)

	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf))
	defer point.PutEncoder(enc)

	batch := nBatch(task.data, maxPtSendCount)
	for i := range batch {
		buf, err := enc.Encode(batch[i])
		if err != nil {
			log.Warnf("failed to encode %d pts (batch %d): %s", len(batch[i]), i, err)
			continue
		}
		if err := sender.postData(buf[0], point.Protobuf, targetURL); err != nil {
			log.Warnf("failed to post %d pts (batch %d): %s", len(batch[i]), i, err)
			continue
		}
	}

	return nil
}

func (sender *Sender) postData(buf []byte, enc point.Encoding, url string) error {
	if sender == nil {
		return nil
	}

	reader := bytes.NewReader(buf)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Length", strconv.FormatInt(
		int64((len(buf))), 10))
	req.Header.Set("Content-Type", enc.HTTPContentType())

	resp, err := sender.httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("url %s, http status code: %d",
			url, resp.StatusCode)
	}
	return nil
}

func nBatch(pts []*point.Point, ptBatch int) (out [][]*point.Point) {
	for i := 0; i < len(pts)/ptBatch; i++ {
		out = append(out, pts[i*ptBatch:(i+1)*ptBatch])
	}
	if num := len(pts) % ptBatch; num > 0 {
		out = append(out, pts[len(pts)-num:])
	}
	return out
}

type ExternalLastErr struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

func FeedEBPFSpan(name string, cat point.Category, data []*point.Point) error {
	return globalBPFTracingSendor.feed(name, cat, data)
}

func FeedPoint(name string, cat point.Category, data []*point.Point) error {
	return globalSender.feed(name, cat, data)
}

func FeedLastError(extnlErr ExternalLastErr) error {
	lastErrURL, err := url.JoinPath(defaultAPIServer, dkLastErr)
	if err != nil {
		return fmt.Errorf("build url: %w", err)
	}
	client := http.Client{}
	data, err := json.Marshal(extnlErr)
	if err != nil {
		return err
	}
	rq, err := http.NewRequest("POST", lastErrURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	rsp, err := client.Do(rq)
	if err != nil {
		return err
	}
	defer rsp.Body.Close() //nolint:errcheck

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("lastErrPostURL, http status code: %d", rsp.StatusCode)
	}

	return nil
}
