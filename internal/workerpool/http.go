// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package workerpool

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

type parameters struct {
	resp http.ResponseWriter
	req  *http.Request
	body *bytes.Buffer
}

func HTTPWrapper(statRespFunc ihttp.HTTPStatusResponse, wkp *WorkerPool, handler http.HandlerFunc) http.HandlerFunc {
	if wkp == nil || !wkp.enabled {
		return handler
	} else {
		return func(resp http.ResponseWriter, req *http.Request) {
			var (
				start           = time.Now()
				pbuf            = bufpool.GetBuffer()
				enterWorkerPool bool
				err             error
			)
			defer func() {
				if !enterWorkerPool {
					bufpool.PutBuffer(pbuf)
				}
			}()
			if _, err = io.Copy(pbuf, req.Body); err != nil {
				wkp.log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			wkp.log.Debugf("### [worker-pool wrapper] method: %s, url: %s, body-size: %d, read-body-cost: %dms",
				req.Method, req.URL.String(), pbuf.Len(), time.Since(start)/time.Millisecond)

			req.Body.Close() // nolint: errcheck,gosec
			req.Body = io.NopCloser(pbuf)
			param := &parameters{
				resp: &ihttp.NopResponseWriter{Raw: resp},
				req:  req,
				body: pbuf,
			}
			job, err := NewJob(WithInput(param),
				WithProcess(func(input interface{}) (output interface{}) {
					param := input.(*parameters)
					handler(param.resp, param.req)

					return nil
				}),
				WithProcessCallback(func(input, output interface{}, cost time.Duration) {
					param := input.(*parameters)
					bufpool.PutBuffer(param.body)
					wkp.log.Debugf("### job status: input: %#v, output: %#v, cost: %dms", input, output, cost/time.Millisecond)
				}),
			)
			if err != nil {
				wkp.log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			if err = wkp.MoreJob(job); err != nil {
				wkp.log.Error(err.Error())
				statRespFunc(resp, req, err)
			} else {
				enterWorkerPool = true
				wkp.log.Debug("HTTP wrapper: new job enters worker-pool success")
				statRespFunc(resp, req, nil)
			}
		}
	}
}
