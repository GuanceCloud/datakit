// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	redisObjectMeasurementName = "database"
	redisDatabaseType          = "Redis"
)

var objectFeedSource = dkio.FeedSource(inputName, "object")

type redisObjectMeasurement struct{}

func (*redisObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   redisObjectMeasurementName,
		Cat:    point.Object,
		Desc:   "Redis object per node (version, uptime, QPS, and settings summary).",
		DescZh: "Redis 节点对象，包含版本、运行时间、QPS 和配置摘要。",
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "Node host"},
			"server":        &inputs.TagInfo{Desc: "Node address host:port"},
			"name":          &inputs.TagInfo{Desc: "Object name, same as server"},
			"port":          &inputs.TagInfo{Desc: "Node port"},
			"version":       &inputs.TagInfo{Desc: "Redis version"},
			"database_type": &inputs.TagInfo{Desc: "Database type, Redis"},
		},
		Fields: map[string]interface{}{
			"uptime": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Server uptime in seconds",
			},
			"qps": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Queries per second, from instantaneous_ops_per_sec",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.UnknownType,
				Unit:     inputs.UnknownUnit,
				Desc:     "Summary of Redis settings",
			},
		},
	}
}

type redisObjectMessage struct {
	Setting map[string]string `json:"setting"`
}

func (m *redisObjectMessage) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// parseInfoObject extracts redis_version, uptime_in_seconds, and instantaneous_ops_per_sec from INFO (server+stats).
func parseInfoObject(info string) (version string, uptime int64, qps float64, err error) {
	scanner := bufio.NewScanner(strings.NewReader(info))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		k, v := line[:idx], strings.TrimSpace(line[idx+1:])
		switch k {
		case "redis_version":
			version = v
		case "uptime_in_seconds":
			uptime, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return "", 0, 0, err
			}
		case "instantaneous_ops_per_sec":
			qps, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return "", 0, 0, err
			}
		}
	}
	return version, uptime, qps, scanner.Err()
}

func getObjectInfo(ctx context.Context, cli collectorClient) (version string, uptime int64, qps float64, err error) {
	info, err := cli.info(ctx, "ALL")
	if err != nil {
		return "", 0, 0, err
	}
	return parseInfoObject(info)
}

func getObjectSettings(ctx context.Context, cli collectorClient) (map[string]string, error) {
	conf, err := cli.configGet(ctx, "*")
	if err != nil {
		return nil, err
	}
	return sanitizeObjectSettings(conf), nil
}

func sanitizeObjectSettings(conf map[string]string) map[string]string {
	settings := make(map[string]string, len(conf))
	for k, v := range conf {
		if v == "" {
			v = notSet
		}
		if isCredentialConfig(k) && v != notSet {
			v = CREDENTIALSTR
		}
		settings[k] = v
	}
	return settings
}

func isCredentialConfig(key string) bool {
	switch key {
	case "requirepass", "masterauth", "tls-key-file-pass", "tls-client-key-file-pass":
		return true
	default:
		return false
	}
}

func (ipt *Input) collectObject(ctx context.Context) {
	if !ipt.Object.Enable || len(ipt.instances) == 0 {
		return
	}
	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, ipt.timeoutDuration)
	defer cancel()

	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, inst := range ipt.instances {
		for _, n := range inst.nodes() {
			version, uptime, qps, err := getObjectInfo(timeoutCtx, n.cli)
			if err != nil {
				l.Debugf("object %s: %s", n.addr, err)
				continue
			}
			if version == "" {
				version = "unknown"
			}
			settings, err := getObjectSettings(timeoutCtx, n.cli)
			if err != nil {
				l.Debugf("object settings %s: %s", n.addr, err)
				settings = map[string]string{}
			}

			port := portFromAddr(n.addr)
			var kvs point.KVs
			kvs = kvs.AddTag("database_type", redisDatabaseType).
				AddTag("name", n.addr).
				AddTag("port", port).
				AddTag("version", version).
				Add("uptime", uptime).
				Add("qps", qps).
				Add("message", (&redisObjectMessage{Setting: settings}).String())

			for k, v := range inst.buildNodeTags(n.addr, n.host) {
				kvs = kvs.AddTag(k, v)
			}
			pts = append(pts, point.NewPoint(redisObjectMeasurementName, kvs, opts...))
		}
	}

	if len(pts) > 0 {
		if err := ipt.feeder.Feed(point.Object, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(objectFeedSource)); err != nil {
			l.Warnf("feed object: %s", err)
		}
	}
}

func (ipt *Input) runObjectCollector(ctx context.Context) {
	if !ipt.Object.Enable {
		return
	}
	ticker := time.NewTicker(ipt.Object.Interval.Duration)
	defer ticker.Stop()

	ipt.collectObject(ctx)
	for {
		select {
		case <-ctx.Done():
			l.Info("object collector stopped")
			return
		case <-ticker.C:
			ipt.collectObject(ctx)
		}
	}
}

func portFromAddr(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return port
}
