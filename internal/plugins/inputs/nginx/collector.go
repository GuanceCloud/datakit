// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

// 默认 http stub status module 模块的数据.
func (ipt *Input) getStubStatusModuleMetric(port int) {
	u := ipt.host + ":" + strconv.Itoa(port) + ipt.path
	resp, err := ipt.client.Get(u)
	if err != nil {
		l.Debugf("%s", err)
		ipt.lastErr = err
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		ipt.lastErr = fmt.Errorf("%s returned HTTP status %s", u, resp.Status)
		return
	}
	r := bufio.NewReader(resp.Body)

	// Active connections
	_, err = r.ReadString(':')
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	line, err := r.ReadString('\n')
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	active, err := strconv.ParseUint(strings.TrimSpace(line), 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}

	// Server accepts handled requests
	_, err = r.ReadString('\n')
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	line, err = r.ReadString('\n')
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	data := strings.Fields(line)
	accepts, err := strconv.ParseUint(data[0], 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}

	handled, err := strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	requests, err := strconv.ParseUint(data[2], 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}

	// Reading/Writing/Waiting
	line, err = r.ReadString('\n')
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	data = strings.Fields(line)
	reading, err := strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	writing, err := strconv.ParseUint(data[3], 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}
	waiting, err := strconv.ParseUint(data[5], 10, 64)
	if err != nil {
		l.Errorf("parse err:%s", err.Error())
		return
	}

	kvs := make(point.KVs, 0, 9)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

	for k, v := range ipt.Tags {
		kvs = kvs.SetTag(k, v)
	}
	kvs = kvs.SetTag("nginx_server", ipt.host)
	kvs = kvs.SetTag("nginx_port", strconv.Itoa(port))

	kvs = kvs.Set("connection_active", active)
	kvs = kvs.Set("connection_accepts", accepts)
	kvs = kvs.Set("connection_handled", handled)
	kvs = kvs.Set("connection_requests", requests)
	kvs = kvs.Set("connection_reading", reading)
	kvs = kvs.Set("connection_writing", writing)
	kvs = kvs.Set("connection_waiting", waiting)
	kvs = kvs.Set("connection_dropped", accepts-handled)

	ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementNginx, kvs, opts...))
}

func (ipt *Input) getVTSMetric(port int) {
	u := ipt.host + ":" + strconv.Itoa(port) + ipt.path
	resp, err := ipt.client.Get(u)
	if err != nil {
		l.Debugf("%s", err)
		ipt.lastErr = err
		return
	}

	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		l.Errorf("%s returned HTTP status %s", u, resp.Status)
		return
	}
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	switch contentType {
	case "application/json":
		ipt.handVTSResponse(resp.Body, port)
	default:
		l.Errorf("%s returned unexpected content type %s", u, contentType)
	}
}

func (ipt *Input) handVTSResponse(r io.Reader, port int) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		l.Errorf(err.Error())
		return
	}
	vtsResp := NginxVTSResponse{tags: map[string]string{}}
	if err := json.Unmarshal(body, &vtsResp); err != nil {
		l.Errorf("decoding JSON response err:%s", err.Error())
		return
	}
	vtsResp.tags["nginx_server"] = ipt.host
	vtsResp.tags["nginx_port"] = strconv.Itoa(port)
	vtsResp.tags["host"] = vtsResp.HostName
	vtsResp.tags["nginx_version"] = vtsResp.Version
	for k, v := range ipt.Tags {
		vtsResp.tags[k] = v
	}

	ipt.makeConnectionsLine(vtsResp)
	ipt.makeServerZoneLine(vtsResp)
	ipt.makeUpstreamZoneLine(vtsResp)
	ipt.makeCacheZoneLine(vtsResp)
}

func (ipt *Input) makeConnectionsLine(vtsResp NginxVTSResponse) {
	kvs := make(point.KVs, 0, 12)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

	for k, v := range vtsResp.tags {
		kvs = kvs.SetTag(k, v)
	}

	kvs = kvs.Set("load_timestamp", vtsResp.LoadTimestamp)
	kvs = kvs.Set("connection_active", vtsResp.Connections.Active)
	kvs = kvs.Set("connection_accepts", vtsResp.Connections.Accepted)
	kvs = kvs.Set("connection_handled", vtsResp.Connections.Handled)
	kvs = kvs.Set("connection_requests", vtsResp.Connections.Requests)
	kvs = kvs.Set("connection_reading", vtsResp.Connections.Reading)
	kvs = kvs.Set("connection_writing", vtsResp.Connections.Writing)
	kvs = kvs.Set("connection_waiting", vtsResp.Connections.Waiting)

	ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementNginx, kvs, opts...))
}

func (ipt *Input) makeServerZoneLine(vtsResp NginxVTSResponse) {
	for k, v := range vtsResp.ServerZones {
		kvs := make(point.KVs, 0, 13)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range vtsResp.tags {
			kvs = kvs.SetTag(kk, vv)
		}
		kvs = kvs.SetTag("server_zone", k)

		kvs = kvs.Set("requests", v.RequestCounter)
		kvs = kvs.Set("received", v.InBytes)
		kvs = kvs.Set("send", v.OutBytes)
		kvs = kvs.Set("response_1xx", v.Responses.OneXx)
		kvs = kvs.Set("response_2xx", v.Responses.TwoXx)
		kvs = kvs.Set("response_3xx", v.Responses.ThreeXx)
		kvs = kvs.Set("response_4xx", v.Responses.FourXx)
		kvs = kvs.Set("response_5xx", v.Responses.FiveXx)

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementServerZone, kvs, opts...))
	}
}

func (ipt *Input) makeUpstreamZoneLine(vtsResp NginxVTSResponse) {
	for upsteamName, upstreams := range vtsResp.UpstreamZones {
		for _, upstream := range upstreams {
			kvs := make(point.KVs, 0, 14)
			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

			for kk, vv := range vtsResp.tags {
				kvs = kvs.SetTag(kk, vv)
			}
			kvs = kvs.SetTag("upstream_zone", upsteamName)
			kvs = kvs.SetTag("upstream_server", upstream.Server)

			kvs = kvs.Set("request_count", upstream.RequestCounter)
			kvs = kvs.Set("received", upstream.InBytes)
			kvs = kvs.Set("send", upstream.OutBytes)
			kvs = kvs.Set("response_1xx", upstream.Responses.OneXx)
			kvs = kvs.Set("response_2xx", upstream.Responses.TwoXx)
			kvs = kvs.Set("response_3xx", upstream.Responses.ThreeXx)
			kvs = kvs.Set("response_4xx", upstream.Responses.FourXx)
			kvs = kvs.Set("response_5xx", upstream.Responses.FiveXx)

			ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementUpstreamZone, kvs, opts...))
		}
	}
}

func (ipt *Input) makeCacheZoneLine(vtsResp NginxVTSResponse) {
	for cacheName, cacheZone := range vtsResp.CacheZones {
		kvs := make(point.KVs, 0, 17)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range vtsResp.tags {
			kvs = kvs.SetTag(kk, vv)
		}
		kvs = kvs.SetTag("cache_zone", cacheName)

		kvs = kvs.Set("max_size", cacheZone.MaxSize)
		kvs = kvs.Set("used_size", cacheZone.UsedSize)
		kvs = kvs.Set("received", cacheZone.InBytes)
		kvs = kvs.Set("send", cacheZone.OutBytes)
		kvs = kvs.Set("responses_miss", cacheZone.Responses.Miss)
		kvs = kvs.Set("responses_bypass", cacheZone.Responses.Bypass)
		kvs = kvs.Set("responses_expired", cacheZone.Responses.Expired)
		kvs = kvs.Set("responses_stale", cacheZone.Responses.Stale)
		kvs = kvs.Set("responses_updating", cacheZone.Responses.Updating)
		kvs = kvs.Set("responses_revalidated", cacheZone.Responses.Revalidated)
		kvs = kvs.Set("responses_hit", cacheZone.Responses.Hit)
		kvs = kvs.Set("responses_scarce", cacheZone.Responses.Scarce)

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementCacheZone, kvs, opts...))
	}
}

func (ipt *Input) getPlusMetric() {
	plusAPIResp := NginxPlusAPIResponse{tags: make(map[string]string)}
	for _, plusAPI := range PlusAPIEndpoints {
		resp, err := ipt.client.Get(ipt.PlusAPIURL + "/" + plusAPI.endpoint)
		if err != nil {
			l.Errorf("error making HTTP request to %s: %s", ipt.URL, err)
			ipt.lastErr = err
			return
		}
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusOK {
			l.Errorf("%s returned HTTP status %s", ipt.URL, resp.Status)
			return
		}
		contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
		switch contentType {
		case "application/json":
			ipt.handlePlusAPIResponse(resp.Body, &plusAPIResp, plusAPI.nest)
		default:
			l.Errorf("%s returned unexpected content type %s", ipt.URL, contentType)
		}
	}
}

func (ipt *Input) handlePlusAPIResponse(r io.Reader, plusAPIResp *NginxPlusAPIResponse, nest string) {
	body, err := io.ReadAll(r)
	if err != nil {
		l.Errorf(err.Error())
		return
	}

	switch nest {
	case NestGeneral:
		if err := json.Unmarshal(body, &plusAPIResp.General); err != nil {
			l.Errorf("decoding JSON response err:%s", err.Error())
			return
		}
		ipt.makeNginxLine(*plusAPIResp)
	case NestServerZone:
		if err := json.Unmarshal(body, &plusAPIResp.Servers); err != nil {
			l.Errorf("decoding JSON response err:%s", err.Error())
			return
		}
		ipt.makeServerLine(*plusAPIResp)
	case NestUpstreams:
		if err := json.Unmarshal(body, &plusAPIResp.Upstreams); err != nil {
			l.Errorf("decoding JSON response err:%s", err.Error())
			return
		}
		ipt.makeUpStreamLine(*plusAPIResp)
	case NestCaches:
		if err := json.Unmarshal(body, &plusAPIResp.Caches); err != nil {
			l.Errorf("decoding JSON response err:%s", err.Error())
			return
		}
		ipt.makeCacheLine(*plusAPIResp)
	case NestLocationZones:
		if err := json.Unmarshal(body, &plusAPIResp.Locations); err != nil {
			l.Errorf("decoding JSON response err:%s", err.Error())
			return
		}
		ipt.makeLocationLine(*plusAPIResp)
	}
}

func (ipt *Input) makeNginxLine(plusAPIResp NginxPlusAPIResponse) {
	kvs := make(point.KVs, 0, 10)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

	for k, v := range plusAPIResp.tags {
		kvs = kvs.SetTag(k, v)
	}

	kvs = kvs.SetTag("nginx_version", plusAPIResp.General.Version)
	kvs = kvs.Set("pid", plusAPIResp.General.Pid)
	kvs = kvs.Set("ppid", plusAPIResp.General.Ppid)

	ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementNginx, kvs, opts...))
}

func (ipt *Input) makeServerLine(plusAPIResp NginxPlusAPIResponse) {
	for k, v := range plusAPIResp.Servers {
		kvs := make(point.KVs, 0, 20)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range plusAPIResp.tags {
			kvs = kvs.SetTag(kk, vv)
		}

		kvs = kvs.SetTag("server_zone", k)
		kvs = kvs.Set("processing", v.Processing)
		kvs = kvs.Set("requests", v.Requests)
		kvs = kvs.Set("responses", v.Responses.Total)
		kvs = kvs.Set("received", v.Received)
		kvs = kvs.Set("send", v.Sent)
		kvs = kvs.Set("discarded", v.Discarded)

		kvs = kvs.Set("response_1xx", v.Responses.OneXX)
		kvs = kvs.Set("response_2xx", v.Responses.TwoXX)
		kvs = kvs.Set("response_3xx", v.Responses.ThreeXX)
		kvs = kvs.Set("response_4xx", v.Responses.FourXX)
		kvs = kvs.Set("response_5xx", v.Responses.FiveXX)

		kvs = kvs.Set("code_200", v.Responses.Codes.Code200)
		kvs = kvs.Set("code_301", v.Responses.Codes.Code301)
		kvs = kvs.Set("code_404", v.Responses.Codes.Code404)
		kvs = kvs.Set("code_503", v.Responses.Codes.Code503)

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementServerZone, kvs, opts...))
	}
}

func (ipt *Input) makeUpStreamLine(plusAPIResp NginxPlusAPIResponse) {
	for upsteamName, upstreams := range plusAPIResp.Upstreams {
		for _, upstream := range upstreams.Peers {
			kvs := make(point.KVs, 0, 20)
			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

			for kk, vv := range plusAPIResp.tags {
				kvs = kvs.SetTag(kk, vv)
			}
			kvs = kvs.SetTag("upstream_zone", upsteamName)
			kvs = kvs.SetTag("upstream_server", upstream.Server)

			kvs = kvs.Set("backup", upstream.Backup)
			kvs = kvs.Set("weight", upstream.Weight)
			kvs = kvs.Set("state", upstream.State)
			kvs = kvs.Set("active", upstream.Active)
			kvs = kvs.Set("request_count", upstream.Requests)
			kvs = kvs.Set("received", upstream.Received)
			kvs = kvs.Set("send", upstream.Sent)
			kvs = kvs.Set("fails", upstream.Fails)
			kvs = kvs.Set("unavail", upstream.Unavail)
			kvs = kvs.Set("response_1xx", upstream.Responses.OneXX)
			kvs = kvs.Set("response_2xx", upstream.Responses.TwoXX)
			kvs = kvs.Set("response_3xx", upstream.Responses.ThreeXX)
			kvs = kvs.Set("response_4xx", upstream.Responses.FourXX)
			kvs = kvs.Set("response_5xx", upstream.Responses.FiveXX)

			ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementUpstreamZone, kvs, opts...))
		}
	}
}

func (ipt *Input) makeCacheLine(plusAPIResp NginxPlusAPIResponse) {
	for k, v := range plusAPIResp.Caches {
		kvs := make(point.KVs, 0, 15)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range plusAPIResp.tags {
			kvs = kvs.SetTag(kk, vv)
		}
		kvs = kvs.SetTag("cache_zone", k)

		kvs = kvs.Set("used_size", v.Size)
		kvs = kvs.Set("max_size", v.MaxSize)

		kvs = kvs.Set("responses_hit", v.Hit.Bytes)
		kvs = kvs.Set("responses_stale", v.Stale.Bytes)
		kvs = kvs.Set("responses_updating", v.Updating.Bytes)
		kvs = kvs.Set("responses_revalidated", v.Revalidated.Bytes)
		kvs = kvs.Set("responses_miss", v.Miss.Bytes)
		kvs = kvs.Set("responses_expired", v.Expired.Bytes)
		kvs = kvs.Set("responses_bypass", v.Bypass.Bytes)

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementServerZone, kvs, opts...))
	}
}

func (ipt *Input) makeLocationLine(plusAPIResp NginxPlusAPIResponse) {
	for locationName, location := range plusAPIResp.Locations {
		kvs := make(point.KVs, 0, 20)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range plusAPIResp.tags {
			kvs = kvs.SetTag(kk, vv)
		}
		kvs.SetTag("location_zone", locationName)

		kvs.Set("requests", location.Requests)
		kvs.Set("response", location.Responses.Total)
		kvs.Set("discarded", location.Discarded)
		kvs.Set("received", location.Received)
		kvs.Set("sent", location.Sent)
		kvs.Set("response_1xx", location.Responses.OneXX)
		kvs.Set("response_2xx", location.Responses.TwoXX)
		kvs.Set("response_3xx", location.Responses.ThreeXX)
		kvs.Set("response_4xx", location.Responses.FourXX)
		kvs.Set("response_5xx", location.Responses.FiveXX)
		kvs.Set("code_200", location.Responses.Codes.Code200)
		kvs.Set("code_301", location.Responses.Codes.Code301)
		kvs.Set("code_404", location.Responses.Codes.Code404)
		kvs.Set("code_503", location.Responses.Codes.Code503)

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(measurementLocationZone, kvs, opts...))
	}
}
