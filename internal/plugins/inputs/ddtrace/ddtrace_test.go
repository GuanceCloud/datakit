// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && darwin

package ddtrace

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ugorji/go/codec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	msgpHandler codec.MsgpackHandle
	encoder     = codec.NewEncoder(nil, &msgpHandler)
	decoder     = codec.NewDecoder(nil, &msgpHandler)
)

func Marshal(src interface{}) ([]byte, error) {
	buf := bufpool.GetBuffer()
	encoder.Reset(buf)
	err := encoder.Encode(src)

	return buf.Bytes(), err
}

func Unmarshal(src io.Reader, dest interface{}) error {
	if src == nil || dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("invalid parameters for msgpack.Unmarshal")
	}

	decoder.Reset(src)

	return decoder.Decode(dest)
}

func msgpackEncoder(ddtraces DDTraces) ([]byte, error) {
	return Marshal(ddtraces)
}

func TestDDTraceAgent(t *testing.T) {
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktraces itrace.DatakitTraces, strikMod bool) {})

	rand.Seed(time.Now().UnixNano())
	// testJsonDDTraces(t)
	testMsgPackDDTraces(t)
}

func testJsonDDTraces(t *testing.T) {
	t.Helper()

	for _, version := range []string{v2, v3, v4} {
		tsvr := httptest.NewServer(handleDDTraceWithVersion(version))
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			buf, err := jsonEncoder(randomDDTraces(3, 10))
			if err != nil {
				t.Error(err.Error())

				return
			}

			req, err := http.NewRequest(method, tsvr.URL+version, bytes.NewBuffer(buf))
			if err != nil {
				t.Error(err.Error())

				return
			}

			for _, contentType := range []string{"application/json", "text/json"} {
				req.Header.Set("Content-Type", contentType)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Error(err.Error())

					return
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					fmt.Printf("request failed with status code %d\n", resp.StatusCode)
				}
			}
		}
	}
}

func testMsgPackDDTraces(t *testing.T) {
	t.Helper()

	for _, version := range []string{v3, v4} {
		tsvr := httptest.NewServer(handleDDTraceWithVersion(version))
		for _, method := range []string{http.MethodPost} {
			buf, err := msgpackEncoder(randomDDTraces(3, 10))
			// buf, err := randomDDTraces(3, 10).MarshalMsg(nil)
			if err != nil {
				t.Error(err.Error())

				return
			}

			req, err := http.NewRequest(method, tsvr.URL+version, bytes.NewBuffer(buf))
			if err != nil {
				t.Error(err.Error())

				return
			}

			req.Header.Set("Content-Type", "application/msgpack")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err.Error())

				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("request failed with status code %d\n", resp.StatusCode)
			}
		}
	}
}
