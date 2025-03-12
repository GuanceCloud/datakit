// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package offload offload data to other data sinks
package offload

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/avast/retry-go"
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
		httpCli: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   time.Second * 30,
					KeepAlive: time.Second * 90,
				}).DialContext,
				MaxIdleConns:          100,
				MaxConnsPerHost:       64,
				IdleConnTimeout:       time.Second * 90,
				TLSHandshakeTimeout:   time.Second * 10,
				ExpectContinueTimeout: time.Second,
			},
		},
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

const batchSize = 128

func (recevier *DKRecver) Send(s uint64, cat point.Category, data []*point.Point) error {
	if len(data) == 0 {
		return nil
	}

	i := s % (uint64)(len(recevier.AddrMap))
	addr := recevier.AddrMap[i][cat]

	if len(recevier.AddrMap) == 0 {
		return fmt.Errorf("no server address")
	}

	enc := point.GetEncoder(point.WithEncEncoding(point.LineProtocol),
		point.WithEncBatchSize(batchSize))
	defer point.PutEncoder(enc)

	dataList, err := enc.Encode(data)
	if err != nil {
		return err
	}

	if len(dataList) == 0 {
		return nil
	}

	for _, d := range dataList {
		buffer := bytes.NewReader(d)
		req, err := http.NewRequest("POST", addr, buffer)
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
	}

	return nil
}
