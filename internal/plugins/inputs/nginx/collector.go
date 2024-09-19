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
	opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

	for k, v := range ipt.Tags {
		kvs = kvs.MustAddTag(k, v)
	}
	kvs = kvs.MustAddTag("nginx_server", ipt.host)
	kvs = kvs.MustAddTag("nginx_port", strconv.Itoa(port))

	kvs = kvs.Add("connection_active", active, false, true)
	kvs = kvs.Add("connection_accepts", accepts, false, true)
	kvs = kvs.Add("connection_handled", handled, false, true)
	kvs = kvs.Add("connection_requests", requests, false, true)
	kvs = kvs.Add("connection_reading", reading, false, true)
	kvs = kvs.Add("connection_writing", writing, false, true)
	kvs = kvs.Add("connection_waiting", waiting, false, true)
	kvs = kvs.Add("connection_dropped", accepts-handled, false, true)

	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(nginx, kvs, opts...))
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
	opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

	for k, v := range vtsResp.tags {
		kvs = kvs.MustAddTag(k, v)
	}

	kvs = kvs.Add("load_timestamp", vtsResp.LoadTimestamp, false, true)
	kvs = kvs.Add("connection_active", vtsResp.Connections.Active, false, true)
	kvs = kvs.Add("connection_accepts", vtsResp.Connections.Accepted, false, true)
	kvs = kvs.Add("connection_handled", vtsResp.Connections.Handled, false, true)
	kvs = kvs.Add("connection_requests", vtsResp.Connections.Requests, false, true)
	kvs = kvs.Add("connection_reading", vtsResp.Connections.Reading, false, true)
	kvs = kvs.Add("connection_writing", vtsResp.Connections.Writing, false, true)
	kvs = kvs.Add("connection_waiting", vtsResp.Connections.Waiting, false, true)

	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(nginx, kvs, opts...))
}

func (ipt *Input) makeServerZoneLine(vtsResp NginxVTSResponse) {
	for k, v := range vtsResp.ServerZones {
		kvs := make(point.KVs, 0, 13)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range vtsResp.tags {
			kvs = kvs.MustAddTag(kk, vv)
		}
		kvs = kvs.MustAddTag("server_zone", k)

		kvs = kvs.Add("requests", v.RequestCounter, false, true)
		kvs = kvs.Add("received", v.InBytes, false, true)
		kvs = kvs.Add("send", v.OutBytes, false, true)
		kvs = kvs.Add("response_1xx", v.Responses.OneXx, false, true)
		kvs = kvs.Add("response_2xx", v.Responses.TwoXx, false, true)
		kvs = kvs.Add("response_3xx", v.Responses.ThreeXx, false, true)
		kvs = kvs.Add("response_4xx", v.Responses.FourXx, false, true)
		kvs = kvs.Add("response_5xx", v.Responses.FiveXx, false, true)

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(ServerZone, kvs, opts...))
	}
}

func (ipt *Input) makeUpstreamZoneLine(vtsResp NginxVTSResponse) {
	for upsteamName, upstreams := range vtsResp.UpstreamZones {
		for _, upstream := range upstreams {
			kvs := make(point.KVs, 0, 14)
			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

			for kk, vv := range vtsResp.tags {
				kvs = kvs.MustAddTag(kk, vv)
			}
			kvs = kvs.MustAddTag("upstream_zone", upsteamName)
			kvs = kvs.MustAddTag("upstream_server", upstream.Server)

			kvs = kvs.Add("request_count", upstream.RequestCounter, false, true)
			kvs = kvs.Add("received", upstream.InBytes, false, true)
			kvs = kvs.Add("send", upstream.OutBytes, false, true)
			kvs = kvs.Add("response_1xx", upstream.Responses.OneXx, false, true)
			kvs = kvs.Add("response_2xx", upstream.Responses.TwoXx, false, true)
			kvs = kvs.Add("response_3xx", upstream.Responses.ThreeXx, false, true)
			kvs = kvs.Add("response_4xx", upstream.Responses.FourXx, false, true)
			kvs = kvs.Add("response_5xx", upstream.Responses.FiveXx, false, true)

			ipt.collectCache = append(ipt.collectCache, point.NewPointV2(UpstreamZone, kvs, opts...))
		}
	}
}

func (ipt *Input) makeCacheZoneLine(vtsResp NginxVTSResponse) {
	for cacheName, cacheZone := range vtsResp.CacheZones {
		kvs := make(point.KVs, 0, 17)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range vtsResp.tags {
			kvs = kvs.MustAddTag(kk, vv)
		}
		kvs = kvs.MustAddTag("cache_zone", cacheName)

		kvs = kvs.Add("max_size", cacheZone.MaxSize, false, true)
		kvs = kvs.Add("used_size", cacheZone.UsedSize, false, true)
		kvs = kvs.Add("received", cacheZone.InBytes, false, true)
		kvs = kvs.Add("send", cacheZone.OutBytes, false, true)
		kvs = kvs.Add("responses_miss", cacheZone.Responses.Miss, false, true)
		kvs = kvs.Add("responses_bypass", cacheZone.Responses.Bypass, false, true)
		kvs = kvs.Add("responses_expired", cacheZone.Responses.Expired, false, true)
		kvs = kvs.Add("responses_stale", cacheZone.Responses.Stale, false, true)
		kvs = kvs.Add("responses_updating", cacheZone.Responses.Updating, false, true)
		kvs = kvs.Add("responses_revalidated", cacheZone.Responses.Revalidated, false, true)
		kvs = kvs.Add("responses_hit", cacheZone.Responses.Hit, false, true)
		kvs = kvs.Add("responses_scarce", cacheZone.Responses.Scarce, false, true)

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(CacheZone, kvs, opts...))
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
	opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

	for k, v := range plusAPIResp.tags {
		kvs = kvs.MustAddTag(k, v)
	}

	kvs = kvs.MustAddTag("nginx_version", plusAPIResp.General.Version)
	kvs = kvs.Add("pid", plusAPIResp.General.Pid, false, true)
	kvs = kvs.Add("ppid", plusAPIResp.General.Ppid, false, true)

	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(nginx, kvs, opts...))
}

func (ipt *Input) makeServerLine(plusAPIResp NginxPlusAPIResponse) {
	for k, v := range plusAPIResp.Servers {
		kvs := make(point.KVs, 0, 20)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range plusAPIResp.tags {
			kvs = kvs.MustAddTag(kk, vv)
		}

		kvs = kvs.MustAddTag("server_zone", k)
		kvs = kvs.Add("processing", v.Processing, false, true)
		kvs = kvs.Add("requests", v.Requests, false, true)
		kvs = kvs.Add("responses", v.Responses.Total, false, true)
		kvs = kvs.Add("received", v.Received, false, true)
		kvs = kvs.Add("send", v.Sent, false, true)
		kvs = kvs.Add("discarded", v.Discarded, false, true)

		kvs = kvs.Add("response_1xx", v.Responses.OneXX, false, true)
		kvs = kvs.Add("response_2xx", v.Responses.TwoXX, false, true)
		kvs = kvs.Add("response_3xx", v.Responses.ThreeXX, false, true)
		kvs = kvs.Add("response_4xx", v.Responses.FourXX, false, true)
		kvs = kvs.Add("response_5xx", v.Responses.FiveXX, false, true)

		kvs = kvs.Add("code_200", v.Responses.Codes.Code200, false, true)
		kvs = kvs.Add("code_301", v.Responses.Codes.Code301, false, true)
		kvs = kvs.Add("code_404", v.Responses.Codes.Code404, false, true)
		kvs = kvs.Add("code_503", v.Responses.Codes.Code503, false, true)

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(ServerZone, kvs, opts...))
	}
}

func (ipt *Input) makeUpStreamLine(plusAPIResp NginxPlusAPIResponse) {
	for upsteamName, upstreams := range plusAPIResp.Upstreams {
		for _, upstream := range upstreams.Peers {
			kvs := make(point.KVs, 0, 20)
			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

			for kk, vv := range plusAPIResp.tags {
				kvs = kvs.MustAddTag(kk, vv)
			}
			kvs = kvs.MustAddTag("upstream_zone", upsteamName)
			kvs = kvs.MustAddTag("upstream_server", upstream.Server)

			kvs = kvs.Add("backup", upstream.Backup, false, true)
			kvs = kvs.Add("weight", upstream.Weight, false, true)
			kvs = kvs.Add("state", upstream.State, false, true)
			kvs = kvs.Add("active", upstream.Active, false, true)
			kvs = kvs.Add("request_count", upstream.Requests, false, true)
			kvs = kvs.Add("received", upstream.Received, false, true)
			kvs = kvs.Add("send", upstream.Sent, false, true)
			kvs = kvs.Add("fails", upstream.Fails, false, true)
			kvs = kvs.Add("unavail", upstream.Unavail, false, true)
			kvs = kvs.Add("response_1xx", upstream.Responses.OneXX, false, true)
			kvs = kvs.Add("response_2xx", upstream.Responses.TwoXX, false, true)
			kvs = kvs.Add("response_3xx", upstream.Responses.ThreeXX, false, true)
			kvs = kvs.Add("response_4xx", upstream.Responses.FourXX, false, true)
			kvs = kvs.Add("response_5xx", upstream.Responses.FiveXX, false, true)

			ipt.collectCache = append(ipt.collectCache, point.NewPointV2(UpstreamZone, kvs, opts...))
		}
	}
}

func (ipt *Input) makeCacheLine(plusAPIResp NginxPlusAPIResponse) {
	for k, v := range plusAPIResp.Caches {
		kvs := make(point.KVs, 0, 15)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range plusAPIResp.tags {
			kvs = kvs.MustAddTag(kk, vv)
		}
		kvs = kvs.MustAddTag("cache_zone", k)

		kvs = kvs.Add("used_size", v.Size, false, true)
		kvs = kvs.Add("max_size", v.MaxSize, false, true)

		kvs = kvs.Add("responses_hit", v.Hit.Bytes, false, true)
		kvs = kvs.Add("responses_stale", v.Stale.Bytes, false, true)
		kvs = kvs.Add("responses_updating", v.Updating.Bytes, false, true)
		kvs = kvs.Add("responses_revalidated", v.Revalidated.Bytes, false, true)
		kvs = kvs.Add("responses_miss", v.Miss.Bytes, false, true)
		kvs = kvs.Add("responses_expired", v.Expired.Bytes, false, true)
		kvs = kvs.Add("responses_bypass", v.Bypass.Bytes, false, true)

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(ServerZone, kvs, opts...))
	}
}

func (ipt *Input) makeLocationLine(plusAPIResp NginxPlusAPIResponse) {
	for locationName, location := range plusAPIResp.Locations {
		kvs := make(point.KVs, 0, 20)
		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.start), point.WithExtraTags(ipt.mergedTags))

		for kk, vv := range plusAPIResp.tags {
			kvs = kvs.MustAddTag(kk, vv)
		}
		kvs.MustAddTag("location_zone", locationName)

		kvs.Add("requests", location.Requests, false, true)
		kvs.Add("response", location.Responses.Total, false, true)
		kvs.Add("discarded", location.Discarded, false, true)
		kvs.Add("received", location.Received, false, true)
		kvs.Add("sent", location.Sent, false, true)
		kvs.Add("response_1xx", location.Responses.OneXX, false, true)
		kvs.Add("response_2xx", location.Responses.TwoXX, false, true)
		kvs.Add("response_3xx", location.Responses.ThreeXX, false, true)
		kvs.Add("response_4xx", location.Responses.FourXX, false, true)
		kvs.Add("response_5xx", location.Responses.FiveXX, false, true)
		kvs.Add("code_200", location.Responses.Codes.Code200, false, true)
		kvs.Add("code_301", location.Responses.Codes.Code301, false, true)
		kvs.Add("code_404", location.Responses.Codes.Code404, false, true)
		kvs.Add("code_503", location.Responses.Codes.Code503, false, true)

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(LocationZone, kvs, opts...))
	}
}
