package output

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/go-retryablehttp"
)

var DataKitAPIServer = "0.0.0.0:9529"

var DataKitTraceServer = ""

const (
	dkLastErr = "/v1/lasterror"

	maxPtSendCount = 512
)

var l = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
}

var _sender *Sender

func Init(logger *logger.Logger) {
	l = logger
	_sender = NewSender(logger)
}

type task struct {
	url  string
	data []*point.Point
	gzip bool
}

type Sender struct {
	ch      chan *task
	httpCli *retryablehttp.Client
}

func NewSender(l *logger.Logger) *Sender {
	retryCli := retryablehttp.NewClient()
	retryCli.RetryWaitMin = time.Second
	retryCli.RetryWaitMax = time.Second * 5
	retryCli.Logger = &retrycliLogger{}
	retryCli.RequestLogHook = retryCallback
	sender := &Sender{
		ch:      make(chan *task, 16),
		httpCli: retryCli,
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
				l.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
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

func gzipData(data []byte) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)

	if _, err := zw.Write(data); err != nil {
		return nil, err
	}

	if err := zw.Flush(); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return z.Bytes(), nil
}

func (sender *Sender) doReq(task *task) error {
	if len(task.data) == 0 || task.url == "" {
		return nil
	}
	if sender.httpCli == nil {
		return fmt.Errorf("no http client")
	}

	enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf),
		point.WithEncBatchSize(maxPtSendCount))
	defer point.PutEncoder(enc)

	bufLi, err := enc.Encode(task.data)
	if err != nil {
		return fmt.Errorf("encode data failed: %w", err)
	}

	for idx := range bufLi {
		if err := postData(bufLi[idx], point.Protobuf, task.gzip, task.url, sender); err != nil {
			l.Warnf("post data: %w", err)
		}
	}

	return nil
}

func postData(buf []byte, enc point.Encoding, gzip bool,
	url string, sender *Sender,
) error {
	if sender == nil {
		return nil
	}

	if gzip {
		if gzipBuf, err := gzipData(buf); err != nil {
			return fmt.Errorf("gzip data error: %w", err)
		} else {
			buf = gzipBuf
		}
	}

	reader := bytes.NewReader(buf)

	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(buf)))
	req.Header.Set("Content-Type", enc.HTTPContentType())
	if gzip {
		req.Header.Set("Content-Encoding", "gzip")
	}

	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return fmt.Errorf("retryablehttp.FromRequest: %w", err)
	}

	resp, err := sender.httpCli.Do(retryReq)
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

type ExternalLastErr struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

func FeedPoint(url string, data []*point.Point, gzip bool) error {
	if _sender == nil {
		return fmt.Errorf("sender not init")
	}
	_sender.ch <- &task{
		url:  url,
		data: data,
		gzip: gzip,
	}
	return nil
}

func FeedLastError(extnlErr ExternalLastErr) error {
	lastErrPostURL := fmt.Sprintf("http://%s%s", DataKitAPIServer, dkLastErr)
	client := http.Client{}
	data, err := json.Marshal(extnlErr)
	if err != nil {
		return err
	}
	rq, err := http.NewRequest("POST", lastErrPostURL, bytes.NewReader(data))
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
