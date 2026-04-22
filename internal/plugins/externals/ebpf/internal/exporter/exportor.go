// Package exporter feed data to datakit
package exporter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/stats"
)

var (
	defaultAPIServer        = "http://0.0.0.0:9529"
	defaultBPFTracingServer = "http://0.0.0.0:9529"

	globalSender *Sender
)

type target struct {
	t map[point.Category]string
}

func (t *target) URL(cat point.Category) string {
	return t.t[cat]
}

func buildTarget(apiServer string, traceServer string) *target {
	r := map[point.Category]string{}
	for _, cat := range point.AllCategories() {
		if cat == point.Tracing {
			if traceServer != "" {
				if u, err := url.JoinPath(traceServer, "/v1/bpftracing"); err != nil {
					log.Errorf("failed to join url `%s` and `%s`: %s",
						traceServer, cat.URL(), err)
				} else {
					r[cat] = u
				}
			}
			continue
		}
		if u, err := url.JoinPath(apiServer, cat.URL()); err != nil {
			log.Errorf("failed to join url `%s` and `%s`: %s",
				apiServer, cat.URL(), err)
		} else {
			r[cat] = u
		}
	}

	for cat, u := range r {
		log.Infof("category: %s, target api url: %s", cat, u)
	}

	return &target{
		t: r,
	}
}

const (
	dkLastErr = "/v1/lasterror"

	maxPtSendCount   = 256
	senderQueueSize  = 64
	senderWorkerSize = 4
)

var log = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	log = nl
}

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
		stats.MustRegister(eBPFMapEntriesVec)
		stats.MustRegister(eBPFMapMaxEntriesVec)
		stats.MustRegister(eBPFMapFillRatioVec)
		stats.MustRegister(eBPFMapCleanupVec)
		stats.MustRegister(eBPFEventDropVec)
		stats.MustRegister(eBPFMapObserveErrVec)
		stats.MustRegister(eAggEntriesVec)
		stats.MustRegister(eAggFlushPointsVec)
		stats.MustRegister(eAggFlushDurationVec)
		stats.MustRegister(eCacheEntriesVec)
		stats.MustRegister(eSenderQueueLen)
		stats.MustRegister(eSenderBatchPoints)
		stats.MustRegister(eSenderBatchBytes)
		stats.MustRegister(eSenderRequestTotal)
		stats.MustRegister(eSenderRequestDuration)
		stats.MustRegister(ePerfLostTotal)
		stats.MustRegister(ePerfReadErrorsTotal)
		stats.MustRegister(eTPacketStatsTotal)
		stats.MustRegister(eNICGroupCountVec)
		stats.MustRegister(eNICGroupRouteCountVec)
		stats.MustRegister(eAsyncQueueWaitDurationVec)
		stats.MustRegister(eAsyncProcessDurationVec)

		var c cfg
		for _, fn := range opts {
			fn(&c)
		}

		sampling := newSampling(ctx, &c)
		globalSender = NewSender(
			ctx, buildTarget(c.apiServer, c.bpftracingServer), sampling)
	}
	initOnce.Do(fn)
}

type task struct {
	targetURL string
	data      []*point.Point
}

type Sender struct {
	ctx      context.Context
	ch       chan *task
	httpCli  *http.Client
	target   *target
	sampling *sampling
}

type marshaler struct {
	pbpts point.PBPoints
	buf   []byte
}

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 16,
		IdleConnTimeout:     time.Second * 90,

		MaxConnsPerHost: 64,
	}
}

func NewSender(ctx context.Context, target *target, sampling *sampling) *Sender {
	if ctx == nil {
		ctx = context.Background()
	}

	sender := &Sender{
		ctx: ctx,
		ch:  make(chan *task, senderQueueSize),
		httpCli: &http.Client{
			Transport: newHTTPTransport(),
		},
		target:   target,
		sampling: sampling,
	}
	for i := 0; i < senderWorkerSize; i++ {
		go sender.runner(ctx)
	}
	return sender
}

func (sender *Sender) runner(ctx context.Context) {
	m := &marshaler{}
	var pending *task

	for {
		t, nextPending, ok := sender.nextTask(ctx, pending)
		if !ok {
			return
		}
		pending = nextPending
		if t == nil {
			continue
		}

		t, pending = sender.collectTask(t, pending)
		if err := sender.request(ctx, m, t); err != nil {
			log.Error(err)
		}
	}
}

func (sender *Sender) feed(name string, cat point.Category, data []*point.Point) error {
	if sender == nil {
		return fmt.Errorf("sender not init")
	}

	catStr := cat.String()

	ePtsVec.WithLabelValues(name, catStr).Add(float64(len(data)))

	// sampling
	if sender.sampling != nil {
		data = sender.sampling.sampling(catStr, data)
	}

	if len(data) == 0 {
		return nil
	}

	targetURL, err := sender.buildRequestURL(name, cat)
	if err != nil {
		return err
	}

	for start := 0; start < len(data); start += maxPtSendCount {
		end := start + maxPtSendCount
		if end > len(data) {
			end = len(data)
		}

		task := &task{
			targetURL: targetURL,
			data:      data[start:end:end],
		}

		if sender.ctx == nil {
			sender.ch <- task
			continue
		}

		select {
		case sender.ch <- task:
			ObserveSenderQueue(len(sender.ch))
		case <-sender.ctx.Done():
			return sender.ctx.Err()
		}
	}

	return nil
}

func (sender *Sender) nextTask(ctx context.Context, pending *task) (*task, *task, bool) {
	if pending != nil {
		return pending, nil, true
	}

	select {
	case t, ok := <-sender.ch:
		ObserveSenderQueue(len(sender.ch))
		return t, nil, ok
	case <-ctx.Done():
		return nil, nil, false
	}
}

func (sender *Sender) collectTask(current *task, pending *task) (*task, *task) {
	if current == nil {
		return nil, pending
	}

	if pending != nil {
		if merged, rest, ok := mergeTask(current, pending); ok {
			current = merged
			pending = rest
		} else {
			return current, pending
		}
	}

	for len(current.data) < maxPtSendCount {
		select {
		case next, ok := <-sender.ch:
			if !ok {
				return current, nil
			}
			if next == nil {
				continue
			}

			merged, rest, ok := mergeTask(current, next)
			if !ok {
				return current, next
			}
			current = merged
			pending = rest
			if pending != nil {
				return current, pending
			}
		default:
			return current, pending
		}
	}

	return current, pending
}

func mergeTask(current, next *task) (*task, *task, bool) {
	switch {
	case current == nil:
		return next, nil, true
	case next == nil:
		return current, nil, true
	case current.targetURL != next.targetURL:
		return current, nil, false
	}

	if len(current.data) >= maxPtSendCount || len(next.data) == 0 {
		return current, nil, true
	}

	remain := maxPtSendCount - len(current.data)
	if remain <= 0 {
		return current, next, true
	}

	if remain >= len(next.data) {
		current.data = append(current.data, next.data...)
		return current, nil, true
	}

	current.data = append(current.data, next.data[:remain]...)
	next.data = next.data[remain:]
	return current, next, true
}

func (sender *Sender) buildRequestURL(name string, cat point.Category) (string, error) {
	if sender == nil {
		return "", fmt.Errorf("sender not init")
	}
	if sender.target == nil {
		return "", fmt.Errorf("no target")
	}

	targetURL := sender.target.URL(cat)
	if targetURL == "" {
		return "", fmt.Errorf("unsupported category: %s", cat)
	}

	return targetURL + "?input=" + url.QueryEscape(name), nil
}

func (sender *Sender) request(ctx context.Context, m *marshaler, data *task) error {
	if data == nil || len(data.data) == 0 {
		return nil
	}
	if sender == nil {
		return fmt.Errorf("sender not init")
	}
	if m == nil {
		return fmt.Errorf("no marshaler")
	}
	if sender.httpCli == nil {
		return fmt.Errorf("no http client")
	}

	payload, err := m.marshal(data.data)
	if err != nil {
		return fmt.Errorf("encode %d pts: %w", len(data.data), err)
	}
	ObserveSenderBatch(len(data.data), len(payload))

	if _, err := sender.postData(ctx, payload, point.Protobuf, data.targetURL); err != nil {
		return fmt.Errorf("post %d pts: %w", len(data.data), err)
	}

	return nil
}

func (m *marshaler) marshal(pts []*point.Point) ([]byte, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	m.pbpts.Arr = m.pbpts.Arr[:0]
	for i := range pts {
		if pts[i] == nil {
			continue
		}
		m.pbpts.Arr = append(m.pbpts.Arr, pts[i].PBPoint())
	}

	size := m.pbpts.Size()
	if size == 0 {
		return nil, nil
	}

	if cap(m.buf) < size {
		m.buf = make([]byte, size)
	} else {
		m.buf = m.buf[:size]
	}

	n, err := m.pbpts.MarshalToSizedBuffer(m.buf)
	if err != nil {
		return nil, err
	}
	return m.buf[:n], nil
}

func (sender *Sender) postData(ctx context.Context, buf []byte, enc point.Encoding, url string) (string, error) {
	if sender == nil {
		return "sender_not_init", fmt.Errorf("sender not init")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	reader := bytes.NewReader(buf)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
	if err != nil {
		return "new_request_error", err
	}

	req.Header.Set("Content-Length", strconv.FormatInt(
		int64((len(buf))), 10))
	req.Header.Set("Content-Type", enc.HTTPContentType())

	start := time.Now()
	resp, err := sender.httpCli.Do(req)
	if err != nil {
		ObserveSenderRequest("request_error", time.Since(start))
		return "request_error", err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		result := "http_error"
		switch resp.StatusCode / 100 {
		case 4:
			result = "http_4xx"
		case 5:
			result = "http_5xx"
		}
		ObserveSenderRequest(result, time.Since(start))
		return result, fmt.Errorf("url %s, http status code: %d",
			url, resp.StatusCode)
	}
	ObserveSenderRequest("ok", time.Since(start))
	return "ok", nil
}

type ExternalLastErr struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

func FeedEBPFSpan(name string, cat point.Category, data []*point.Point) error {
	return globalSender.feed(name, cat, data)
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
