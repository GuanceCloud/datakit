// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Attributes struct {
	CommonAttributes map[string]string `json:"commonAttributes"`
}

type Record struct {
	Data []byte `json:"data"`
}

type requestData struct {
	RequestID string    `json:"requestId"`
	Timestamp int64     `json:"timestamp"`
	Records   []*Record `json:"records,omitempty"`
}

type firehoseResponseData struct {
	RequestID    string `json:"requestId"`
	Timestamp    int64  `json:"timestamp"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func respData(requestID string, errMsg string) []byte {
	r := &firehoseResponseData{RequestID: requestID}
	t := time.Now().UnixMilli()
	r.Timestamp = t
	if errMsg != "" {
		r.ErrorMessage = errMsg
	}
	bts, _ := json.Marshal(r)
	return bts
}

func resolveBodyEncoding(req *http.Request) string {
	if encoding := strings.TrimSpace(req.URL.Query().Get("encoding")); encoding != "" {
		return encoding
	}
	return req.Header.Get("Content-Encoding")
}

func gzipDecompress(compressedData []byte) ([]byte, error) {
	// 创建一个 reader 来读取压缩数据
	byteReader := bytes.NewReader(compressedData)

	// 创建 gzip.Reader
	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return nil, fmt.Errorf("new gzip reader error: %w", err)
	}
	defer gzipReader.Close() // nolint:errcheck

	// 读取解压后的数据
	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return decompressedData, nil
}

func getHeaders(header http.Header) map[string]string {
	headers := make(map[string]string)
	headers["request_id"] = header.Get("X-Amz-Firehose-Request-Id")
	headers["arn"] = header.Get("X-Amz-Firehose-Source-Arn")
	headers["firehose_protocol_version"] = header.Get("X-Amz-Firehose-Protocol-Version")

	commonAttr := header.Get("X-Amz-Firehose-Common-Attributes")
	if commonAttr != "" {
		attr := &Attributes{}
		if err := json.Unmarshal([]byte(commonAttr), attr); err == nil {
			for k, v := range attr.CommonAttributes {
				headers[k] = v
			}
		}
	}

	return headers
}
