// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package offload offload data to other data sinks
package offload

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/avast/retry-go"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

const DKRcv = "datakit-http"

func sendReq(req *http.Request, cli *http.Client) (resp *http.Response, err error) {
	if err := retry.Do(
		func() error {
			resp, err = cli.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode/100 == 5 { // server-side error
				return fmt.Errorf("doSendReq: %s", resp.Status)
			}
			return nil
		},

		retry.Attempts(4),
		retry.Delay(time.Second*1),
		retry.OnRetry(func(n uint, err error) {
			l.Warnf("on %dth retry, error: %s", n, err)
		}),
	); err != nil {
		return resp, err
	}

	return resp, err
}

type DKRecver struct {
	httpCli *http.Client

	Addresses []string
	AddrMap   []map[point.Category]string

	sync.Mutex
}

func NewDKRecver(addresses []string) (*DKRecver, error) {
	receiver := &DKRecver{
		httpCli: ihttp.Cli(nil),
	}

	receiver.Addresses = append([]string{}, addresses...)
	receiver.AddrMap = make([]map[point.Category]string, len(addresses))

	allCat := point.AllCategories()

	for i, addr := range addresses {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, fmt.Errorf("parse url '%s' failed: %w", addr, err)
		}
		receiver.AddrMap[i] = map[point.Category]string{}
		for _, cat := range allCat {
			receiver.AddrMap[i][cat] = fmt.Sprintf("%s://%s%s",
				u.Scheme, u.Host, cat.URL())
		}
	}

	return receiver, nil
}

func (recevier *DKRecver) Send(s uint64, cat point.Category, data []*dkpt.Point) error {
	if len(data) == 0 {
		return nil
	}

	i := s % (uint64)(len(recevier.AddrMap))
	addr := recevier.AddrMap[i][cat]

	if len(recevier.AddrMap) == 0 {
		return fmt.Errorf("no server address")
	}

	dataStr := []string{}
	for _, pt := range data {
		if pt != nil {
			dataStr = append(dataStr, pt.String())
		}
	}
	reader := strings.NewReader(strings.Join(dataStr, "\n"))
	req, err := http.NewRequest("POST", addr, reader)
	if err != nil {
		return err
	}

	resp, err := sendReq(req, recevier.httpCli)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		r := make([]byte, 256)
		_, _ = resp.Body.Read(r)
		return fmt.Errorf("lastErrPostURL, http status code: %d, body: %s", resp.StatusCode, string(r))
	}

	return nil
}
