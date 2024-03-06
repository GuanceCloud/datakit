// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var (
	defaultPrometheusioConnectKeepAlive = time.Second * 20
	defaultPromElection                 = false /*collect self node, not election*/
)

type promRunner struct {
	conf     *promConfig
	pm       *iprom.Prom
	feeder   dkio.Feeder
	lastTime time.Time

	currentURL   string
	instanceTags map[string]string // map["urlstr"] = "url.Host"
}

func newPromRunnerWithTomlConfig(discovery *Discovery, configStr string) ([]*promRunner, error) {
	cfgs, err := parsePromConfigs(configStr)
	if err != nil {
		return nil, fmt.Errorf("parse config error: %w", err)
	}

	var res []*promRunner

	for _, c := range cfgs {
		p, err := newPromRunnerWithConfig(discovery, c)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	return res, nil
}

func newPromRunnerWithConfig(discovery *Discovery, c *promConfig) (*promRunner, error) {
	p := &promRunner{
		conf:         c,
		feeder:       discovery.cfg.Feeder,
		lastTime:     time.Now(),
		instanceTags: make(map[string]string),
	}

	hosts, err := parseURLHost(c)
	if err != nil {
		return nil, fmt.Errorf("parse urls error: %w", err)
	}
	p.instanceTags = hosts

	streamSize := discovery.cfg.StreamSize

	callbackFunc := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		// append instance tag to points
		if instance, ok := p.instanceTags[p.currentURL]; ok {
			for _, pt := range pts {
				pt.AddTag("instance", instance)
			}
		}

		if p.conf.AsLogging != nil && p.conf.AsLogging.Enable {
			for _, pt := range pts {
				err := p.feeder.FeedV2(point.Logging, []*point.Point{pt},
					dkio.WithCollectCost(time.Since(p.lastTime)),
					dkio.WithElection(defaultPromElection),
					dkio.WithInputName(pt.Name()),
					dkio.WithBlocking(true))
				if err != nil {
					klog.Warnf("failed to feed prom logging: %s, ignored", err)
				}
			}
		} else {
			err := p.feeder.FeedV2(point.Metric, pts,
				dkio.WithCollectCost(time.Since(p.lastTime)),
				dkio.WithElection(defaultPromElection),
				dkio.WithInputName(p.conf.Source),
				dkio.WithBlocking(true))
			if err != nil {
				klog.Warnf("failed to feed prom metrics: %s, ignored", err)
			}
		}
		return nil
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(c.Source),
		iprom.WithTimeout(c.Timeout),
		iprom.WithKeepAlive(defaultPrometheusioConnectKeepAlive),
		iprom.WithIgnoreReqErr(c.IgnoreReqErr),
		iprom.WithMetricTypes(c.MetricTypes),
		iprom.WithMetricNameFilter(c.MetricNameFilter),
		iprom.WithMetricNameFilterIgnore(c.MetricNameFilterIgnore),
		iprom.WithMeasurementPrefix(c.MeasurementPrefix),
		iprom.WithMeasurementName(c.MeasurementName),
		iprom.WithMeasurements(c.Measurements),
		iprom.WithTLSOpen(c.TLSOpen),
		iprom.WithUDSPath(c.UDSPath),
		iprom.WithCacertFile(c.CacertFile),
		iprom.WithCertFile(c.CertFile),
		iprom.WithKeyFile(c.KeyFile),
		iprom.WithTagsIgnore(c.TagsIgnore),
		iprom.WithTagsRename(c.TagsRename),
		iprom.WithAsLogging(c.AsLogging),
		iprom.WithIgnoreTagKV(c.IgnoreTagKV),
		iprom.WithHTTPHeaders(c.HTTPHeaders),
		iprom.WithTags(c.Tags),
		iprom.WithDisableInfoTag(c.DisableInfoTag),
		iprom.WithAuth(c.Auth),
		iprom.WithMaxBatchCallback(streamSize, callbackFunc),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create prom: %w", err)
	}

	p.pm = pm
	return p, nil
}

func (p *promRunner) runOnce() {
	if p.conf == nil {
		return
	}
	if time.Since(p.lastTime) < p.conf.Interval {
		return
	}

	klog.Debugf("running collect from source %s", p.conf.Source)
	p.lastTime = time.Now()

	for _, u := range p.conf.URLs {
		p.currentURL = u
		// use callback processor, not return pts
		_, err := p.pm.CollectFromHTTPV2(u)
		if err != nil {
			klog.Warnf("failed to collect prom: %s", err)
			return
		}
	}
}
