//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/hashicorp/go-retryablehttp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var DataKitAPIServer = "0.0.0.0:9529"

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
	data []*point.Point // lineproto
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
	if data == nil {
		return nil
	}

	l := len(data.data)

	for i := 0; i < l/maxPtSendCount; i++ {
		if err := sender.doReq(data.url, data.data[i*maxPtSendCount:(i+1)*maxPtSendCount]); err != nil {
			return fmt.Errorf("fail and stop: data[%d:%d]: %w", i*maxPtSendCount, (i+1)*maxPtSendCount, err)
		}
	}

	if l%maxPtSendCount != 0 {
		if err := sender.doReq(data.url, data.data[l-l%maxPtSendCount:]); err != nil {
			return fmt.Errorf("fail and stop: data[%d:]: %w", l-l%maxPtSendCount, err)
		}
	}
	return nil
}

func (sender *Sender) doReq(url string, data []*point.Point) error {
	if len(data) == 0 || url == "" {
		return nil
	}
	if sender.httpCli == nil {
		return fmt.Errorf("no http client")
	}
	dataStr := []string{}
	for _, pt := range data {
		if pt != nil {
			dataStr = append(dataStr, pt.String())
		}
	}
	reader := strings.NewReader(strings.Join(dataStr, "\n"))
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
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
		return fmt.Errorf("lastErrPostURL, http status code: %d", resp.StatusCode)
	}

	return nil
}

type ExternalLastErr struct {
	Input      string `json:"input"`
	ErrContent string `json:"err_content"`
}

func FeedMeasurement(url string, data []*point.Point) error {
	if _sender == nil {
		return fmt.Errorf("sender not init")
	}
	_sender.ch <- &task{
		url:  url,
		data: data,
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
