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
	opts = append(opts, point.WithTime(ipt.start))

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
	opts = append(opts, point.WithTime(ipt.start))

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
		opts = append(opts, point.WithTime(ipt.start))

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
			opts = append(opts, point.WithTime(ipt.start))

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
		opts = append(opts, point.WithTime(ipt.start))

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
