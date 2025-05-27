// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (j *JolokiaAgent) Collect() {
	log = logger.SLogger("jolokia")
	if j.L == nil {
		j.L = log
	}
	j.L.Infof("%s input started...", j.PluginName)

	if j.g == nil {
		j.g = goroutine.NewGroup(goroutine.Option{Name: "jolokia-" + j.PluginName})
	}

	j = j.adaptor()

	var (
		duration time.Duration
		err      error
	)
	if len(j.Interval) > 0 {
		duration, err = time.ParseDuration(j.Interval)
		if err != nil {
			j.L.Error("time.ParseDuration: %s", err)
			return
		}
	} else {
		duration = time.Second * 10
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()
	start := time.Now()

	for {
		if j.pause {
			j.L.Debugf("Jolokia plugin %s paused", j.PluginName)
		} else {
			collectStart := time.Now()
			if err := j.Gather(start.UnixNano()); err != nil {
				j.Feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(j.PluginName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			}

			if len(j.collectCache) > 0 {
				if err := j.Feeder.FeedV2(point.Metric,
					j.collectCache,
					dkio.WithCollectCost(time.Since(collectStart)),
					dkio.WithElection(j.Election),
					dkio.WithInputName(j.PluginName)); err != nil {
					j.L.Errorf("Feed: %s, ignored", err.Error())
				}

				j.collectCache = j.collectCache[:0] // clean cache
			} else {
				j.L.Warn("no point, ignored")
			}
		}

		select {
		case tt := <-tick.C:
			start = time.UnixMilli(inputs.AlignTimeMillSec(tt, start.UnixMilli(), duration.Milliseconds()))

		case <-datakit.Exit.Wait():
			j.L.Infof("input %s exit", j.PluginName)
			return

		case <-j.SemStop.Wait():
			j.L.Infof("input %s return", j.PluginName)
			return

		case j.pause = <-j.pauseCh:
			j.L.Infof("Jolokia plugin %q paused? %v", j.PluginName, j.pause)
		}
	}
}

func (j *JolokiaAgent) Gather(ptTS int64) error {
	if j.gatherer == nil {
		j.gatherer = newGatherer(j.createMetrics())
	}

	// Initialize clients once
	if j.clients == nil {
		j.clients = make([]*jclient, 0, len(j.URLs))
		for _, url := range j.URLs {
			client, err := j.createClient(url)
			if err != nil {
				return err
			}
			j.clients = append(j.clients, client)
		}
	}

	for _, client := range j.clients {
		func(client *jclient) {
			j.g.Go(func(ctx context.Context) error {
				// up指标
				client.upState = 1

				pts, _ := j.collectCustomerObjectMeasurement(client)
				if err := j.Feeder.FeedV2(point.CustomObject, pts,
					dkio.WithCollectCost(time.Since(time.Now())),
					dkio.WithElection(j.Election),
					dkio.WithInputName(j.PluginName+"/CO")); err != nil {
					j.L.Errorf("Feed: %s, ignored", err.Error())
				}

				err := j.gatherer.gather(client, j, ptTS)
				if err != nil {
					client.upState = 0
					j.L.Errorf("unable to gather metrics for %s: %v", client.url, err)
				}

				pts, _ = j.buildUpPoints(client)
				if err := j.Feeder.FeedV2(point.Metric, pts,
					dkio.WithCollectCost(time.Since(time.Now())),
					dkio.WithElection(j.Election),
					dkio.WithInputName(j.PluginName)); err != nil {
					j.L.Errorf("Feed: %s, ignored", err.Error())
				}

				return nil
			})
		}(client)
	}

	return j.g.Wait()
}

func (j *JolokiaAgent) createMetrics() []Metric {
	var metrics []Metric

	for _, config := range j.Metrics {
		metrics = append(metrics,
			NewMetric(config, j.DefaultFieldPrefix, j.DefaultFieldSeparator, j.DefaultTagPrefix),
		)
	}

	return metrics
}

func (j *JolokiaAgent) createClient(url string) (*jclient, error) {
	return newClient(url, &jclientConfig{
		username:     j.Username,
		password:     j.Password,
		respTimeout:  j.ResponseTimeout,
		ClientConfig: j.ClientConfig,
	})
}

func (j *JolokiaAgent) adaptor() *JolokiaAgent {
	for i, m := range j.Metrics {
		var t string
		if m.FieldPrefix != nil {
			t = strings.ReplaceAll(*m.FieldPrefix, "#", "$")
			m.FieldPrefix = &t
		}

		if m.FieldSeparator != nil {
			t = strings.ReplaceAll(*m.FieldSeparator, "#", "$")
			m.FieldSeparator = &t
		}

		if m.FieldName != nil {
			t = strings.ReplaceAll(*m.FieldName, "#", "$")
			m.FieldName = &t
		}

		j.Metrics[i] = m
	}
	return j
}

// A Metric represents a specification for a
// Jolokia read request, and the transformations
// to apply to points generated from the responses.
type Metric struct {
	Name           string
	Mbean          string
	Paths          []string
	FieldName      string
	FieldPrefix    string
	FieldSeparator string
	TagPrefix      string
	TagKeys        []string

	mbeanDomain     string
	mbeanProperties []string
}

// MetricConfig represents a TOML form of a Metric with some optional fields.
type MetricConfig struct {
	Name           string   `toml:"name"`
	Mbean          string   `toml:"mbean"`
	Paths          []string `toml:"paths"`
	FieldName      *string  `toml:"field_name"`
	FieldPrefix    *string  `toml:"field_prefix"`
	FieldSeparator *string  `toml:"field_separator"`
	TagPrefix      *string  `toml:"tag_prefix"`
	TagKeys        []string `toml:"tag_keys"`
}

func NewMetric(config MetricConfig, defaultFieldPrefix, defaultFieldSeparator, defaultTagPrefix string) Metric {
	metric := Metric{
		Name:    config.Name,
		Mbean:   config.Mbean,
		Paths:   config.Paths,
		TagKeys: config.TagKeys,
	}

	if config.FieldName != nil {
		metric.FieldName = *config.FieldName
	}

	if config.FieldPrefix == nil {
		metric.FieldPrefix = defaultFieldPrefix
	} else {
		metric.FieldPrefix = *config.FieldPrefix
	}

	if config.FieldSeparator == nil {
		metric.FieldSeparator = defaultFieldSeparator
	} else {
		metric.FieldSeparator = *config.FieldSeparator
	}

	if config.TagPrefix == nil {
		metric.TagPrefix = defaultTagPrefix
	} else {
		metric.TagPrefix = *config.TagPrefix
	}

	mbeanDomain, mbeanProperties := parseMbeanObjectName(config.Mbean)
	metric.mbeanDomain = mbeanDomain
	metric.mbeanProperties = mbeanProperties

	return metric
}

func (m Metric) MatchObjectName(name string) bool {
	if name == m.Mbean {
		return true
	}

	mbeanDomain, mbeanProperties := parseMbeanObjectName(name)
	if mbeanDomain != m.mbeanDomain {
		return false
	}

	if len(mbeanProperties) != len(m.mbeanProperties) {
		return false
	}

NEXT_PROPERTY:
	for _, mbeanProperty := range m.mbeanProperties {
		for i := range mbeanProperties {
			if mbeanProperties[i] == mbeanProperty {
				continue NEXT_PROPERTY
			}
		}

		return false
	}

	return true
}

func (m Metric) MatchAttributeAndPath(attribute, innerPath string) bool {
	path := attribute
	if innerPath != "" {
		path = path + "/" + innerPath
	}

	for i := range m.Paths {
		if path == m.Paths[i] {
			return true
		}
	}

	return false
}

func parseMbeanObjectName(name string) (string, []string) {
	index := strings.Index(name, ":")
	if index == -1 {
		return name, []string{}
	}

	domain := name[:index]

	if index+1 > len(name) {
		return domain, []string{}
	}

	return domain, strings.Split(name[index+1:], ",")
}

func (j *JolokiaAgent) buildUpPoints(client *jclient) ([]*point.Point, error) {
	var CoPts []*point.Point
	uu, _ := url.Parse(client.url)
	h, p, err := net.SplitHostPort(uu.Host)
	var host string
	var port string
	if err == nil {
		host = h
		port = p
	} else {
		host = uu.Host
		j.L.Errorf("failed to split host and port: %s", err)
	}

	tags := map[string]string{
		"job":      j.PluginName,
		"instance": fmt.Sprintf("%s:%s", host, port),
	}
	fields := map[string]interface{}{
		"up": client.upState,
	}

	Copt := &inputs.UpMeasurement{
		Name:     "collector",
		Tags:     tags,
		Fields:   fields,
		Election: j.Election,
	}

	CoPts = append(CoPts, Copt.Point())
	if len(CoPts) > 0 {
		for k, v := range j.Tags {
			for _, pt := range CoPts {
				pt.AddTag(k, v)
			}
		}
		return CoPts, nil
	}
	return []*point.Point{}, nil
}
