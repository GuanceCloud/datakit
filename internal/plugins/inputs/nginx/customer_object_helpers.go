// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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
		host := strings.TrimPrefix(ipt.host, "http://")
		portDisplay := fmt.Sprintf("%d", ipt.Ports[0])
		if ipt.Ports[0] != ipt.Ports[1] {
			portDisplay = fmt.Sprintf("%d-%d", ipt.Ports[0], ipt.Ports[1])
		}

		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%s", host, portDisplay),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf("%s-%s:%s", inputName, host, portDisplay),
			"host":          ipt.host,
			"ip":            fmt.Sprintf("%s:%d", host, ipt.Ports[0]),
			"col_co_status": ipt.CollectCoStatus,
		}
		m := &customerObjectMeasurement{
			name:     "web_server",
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

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func (ipt *Input) getVersionAndUptime(port int) error {
	u := ipt.host + ":" + strconv.Itoa(port) + ipt.path
	resp, err := ipt.client.Get(u)
	if err != nil {
		l.Errorf("%s", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Errorf("failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return err
	}
	// 获取 NGINX 版本信息
	serverHeader := resp.Header.Get("Server")
	if serverHeader != "" {
		parts := strings.Split(serverHeader, "/")
		if len(parts) == 2 {
			ipt.Version = parts[1]
		} else {
			l.Errorf("unexpected server header format: %s", serverHeader)
			return fmt.Errorf("the server isn't nginx: %s", serverHeader)
		}
	} else {
		l.Errorf("server header not found")
		return fmt.Errorf("the server isn't nginx: %s", serverHeader)
	}
	return nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}
	host := strings.TrimPrefix(ipt.host, "http://")
	portDisplay := fmt.Sprintf("%d", ipt.Ports[0])
	if ipt.Ports[0] != ipt.Ports[1] {
		portDisplay = fmt.Sprintf("%d-%d", ipt.Ports[0], ipt.Ports[1])
	}

	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%s", host, portDisplay),
		"version":      ipt.Version,
		"uptime":       fmt.Sprintf("%d", ipt.Uptime),
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s:%s", inputName, host, portDisplay),
		"host":          host,
		"ip":            fmt.Sprintf("%s:%d", host, ipt.Ports[0]),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:     "web_server",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ipt.setIptLastCOInfo(m)
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) FeedCoPts() {
	err := ipt.getVersionAndUptime(ipt.Ports[0])
	if err != nil {
		ipt.FeedCoByErr(err)
		return
	}
	pts, _ := ipt.collectCustomerObjectMeasurement()
	if err := ipt.feeder.Feed(gcPoint.CustomObject, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithSource(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		l.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByErr(err error) {
	tn := time.Now()
	ipt.setInptErrCOMsg(err.Error())
	ipt.setIptErrCOStatus()
	pts := ipt.getCoPointByColErr()
	if err := ipt.feeder.Feed(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(tn)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(customObjectFeedName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}
