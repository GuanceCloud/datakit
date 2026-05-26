// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jvm

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/jolokia"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var upFeedSource = dkio.FeedSource(inputName, "up")

// feedUpMetric feeds up metric for a client.
func (ipt *Input) feedUpMetric(client *jolokia.Client, upState int) {
	clientURL := client.URL()
	tags := map[string]string{
		"job":      inputName,
		"instance": ipt.getUpInstance(clientURL),
	}

	fields := map[string]interface{}{
		"up": upState,
	}

	m := &inputs.UpMeasurement{
		Name:     inputs.CollectorUpMeasurement,
		Tags:     tags,
		Fields:   fields,
		Election: ipt.Election,
	}

	pt := m.Point()
	// Merge with input tags
	for k, v := range ipt.Tags {
		pt.AddTag(k, v)
	}

	if err := ipt.Feeder.Feed(point.Metric,
		[]*point.Point{pt},
		dkio.WithCollectCost(time.Since(time.Now())),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(upFeedSource)); err != nil {
		l.Errorf("Feed up metric failed for %s: %s, ignored", clientURL, err.Error())
	}
}

func (ipt *Input) getUpInstance(clientURL string) string {
	uu, _ := url.Parse(clientURL)
	h, p, err := net.SplitHostPort(uu.Host)
	var host string
	var port int
	if err == nil {
		host = h
		port, _ = strconv.Atoi(p)
	} else {
		host = uu.Host
		l.Warnf("failed to split host and port: %s", err)
	}
	ipPort := fmt.Sprintf("%s:%d", host, port)
	return ipPort
}
