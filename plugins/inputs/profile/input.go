// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package profile datakit collector
package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName            = "profile"
	profileMaxSize       = 1 << 23
	ProxySaveErrorHeader = "X-Proxy-Error"
	sampleConfig         = `
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  ## set true to enable election, pull mode only
  election = true

## go pprof config
## collect profiling data in pull mode
#[[inputs.profile.go]]
  ## pprof url
  #url = "http://localhost:6060"

  ## pull interval, should be greater or equal than 10s
  #interval = "10s"

  ## service name
  #service = "go-demo"

  ## app env
  #env = "dev"

  ## app version
  #version = "0.0.0"

  ## types to pull
  ## values: cpu, goroutine, heap, mutex, block
  #enabled_types = ["cpu","goroutine","heap","mutex","block"]

#[inputs.profile.go.tags]
  # tag1 = xxxxx
`
)

var (
	log = logger.DefaultSLogger(inputName)

	_ inputs.HTTPInput     = &Input{}
	_ inputs.InputV2       = &Input{}
	_ inputs.ElectionInput = (*Input)(nil)

	workSpaceUUID         string
	workSpaceUUIDInitLock sync.Mutex
	// A Regexp is concurrent safe, so we can define this var globally.
	workSpaceUUIDRegexp = regexp.MustCompile(`ws_uuid"\s*:\s*"([^"]*?)"`)

	pointCache     *profileCache
	pointCacheOnce sync.Once
)

func InitCache() {
	pointCacheOnce.Do(func() {
		pointCache = newProfileCache(32, 4096)
	})
}

type profileCache struct {
	pointsMap map[string]*profileBase
	heap      *minHeap
	maxSize   int
	lock      *sync.Mutex // lock: map and heap access lock
}

type minHeap struct {
	buckets []*profileBase
	indexes map[*profileBase]int
}

func newMinHeap(initCap int) *minHeap {
	return &minHeap{
		buckets: make([]*profileBase, 0, initCap),
		indexes: make(map[*profileBase]int, initCap),
	}
}

func (mh *minHeap) Swap(i, j int) {
	mh.indexes[mh.buckets[i]], mh.indexes[mh.buckets[j]] = j, i
	mh.buckets[i], mh.buckets[j] = mh.buckets[j], mh.buckets[i]
}

func (mh *minHeap) Less(i, j int) bool {
	return mh.buckets[i].birth.Before(mh.buckets[j].birth)
}

func (mh *minHeap) Len() int {
	return len(mh.buckets)
}

func (mh *minHeap) siftUp(idx int) {
	if idx >= len(mh.buckets) {
		errMsg := fmt.Sprintf("siftUp: index[%d] out of bounds[%d]", idx, len(mh.buckets))
		log.Error(errMsg)
		panic(errMsg)
	}

	for idx > 0 {
		parent := (idx - 1) / 2

		if !mh.Less(idx, parent) {
			break
		}

		// Swap
		mh.Swap(idx, parent)
		idx = parent
	}
}

func (mh *minHeap) siftDown(idx int) {
	for {
		left := idx*2 + 1
		if left >= mh.Len() {
			break
		}

		minIdx := idx
		if mh.Less(left, minIdx) {
			minIdx = left
		}

		right := left + 1
		if right < mh.Len() && mh.Less(right, minIdx) {
			minIdx = right
		}

		if minIdx == idx {
			break
		}

		mh.Swap(idx, minIdx)
		idx = minIdx
	}
}

func (mh *minHeap) push(pb *profileBase) {
	mh.buckets = append(mh.buckets, pb)
	mh.indexes[pb] = mh.Len() - 1
	mh.siftUp(mh.Len() - 1)
}

func (mh *minHeap) pop() *profileBase {
	if mh.Len() == 0 {
		return nil
	}

	top := mh.getTop()
	mh.remove(top)
	return top
}

func (mh *minHeap) remove(pb *profileBase) {
	idx, ok := mh.indexes[pb]
	if !ok {
		log.Errorf("pb not found in the indexes, profileID = %s", pb.profileID)
		return
	}
	if idx >= mh.Len() {
		errMsg := fmt.Sprintf("remove: index[%d] out of bounds [%d]", idx, mh.Len())
		log.Error(errMsg)
		panic(errMsg)
	}

	if mh.buckets[idx] != pb {
		errMsg := fmt.Sprintf("remove: idx of the buckets[%p] not equal the removing target[%p]", mh.buckets[idx], pb)
		log.Error(errMsg)
		panic(errMsg)
	}
	// delete the idx
	mh.buckets[idx] = mh.buckets[mh.Len()-1]
	mh.indexes[mh.buckets[idx]] = idx
	mh.buckets = mh.buckets[:mh.Len()-1]

	if idx < mh.Len() {
		mh.siftDown(idx)
	}
	delete(mh.indexes, pb)
}

func (mh *minHeap) getTop() *profileBase {
	if mh.Len() == 0 {
		return nil
	}
	return mh.buckets[0]
}

type profileBase struct {
	profileID string
	birth     time.Time
	point     *point.Point
}

func newProfileCache(initCap int, maxCap int) *profileCache {
	if initCap < 32 {
		initCap = 32
	} else if initCap > 256 {
		initCap = 256
	}

	if maxCap < initCap {
		maxCap = initCap
	} else if maxCap > 8196 {
		maxCap = 8196
	}

	return &profileCache{
		pointsMap: make(map[string]*profileBase, initCap),
		heap:      newMinHeap(initCap),
		maxSize:   maxCap,
		lock:      &sync.Mutex{},
	}
}

func (pc *profileCache) push(profileID string, birth time.Time, point *point.Point) {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	if pc.heap.Len() >= pc.maxSize {
		pb := pc.heap.pop()
		if pb != nil {
			delete(pc.pointsMap, pb.profileID)

			log.Warnf("由于达到cache存储数量上限，最早的point数据被丢弃，profileID = [%s], profileTime = [%s]",
				pb.profileID, pb.birth.Format(time.RFC3339))
		}
	}

	newPB := &profileBase{
		profileID: profileID,
		birth:     birth,
		point:     point,
	}

	pc.pointsMap[profileID] = newPB
	pc.heap.push(newPB)
}

func (pc *profileCache) drop(profileID string) *point.Point {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	if pb, ok := pc.pointsMap[profileID]; ok {
		delete(pc.pointsMap, profileID)
		pc.heap.remove(pb)

		if len(pc.pointsMap) != pc.heap.Len() {
			log.Warnf("cache map size do not equals heap size, map size = [%d], heap size = [%d]",
				len(pc.pointsMap), pc.heap.Len())
		}
		return pb.point
	}
	return nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
			Election: true,
			semStop:  cliutils.NewSem(),
		}
	})
}

type Input struct {
	Endpoints []string      `toml:"endpoints"`
	Go        []*GoProfiler `toml:"go"`

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

// uploadResponse {"content":{"profileID":"fa9c3d16-1cfc-4e37-950d-129cbebd1cdb"}}.
type uploadResponse struct {
	Content *struct {
		ProfileID string `json:"profileID"`
	} `json:"content"`
}

func queryWorkSpaceUUID() (string, error) {
	if workSpaceUUID != "" {
		return workSpaceUUID, nil
	}

	workSpaceUUIDInitLock.Lock()
	defer workSpaceUUIDInitLock.Unlock()

	if workSpaceUUID != "" {
		return workSpaceUUID, nil
	}

	tokens := config.Cfg.DataWay.GetTokens()
	if len(tokens) == 0 {
		return "", fmt.Errorf("dataway token missing")
	}
	ws := dkhttp.Workspace{Token: tokens}
	wsJSON, err := json.Marshal(ws)
	if err != nil {
		return "", fmt.Errorf("json marshal fail: %w", err)
	}
	resp, err := config.Cfg.DataWay.WorkspaceQuery(wsJSON)
	if err != nil {
		return "", fmt.Errorf("workspace query fail: %w, query body: %s", err, string(wsJSON))
	}
	// for lint:errCheck
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body fail:%w", err)
	}

	matches := workSpaceUUIDRegexp.FindSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("no match for express[%s] found, body [%s]", workSpaceUUIDRegexp.String(), string(body))
	}
	workSpaceUUID = string(matches[1])
	return workSpaceUUID, nil
}

func profilingProxyURL() (*url.URL, error) {
	lastErr := fmt.Errorf("no dataway endpoint available now")

	endpoints := config.Cfg.DataWay.GetAvailableEndpoints()

	if len(endpoints) == 0 {
		return nil, lastErr
	}

	for _, ep := range endpoints {
		rawURL, ok := ep.GetCategoryURL()[datakit.ProfilingUpload]
		if !ok || rawURL == "" {
			lastErr = fmt.Errorf("profiling upload url empty")
			continue
		}

		URL, err := url.Parse(rawURL)
		if err != nil {
			lastErr = fmt.Errorf("profiling upload url [%s] parse err:%w", rawURL, err)
			continue
		}
		return URL, nil
	}
	return nil, lastErr
}

// RegHTTPHandler simply proxy profiling request to dataway.
func (i *Input) RegHTTPHandler() {
	URL, err := profilingProxyURL()
	if err != nil {
		log.Errorf("no profiling proxy url available: %s", err)
		return
	}

	InitCache()

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// not a post request
			if req.Body == nil {
				req.Header.Set(ProxySaveErrorHeader, "profiling request body is nil")
				log.Errorf("profiling request body is nil")
				// Set req.URL to nil will trigger a proxy err and then the request will be terminated immediately
				req.URL = nil
				return
			}

			bodyBytes, err := ioutil.ReadAll(http.MaxBytesReader(nil, req.Body, profileMaxSize))
			if err != nil {
				req.Header.Set(ProxySaveErrorHeader, fmt.Sprintf("readall profile body err: %s", err))
				log.Errorf("read profile body err: %s", err)
				req.URL = nil
				return
			}
			_ = req.Body.Close()

			// use to repeatable read
			bodyReader := bytes.NewReader(bodyBytes)
			req.Body = ioutil.NopCloser(bodyReader)

			defer func() {
				_ = req.Body.Close()
				// reset http body
				if _, err := bodyReader.Seek(0, io.SeekStart); err != nil {
					log.Errorf("seek body to begin fail: %s", err)
				}
				req.Body = ioutil.NopCloser(bodyReader)
			}()

			wsID, err := queryWorkSpaceUUID()
			if err != nil {
				req.Header.Set(ProxySaveErrorHeader, fmt.Sprintf("query workspace uuid fail: %s", err))
				log.Errorf("query workspace id fail: %s", err)
				req.URL = nil
				return
			}

			profileID, unixNano, err := cache(req)
			if err != nil {
				req.Header.Set(ProxySaveErrorHeader, fmt.Sprintf("cache profile data fail: %s", err))
				log.Errorf("send profile to datakit io fail: %s", err)
				req.URL = nil
				return
			}

			req.URL = URL
			req.Host = URL.Host // must override the host

			log.Infof("receive profiling request, bodyLength: %d, datakit will proxy the request to url [%s], workspaceID: [%s]",
				req.ContentLength, URL.String(), wsID)

			req.Header.Set("X-Datakit-Workspace", wsID)
			req.Header.Set("X-Datakit-Profileid", profileID)
			req.Header.Set("X-Datakit-Unixnano", strconv.FormatInt(unixNano, 10))

			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},

		ModifyResponse: func(resp *http.Response) error {
			// log proxy error

			if resp.StatusCode/100 > 2 {
				log.Errorf("profile proxy response http status: %s", resp.Status)
			} else {
				log.Infof("profile proxy response http status: %s", resp.Status)
			}
			if resp.Body != nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Errorf("read profile proxy response body err: %s", err)
					return nil
				}
				if len(body) > 0 {
					_ = resp.Body.Close()
					resp.Body = ioutil.NopCloser(bytes.NewReader(body))
				}

				if resp.StatusCode/100 > 2 {
					log.Errorf("upload profile binary response: %s", string(body))
				} else {
					log.Infof("upload profile binary response: %s", string(body))

					var resp uploadResponse

					if err := json.Unmarshal(body, &resp); err != nil {
						return fmt.Errorf("json unmarshal upload profile binary response err: %w", err)
					}

					if resp.Content == nil || resp.Content.ProfileID == "" {
						return fmt.Errorf("fetch profile upload response profileID fail")
					}

					if err := sendToIO(resp.Content.ProfileID); err != nil {
						return fmt.Errorf("发送profile元数据失败: %w", err)
					}
				}
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			proxyErr := r.Header.Get(ProxySaveErrorHeader)
			if proxyErr != "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(proxyErr))
				log.Errorf("proxy error handler get err: %s, %s", proxyErr, err.Error())
			} else {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(err.Error()))
				log.Errorf("proxy error handler get err: %s", err.Error())
			}
		},
	}

	for _, endpoint := range i.Endpoints {
		dkhttp.RegHTTPHandler(http.MethodPost, endpoint, proxy.ServeHTTP)
		log.Infof("pattern: %s registered", endpoint)
	}
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("the input %s is running...", inputName)

	group := goroutine.NewGroup(goroutine.Option{
		Name: "profile",
		PanicCb: func(b []byte) bool {
			log.Error(string(b))
			return false
		},
	})

	for _, g := range i.Go {
		func(g *GoProfiler) {
			group.Go(func(ctx context.Context) error {
				if err := g.run(i); err != nil {
					log.Errorf("go profile collect error: %s", err.Error())
				}
				return nil
			})
		}(g)
	}

	if err := group.Wait(); err != nil {
		log.Errorf("profile collect err: %s", err.Error())
	}
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}
