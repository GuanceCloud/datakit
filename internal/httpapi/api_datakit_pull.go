// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkfilter "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
)

const (
	defaultDatakitPullInterval = 30 * time.Minute
	datakitPullCacheTTL        = time.Second
)

var datakitPullCategories = []string{
	datakit.CategoryRUM,
	datakit.CategoryLogging,
}

var defaultDatakitPullCache datakitPullCache

type datakitPullResponse struct {
	Filters      map[string]dkfilter.FilterConditions `json:"filters"`
	PullInterval string                               `json:"pull_interval"`
}

type datakitPullCache struct {
	sync.Mutex

	path      string
	modTime   time.Time
	size      int64
	checkedAt time.Time
	resp      *datakitPullResponse
}

func RegDatakitPullHTTPRoute() {
	RegHTTPRoute(http.MethodGet, datakit.DatakitPull, apiDatakitPull)
}

func apiDatakitPull(_ http.ResponseWriter, req *http.Request, _ ...interface{}) (interface{}, error) {
	if req.ContentLength != 0 {
		return nil, uhttp.Error(ErrInvalidRequest, "GET /v1/datakit/pull does not accept request body")
	}

	if req.URL.Query().Get("filters") != "true" {
		return nil, uhttp.Error(ErrInvalidRequest, "only filters=true is supported")
	}

	body, err := json.Marshal(defaultDatakitPullCache.load(filepath.Join(datakit.DataDir, ".pull")))
	if err != nil {
		return nil, uhttp.Error(ErrInvalidAPIHandler, "marshal datakit pull response failed")
	}

	return uhttp.RawJSONBody(body), nil
}

func (c *datakitPullCache) load(path string) *datakitPullResponse {
	now := time.Now()

	c.Lock()
	defer c.Unlock()

	samePath := c.path == path
	if c.resp != nil && samePath && now.Sub(c.checkedAt) < datakitPullCacheTTL {
		return c.resp
	}

	info, err := os.Stat(path)
	c.checkedAt = now
	c.path = path

	if err != nil {
		if !os.IsNotExist(err) {
			l.Debugf("stat datakit pull cache failed: %s", err)
		}

		c.modTime = time.Time{}
		c.size = 0
		c.resp = newDefaultDatakitPullResponse()
		return c.resp
	}

	if c.resp != nil && samePath && info.ModTime().Equal(c.modTime) && info.Size() == c.size {
		return c.resp
	}

	c.resp = readDatakitPullResponse(path)
	c.modTime = info.ModTime()
	c.size = info.Size()
	return c.resp
}

func readDatakitPullResponse(path string) *datakitPullResponse {
	body, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			l.Debugf("read datakit pull cache failed: %s", err)
		}
		return newDefaultDatakitPullResponse()
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return newDefaultDatakitPullResponse()
	}

	var cache dkfilter.Filters
	if err := json.Unmarshal(body, &cache); err != nil {
		l.Debugf("decode datakit pull cache failed: %s", err)
		return newDefaultDatakitPullResponse()
	}

	resp := newDefaultDatakitPullResponse()
	for _, category := range datakitPullCategories {
		if rules, ok := cache.Filters[category]; ok {
			resp.Filters[category] = rules
		}
	}

	if cache.PullInterval > 0 {
		resp.PullInterval = formatDatakitPullInterval(cache.PullInterval)
	}

	return resp
}

func newDefaultDatakitPullResponse() *datakitPullResponse {
	filters := make(map[string]dkfilter.FilterConditions, len(datakitPullCategories))
	for _, category := range datakitPullCategories {
		filters[category] = dkfilter.FilterConditions{}
	}

	return &datakitPullResponse{
		Filters:      filters,
		PullInterval: formatDatakitPullInterval(defaultDatakitPullInterval),
	}
}

func formatDatakitPullInterval(duration time.Duration) string {
	switch {
	case duration%time.Hour == 0:
		return strconv.FormatInt(int64(duration/time.Hour), 10) + "h"
	case duration%time.Minute == 0:
		return strconv.FormatInt(int64(duration/time.Minute), 10) + "m"
	default:
		return duration.String()
	}
}
