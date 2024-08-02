// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type errMessage struct {
	Input      string `json:"input"`
	Source     string `json:"source"`
	ErrContent string `json:"err_content"`
}

type IAPIPutLastError interface {
	FeedLastError(err string, opts ...metrics.LastErrorOption)
}

// apiPutLastError 获取外部采集器错误，并 feederror 处理.
func apiPutLastError(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid API handle")
	}

	f, ok := args[0].(IAPIPutLastError)
	if !ok {
		return nil, fmt.Errorf("invalid feeder, got type %s", reflect.TypeOf(args[0]))
	}

	j, err := io.ReadAll(req.Body)
	if err != nil {
		l.Errorf("Read body error: %s", err.Error())
		return nil, err
	}
	defer req.Body.Close() //nolint:errcheck

	var em errMessage
	if err = json.Unmarshal(j, &em); err != nil {
		l.Errorf("json.Unmarshal: %s", err)
		return nil, fmt.Errorf("invalid JSON body: %w", err)
	}

	if em.Input == "" || em.ErrContent == "" {
		return nil, fmt.Errorf("input or errcontent can not be nil")
	}

	l.Debugf("error_name: %s, error_content:%s", em.Input, em.ErrContent)

	f.FeedLastError(em.ErrContent, metrics.WithLastErrorInput(em.Input), metrics.WithLastErrorSource(em.Source))

	return nil, nil
}
