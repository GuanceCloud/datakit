// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setIptCOStatus() {
	ipt.CollectCoStatus = "OK"
}

func (ipt *Input) setIptErrCOStatus() {
	ipt.CollectCoStatus = "NotOK"
}

func (ipt *Input) setIptLastCOInfo(m *customerObjectMeasurement) {
	ipt.LastCustomerObject = m
}

func (ipt *Input) setInptErrCOMsg(s string) {
	ipt.CollectCoErrMsg = s
}

func (ipt *Input) setIptLastCOInfoByErr() {
	ipt.LastCustomerObject.tags["reason"] = ipt.CollectCoErrMsg
	ipt.LastCustomerObject.tags["col_co_status"] = ipt.CollectCoStatus
}

func (ipt *Input) getCoPointByColErr() []*gcPoint.Point {
	ms := []inputs.MeasurementV2{}
	if ipt.LastCustomerObject == nil {
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf(inputName+"-%s:%d", ipt.Host, ipt.Port),
			"host":          ipt.Host,
			"ip":            fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
			"col_co_status": ipt.CollectCoStatus,
		}
		m := &customerObjectMeasurement{
			name:     "database",
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		ipt.setIptLastCOInfo(m)
		ms = append(ms, m)
	} else {
		ipt.setIptLastCOInfoByErr()
		ms = append(ms, ipt.LastCustomerObject)
	}
	pts := getPointsFromMeasurement(ms)

	return pts
}

func (ipt *Input) getVersionAndUptime() error {
	// Set Redis uptime and version
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
	info, err := ipt.client.Info(ctx, "server").Result()
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}
	ipt.Version = parseVersion(info)
	ipt.Uptime = parseUptime(info)
	return nil
}

func (ipt *Input) FeedCoErr(err error) {
	ipt.setInptErrCOMsg(err.Error())
	ipt.setIptErrCOStatus()
	pts := ipt.getCoPointByColErr()
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(ipt.start)),
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}
