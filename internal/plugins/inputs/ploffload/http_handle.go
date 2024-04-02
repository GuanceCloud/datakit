// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ploffload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
)

type ptsData struct {
	inputName string
	cat       point.Category
	pts       []*point.Point
}

func (ipt *Input) handlePlOffload(resp http.ResponseWriter, req *http.Request) {
	ptsData, err := getRequestData(req)
	if err != nil {
		log.Error(err.Error())
		httpStatusRespFunc(resp, req, err)
		return
	}

	log.Debugf("category: %s, pts num: %d", ptsData.cat, len(ptsData.pts))
	if err := ipt.feeder.Feed(ptsData.inputName, ptsData.cat, ptsData.pts, nil); err != nil {
		log.Error(err.Error())
		httpStatusRespFunc(resp, req, err)
		return
	}

	httpStatusRespFunc(resp, req, nil)
}

// getRequestData get data from request
//
// this func modified from internal/httpapi/api_write.go/apiWrite.
func getRequestData(req *http.Request) (*ptsData, error) {
	var err error

	input := inputName
	urlPath := req.URL.Path

	var categoryStr string
	if idx := strings.LastIndex(urlPath, "/"); idx >= 0 {
		categoryStr = strings.ToLower(urlPath[idx+1:])
	} else {
		categoryStr = point.SLogging
	}

	opts := []point.Option{
		point.WithPrecision(point.NS),
		point.WithTime(time.Now()),
	}

	switch categoryStr {
	// set specific options on them
	case point.SMetric:
		opts = append(opts, point.DefaultMetricOptions()...)
	case point.SLogging:
		opts = append(opts, point.DefaultLoggingOptions()...)
	case point.SObject:
		opts = append(opts, point.DefaultObjectOptions()...)

	case point.SNetwork,
		point.STracing,
		point.SKeyEvent: // pass

		// set input-name for them
	case point.SCustomObject:
		input = "custom_object"
	case point.SSecurity:
		input = "scheck"

	default:
		log.Debugf("invalid category: %q", categoryStr)
		return nil, uhttp.Errorf(httpapi.ErrInvalidCategory, "invalid URL %q", urlPath)
	}

	q := req.URL.Query()

	if x := q.Get(httpapi.ArgInput); x != "" {
		input = x
	}

	if x := q.Get(httpapi.ArgPrecision); x != "" {
		opts = append(opts, point.WithPrecision(point.PrecStr(x)))
	}

	body := bufpool.GetBuffer()
	defer bufpool.PutBuffer(body)

	// _, err = io.Copy(body, req.Body)
	// if err != nil {
	// return nil, err
	// }

	_, err = io.Copy(body, req.Body)
	if err != nil {
		return nil, err
	}

	bodyBuf := body.Bytes()

	if len(bodyBuf) == 0 {
		return nil, httpapi.ErrEmptyBody
	}

	cntTyp := httpapi.GetPointEncoding(req.Header)

	pts, err := httpapi.HandleWriteBody(body.Bytes(), cntTyp, opts...)
	if err != nil {
		if errors.Is(err, point.ErrInvalidLineProtocol) {
			return nil, uhttp.Errorf(httpapi.ErrInvalidLinePoint, "%s: body(%d bytes)", err.Error(), len(bodyBuf))
		}
		return nil, err
	}

	if len(pts) == 0 {
		return nil, httpapi.ErrNoPoints
	}

	// add extra tags
	ignoreGlobalTags := false
	for _, arg := range []string{
		httpapi.ArgIgnoreGlobalHostTags,
		httpapi.ArgIgnoreGlobalTags, // deprecated
	} {
		if x := q.Get(arg); x != "" {
			ignoreGlobalTags = true
		}
	}

	if !ignoreGlobalTags {
		appendTags(pts, datakit.GlobalHostTags())
	}

	if x := q.Get(httpapi.ArgGlobalElectionTags); x != "" {
		appendTags(pts, datakit.GlobalElectionTags())
	}

	log.Debugf("received %d(%s) points from %s",
		len(pts), urlPath, input)

	// under strict mode, any point warning are errors
	strict := false
	if q.Get(httpapi.ArgStrict) != "" {
		strict = true
	}
	if strict {
		for _, pt := range pts {
			if arr := pt.Warns(); len(arr) > 0 {
				switch cntTyp {
				case point.JSON:
					return nil, uhttp.Errorf(httpapi.ErrInvalidJSONPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				case point.LineProtocol:
					return nil, uhttp.Errorf(httpapi.ErrInvalidLinePoint, "%s: %s", arr[0].Type, arr[0].Msg)
				case point.Protobuf:
					return nil, uhttp.Errorf(httpapi.ErrInvalidProtobufPoint, "%s: %s", arr[0].Type, arr[0].Msg)
				}
			}
		}
	}

	pData := &ptsData{
		inputName: input,
		cat:       point.CatString(categoryStr),
		pts:       pts,
	}

	return pData, nil
}

func appendTags(pts []*point.Point, tags map[string]string) {
	for k, v := range tags {
		for _, pt := range pts {
			pt.AddTag(k, v)
		}
	}
}

type workerPool struct {
	jobLen int
	job    chan *storage.Request

	fn http.HandlerFunc

	semStop *cliutils.Sem
}

func (w *workerPool) Start() {
	if w.jobLen <= 0 {
		w.jobLen = 16
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "ploffload_worker_pool"})

	for i := 0; i < w.jobLen; i++ {
		g.Go(func(ctx context.Context) error {
			for {
				select {
				case reqpb := <-w.job:
					req := convStorageReq2HTTPReq(reqpb)
					if w.fn != nil {
						w.fn(&httpapi.NopResponseWriter{}, req)
					}
				case <-w.semStop.Wait():
					return nil
				}
			}
		})
	}
}

func (w *workerPool) Stop() {
	w.semStop.Close()
}

func (w *workerPool) FeedChan() chan<- *storage.Request {
	return w.job
}

func NewWorkerPool(jobLen int, handle http.HandlerFunc) *workerPool {
	return &workerPool{
		jobLen:  jobLen,
		job:     make(chan *storage.Request, jobLen),
		fn:      handle,
		semStop: cliutils.NewSem(),
	}
}

func HTTPStorageWrapperWithWkrPool(key uint8,
	statRespFunc httpapi.HTTPStatusResponse,
	s *storage.Storage,
	wkrPool *workerPool,
	handler http.HandlerFunc,
) http.HandlerFunc {
	if s == nil || !s.Enabled() {
		return handler
	} else {
		return func(resp http.ResponseWriter, req *http.Request) {
			pbuf := bufpool.GetBuffer()
			defer bufpool.PutBuffer(pbuf)

			_, err := io.Copy(pbuf, req.Body)
			if err != nil {
				log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			reqpb := &storage.Request{
				Method:           req.Method,
				Url:              req.URL.String(),
				Proto:            req.Proto,
				ProtoMajor:       int32(req.ProtoMajor),
				ProtoMinor:       int32(req.ProtoMinor),
				Header:           storage.ConvertMapToMapEntries(req.Header),
				Body:             pbuf.Bytes(),
				ContentLength:    req.ContentLength,
				TransferEncoding: req.TransferEncoding,
				Close:            req.Close,
				Host:             req.Host,
				Form:             storage.ConvertMapToMapEntries(req.Form),
				PostForm:         storage.ConvertMapToMapEntries(req.PostForm),
				RemoteAddr:       req.RemoteAddr,
				RequestUri:       req.RequestURI,
			}

			buf, err := proto.Marshal(reqpb)
			if err != nil {
				log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			if wkrPool != nil {
				select {
				case wkrPool.FeedChan() <- reqpb:
					// 线程池优先，若队列满则进入磁盘缓存阶段
					// 尽可能的减少磁盘读写
					return
				default:
				}
			}

			if err = s.Put(key, buf); err != nil {
				log.Error(err.Error())
				statRespFunc(resp, req, err)
			} else {
				log.Debug("HTTP wrapper: new data put into local-cache success")
				statRespFunc(resp, req, nil)
			}
		}
	}
}

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		var (
			e1 *uhttp.HttpError
			e2 *uhttp.MsgError
		)
		switch {
		case errors.As(err, &e1):
			resp.WriteHeader(e1.HttpCode)
			if v := jsonErrFmt(e1, ""); len(v) > 0 {
				_, _ = resp.Write(v)
			}
			return
		case errors.As(err, &e2):
			resp.WriteHeader(e2.HttpError.HttpCode)
			if v := jsonErrFmt(e2.HttpError, e2.Fmt, e2.Args); len(v) > 0 {
				_, _ = resp.Write(v)
			}

			return
		default:
			resp.WriteHeader(http.StatusInternalServerError)
			if v := jsonErrFmt(uhttp.NewErr(err,
				http.StatusInternalServerError), ""); len(v) > 0 {
				_, _ = resp.Write(v)
			}
			return
		}
	} else {
		resp.WriteHeader(200)
	}
}

func jsonErrFmt(err *uhttp.HttpError, fmtStr string, args ...any) []byte {
	var errStr string
	if fmtStr != "" {
		if len(args) > 0 {
			errStr = fmt.Sprintf(fmtStr, args...)
		} else {
			errStr = fmtStr
		}
	}

	resp := &uhttp.BodyResp{
		HttpError: err,
		Message:   errStr,
	}

	buf, _ := json.Marshal(resp)
	return buf
}

func convStorageReq2HTTPReq(reqpb *storage.Request) *http.Request {
	req := &http.Request{
		Method:           reqpb.Method,
		Proto:            reqpb.Proto,
		ProtoMajor:       int(reqpb.ProtoMajor),
		ProtoMinor:       int(reqpb.ProtoMinor),
		Header:           storage.ConvertMapEntriesToMap(reqpb.Header),
		Body:             io.NopCloser(bytes.NewBuffer(reqpb.Body)),
		ContentLength:    reqpb.ContentLength,
		TransferEncoding: reqpb.TransferEncoding,
		Close:            reqpb.Close,
		Host:             reqpb.Host,
		Form:             storage.ConvertMapEntriesToMap(reqpb.Form),
		PostForm:         storage.ConvertMapEntriesToMap(reqpb.PostForm),
		RemoteAddr:       reqpb.RemoteAddr,
		RequestURI:       reqpb.RequestUri,
	}
	var err error
	if req.URL, err = url.Parse(reqpb.Url); err != nil {
		log.Errorf("### parse raw URL: %s failed: %s", reqpb.Url, err.Error())
	}

	return req
}
