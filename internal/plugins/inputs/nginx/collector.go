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
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// 默认 http stub status module 模块的数据.
func (ipt *Input) getStubStatusModuleMetric() {
	resp, err := ipt.client.Get(ipt.URL)
	if err != nil {
		l.Errorf("error making HTTP request to %s: %s", ipt.URL, err)
		ipt.lastErr = err
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		ipt.lastErr = fmt.Errorf("%s returned HTTP status %s", ipt.URL, resp.Status)
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

	tags := getTags(ipt.URL)
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	if ipt.Election {
		tags = inputs.MergeTags(ipt.Tagger.ElectionTags(), tags, ipt.URL)
	} else {
		tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, ipt.URL)
	}

	fields := map[string]interface{}{
		"connection_active":   active,
		"connection_accepts":  accepts,
		"connection_handled":  handled,
		"connection_requests": requests,
		"connection_reading":  reading,
		"connection_writing":  writing,
		"connection_waiting":  waiting,
	}
	metric := &NginxMeasurement{
		name:   nginx,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}
	ipt.collectCache = append(ipt.collectCache, metric.Point())
}

func (ipt *Input) getVTSMetric() {
	resp, err := ipt.client.Get(ipt.URL)
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
		ipt.handVTSResponse(resp.Body)
	default:
		l.Errorf("%s returned unexpected content type %s", ipt.URL, contentType)
	}
}

func (ipt *Input) handVTSResponse(r io.Reader) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		l.Errorf(err.Error())
		return
	}
	vtsResp := NginxVTSResponse{}
	if err := json.Unmarshal(body, &vtsResp); err != nil {
		l.Errorf("decoding JSON response err:%s", err.Error())
		return
	}
	t := time.Unix(0, vtsResp.Now*1000000)
	vtsResp.tags = getTags(ipt.URL)

	vtsResp.tags["host"] = vtsResp.HostName
	vtsResp.tags["nginx_version"] = vtsResp.Version

	for k, v := range ipt.Tags {
		vtsResp.tags[k] = v
	}

	ipt.makeConnectionsLine(vtsResp, t)
	ipt.makeServerZoneLine(vtsResp, t)
	ipt.makeUpstreamZoneLine(vtsResp, t)
	ipt.makeCacheZoneLine(vtsResp, t)
}

func (ipt *Input) makeConnectionsLine(vtsResp NginxVTSResponse, t time.Time) {
	tags := map[string]string{}
	for k, v := range vtsResp.tags {
		tags[k] = v
	}
	fields := map[string]interface{}{
		"load_timestamp":      vtsResp.LoadTimestamp,
		"connection_active":   vtsResp.Connections.Active,
		"connection_accepts":  vtsResp.Connections.Accepted,
		"connection_handled":  vtsResp.Connections.Handled,
		"connection_requests": vtsResp.Connections.Requests,
		"connection_reading":  vtsResp.Connections.Reading,
		"connection_writing":  vtsResp.Connections.Writing,
		"connection_waiting":  vtsResp.Connections.Waiting,
	}

	if ipt.Election {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
	} else {
		tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
	}

	metric := &NginxMeasurement{
		name:   nginx,
		tags:   tags,
		fields: fields,
		ts:     t,
	}
	ipt.collectCache = append(ipt.collectCache, metric.Point())
}

func (ipt *Input) makeServerZoneLine(vtsResp NginxVTSResponse, t time.Time) {
	tags := map[string]string{}
	for k, v := range vtsResp.tags {
		tags[k] = v
	}
	for k, v := range vtsResp.ServerZones {
		tags["server_zone"] = k
		fields := map[string]interface{}{
			"requests": v.RequestCounter,
			"received": v.InBytes,
			"send":     v.OutBytes,

			"response_1xx": v.Responses.OneXx,
			"response_2xx": v.Responses.TwoXx,
			"response_3xx": v.Responses.ThreeXx,
			"response_4xx": v.Responses.FourXx,
			"response_5xx": v.Responses.FiveXx,
		}

		if ipt.Election {
			tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
		} else {
			tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
		}

		metric := &NginxMeasurement{
			name:   ServerZone,
			tags:   tags,
			fields: fields,
			ts:     t,
		}
		ipt.collectCache = append(ipt.collectCache, metric.Point())
	}
}

func (ipt *Input) makeUpstreamZoneLine(vtsResp NginxVTSResponse, t time.Time) {
	tags := map[string]string{}
	for k, v := range vtsResp.tags {
		tags[k] = v
	}
	for upsteamName, upstreams := range vtsResp.UpstreamZones {
		tags["upstream_zone"] = upsteamName
		for _, upstream := range upstreams {
			tags["upstream_server"] = upstream.Server
			fields := map[string]interface{}{
				"request_count": upstream.RequestCounter,
				"received":      upstream.InBytes,
				"send":          upstream.OutBytes,

				"response_1xx": upstream.Responses.OneXx,
				"response_2xx": upstream.Responses.TwoXx,
				"response_3xx": upstream.Responses.ThreeXx,
				"response_4xx": upstream.Responses.FourXx,
				"response_5xx": upstream.Responses.FiveXx,
			}

			if ipt.Election {
				tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
			} else {
				tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
			}

			metric := &UpstreamZoneMeasurement{
				name:   UpstreamZone,
				tags:   tags,
				fields: fields,
				ts:     t,
			}
			ipt.collectCache = append(ipt.collectCache, metric.Point())
		}
	}
}

func (ipt *Input) makeCacheZoneLine(vtsResp NginxVTSResponse, t time.Time) {
	tags := map[string]string{}
	for k, v := range vtsResp.tags {
		tags[k] = v
	}
	for cacheName, cacheZone := range vtsResp.CacheZones {
		tags["cache_zone"] = cacheName
		fields := map[string]interface{}{
			"max_size":  cacheZone.MaxSize,
			"used_size": cacheZone.UsedSize,
			"received":  cacheZone.InBytes,
			"send":      cacheZone.OutBytes,

			"responses_miss":        cacheZone.Responses.Miss,
			"responses_bypass":      cacheZone.Responses.Bypass,
			"responses_expired":     cacheZone.Responses.Expired,
			"responses_stale":       cacheZone.Responses.Stale,
			"responses_updating":    cacheZone.Responses.Updating,
			"responses_revalidated": cacheZone.Responses.Revalidated,
			"responses_hit":         cacheZone.Responses.Hit,
			"responses_scarce":      cacheZone.Responses.Scarce,
		}

		if ipt.Election {
			tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
		} else {
			tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
		}

		metric := &CacheZoneMeasurement{
			name:   CacheZone,
			tags:   tags,
			fields: fields,
			ts:     t,
		}
		ipt.collectCache = append(ipt.collectCache, metric.Point())
	}
}
