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
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const PlOffloadRcv = "ploffload"

type PloffloadReceiver struct {
	httpCli *http.Client

	Addresses []string
	AddrMap   []map[point.Category]string

	sync.Mutex
}

func NewPloffloadReceiver(addresses []string) (*PloffloadReceiver, error) {
	receiver := &PloffloadReceiver{
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
			return nil, fmt.Errorf("parse url '%s'  failed: %w", addr, err)
		}
		receiver.AddrMap[i] = map[point.Category]string{}
		for _, cat := range allCat {
			receiver.AddrMap[i][cat] = fmt.Sprintf("%s://%s/v1/write/ploffload/%s",
				u.Scheme, u.Host, cat.String())
		}
	}

	return receiver, nil
}

const plOffldDefaultEnc = point.Protobuf

func (recevier *PloffloadReceiver) Send(s uint64, cat point.Category, data []*point.Point) error {
	if len(data) == 0 {
		return nil
	}

	i := s % (uint64)(len(recevier.AddrMap))
	addr := recevier.AddrMap[i][cat]
	if len(recevier.AddrMap) == 0 {
		return fmt.Errorf("no server address")
	}

	enc := point.GetEncoder(point.WithEncEncoding(plOffldDefaultEnc),
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

		req.Header.Add("Content-Length", strconv.FormatInt(
			int64((len(d))), 10))
		req.Header.Add("Content-Type", plOffldDefaultEnc.HTTPContentType())

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
