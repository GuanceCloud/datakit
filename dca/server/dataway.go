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
			doUploadHostStatus(sender, interval*4)
		}
	}
}

// doUploadHostStatus uploads the availability of every datakit. Rows that are
// not updated for staleAfter are reported as offline: a connection may be half
// open (no close frame, no read timeout on the peer) and must not keep a host
// online forever.
func doUploadHostStatus(sender *dataway.DialtestingSender, staleAfter time.Duration) {
	if sender == nil {
		l.Warnf("dataway sender is nil, skip uploading host status")
		return
	}
	// Judge all rows at snapshot time; earlier uploads may take longer than staleAfter.
	observedAt := time.Now()
	res := []*ws.DataKit{}
	query := "select * from datakit"

	if err := datakitDB.Select(
		query, &res,
	); err != nil {
		l.Errorf("failed to query datakit list : %s", err.Error())
		return
	}

	for _, dk := range currentRows(res) {
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
		if dk.Status == ws.StatusRunning && rowIsFresh(dk, observedAt, staleAfter) {
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

// rowIsFresh reports whether the datakit reported something recently enough to
// be trusted as "online". A non-positive staleAfter disables the check.
func rowIsFresh(dk *ws.DataKit, observedAt time.Time, staleAfter time.Duration) bool {
	if dk == nil {
		return false
	}

	if staleAfter <= 0 {
		return true
	}

	return observedAt.Sub(time.UnixMilli(dk.UpdatedAt)) <= staleAfter
}

// currentRows keeps only the newest row of every host, per workspace.
//
// The availability of a host must be reported once: rows left behind by an
// older conn id (or by a previous datakit process) are stale, and reporting
// them would mark a perfectly healthy host as offline.
//
// The workspace is part of the key: every workspace uploads through its own
// dataway, so a host of the same name in another workspace is another host.
func currentRows(rows []*ws.DataKit) []*ws.DataKit {
	latest := map[string]*ws.DataKit{}

	for _, row := range rows {
		if row == nil {
			continue
		}

		key := row.WorkspaceUUID + "/" + row.HostName

		cur, ok := latest[key]
		switch {
		case !ok:
		case row.UpdatedAt > cur.UpdatedAt:
		case row.UpdatedAt == cur.UpdatedAt && row.Status == ws.StatusRunning && cur.Status != ws.StatusRunning:
		default:
			continue
		}

		latest[key] = row
	}

	out := make([]*ws.DataKit, 0, len(latest))
	for _, row := range latest {
		out = append(out, row)
	}

	return out
}
