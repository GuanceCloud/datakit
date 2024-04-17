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
	AddrMap   []map[point.Category][2]string

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
	receiver.AddrMap = make([]map[point.Category][2]string, len(addresses))

	allCat := point.AllCategories()

	for i, addr := range addresses {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, fmt.Errorf("parse url '%s' failed: %w", addr, err)
		}
		receiver.AddrMap[i] = map[point.Category][2]string{}
		for _, cat := range allCat {
			receiver.AddrMap[i][cat] = [2]string{fmt.Sprintf("%s://%s%s",
				u.Scheme, u.Host, cat.URL()), u.Host}
		}
	}

	return receiver, nil
}

const batchSize = 128

func (recevier *DKRecver) Send(s uint64, cat point.Category, data []*point.Point) (err error) {
	tn := time.Now()
	if len(data) == 0 {
		return nil
	}

	i := s % (uint64)(len(recevier.AddrMap))
	addr := recevier.AddrMap[i][cat]

	totalPts := len(data)
	var batchN int
	defer func() {
		catStr := cat.String()
		if err != nil {
			if batchN > 0 {
				batchN -= 1
			}
			c := float64(totalPts - (batchN)*batchSize)
			ptOffloadErrorCountVec.WithLabelValues(catStr, DKRcv, addr[1]).Add(c)
			ptOffloadCountVec.WithLabelValues(catStr, DKRcv, addr[1]).Add(c)
		} else {
			ptOffloadCountVec.WithLabelValues(catStr, DKRcv, addr[1]).Add(float64(totalPts))
		}

		ptOffloadCostVec.WithLabelValues(catStr, DKRcv, addr[1]).Observe(
			float64(time.Since(tn)) / float64(time.Second))
	}()

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
		batchN++
		buffer := bytes.NewReader(d)
		req, err := http.NewRequest("POST", addr[0], buffer)
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
