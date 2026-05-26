// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"encoding/json"
	"net/url"
	"path"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	uploadMetricName   = "dk_host_available"
)

func UploadHostStatus(interval time.Duration, closeCh <-chan struct{}) {
	sender := &dataway.DialtestingSender{}
	if err := sender.Init(&dataway.DialtestingSenderOpt{
		HTTPTimeout: defaultHTTPTimeout,
	}); err != nil {
		l.Errorf("init dataway sender failed, err: %s", err)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-closeCh:
			l.Infof("close upload host status ticker")
			return
		case <-ticker.C:
			// send host status
			doUploadHostStatus(sender)
		}
	}
}

func doUploadHostStatus(sender *dataway.DialtestingSender) {
	if sender == nil {
		l.Warnf("dataway sender is nil, skip uploading host status")
		return
	}
	res := []*ws.DataKit{}
	query := "select * from datakit"

	if err := datakitDB.Select(
		query, &res,
	); err != nil {
		l.Errorf("failed to query datakit list : %s", err.Error())
		return
	}

	for _, dk := range res {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(dk.Config), &config); err != nil {
			l.Warnf("parse datakit config failed: %s", err.Error())
			continue
		}

		if config == nil {
			l.Warnf("datakit config is empty")
			continue
		}

		datawayURL := ""
		if value, ok := config["dataway_url"]; ok {
			if text, isValid := value.(string); isValid {
				datawayURL = text
			}
		}

		if datawayURL == "" {
			l.Warnf("empty dataway url")
			continue
		}

		status := "offline"
		if dk.Status == ws.StatusRunning {
			status = "online"
		}

		opts := point.DefaultMetricOptions()
		var kvs point.KVs
		kvs = kvs.AddTag("host", dk.HostName)
		kvs = kvs.AddTag("host_ip", dk.IP)
		kvs = kvs.AddTag("status", status)
		kvs = kvs.Add("value", 1)
		pt := point.NewPoint(uploadMetricName, kvs, opts...)

		u, err := url.Parse(datawayURL)
		if err != nil {
			l.Warnf("get invalid url, ignored: %s", err.Error())
			continue
		}
		u.Path = path.Join(u.Path, datakit.Metric)
		urlStr := u.String()
		if err := sender.WriteData(urlStr, []*point.Point{pt}); err != nil {
			l.Errorf("failed to upload host status: %s", err.Error())
		}
	}
}
