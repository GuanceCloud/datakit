// Package feed feed data to datakit
package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/net/context/ctxhttp"
)

var DataKitAPIServer = "0.0.0.0:9529"

var dkLastErr = "/v1/lasterror"

type ExternalLastErr struct {
	Input      string `json:"input"`
	ErrContent string `json:"err_content"`
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

func WriteData(data []byte, urlPath string) error {
	// dataway path
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer timeoutCancel()

	resp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		return nil
	default:
		return fmt.Errorf("post to %s failed(HTTP: %d): %s", urlPath, resp.StatusCode, string(body))
	}
}

func FeedMeasurement(measurements []inputs.Measurement, path string) error {
	lines := [][]byte{}
	for _, m := range measurements {
		if m == nil {
			continue
		}
		if pt, err := m.LineProto(); err != nil {
			return err
		} else {
			ptstr := pt.String()
			lines = append(lines, []byte(ptstr))
		}
	}

	if err := WriteData(bytes.Join(lines, []byte("\n")), path); err != nil {
		return err
	}
	return nil
}
