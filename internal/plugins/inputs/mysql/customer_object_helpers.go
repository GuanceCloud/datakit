// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"fmt"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setIptCOStatus() {
	ipt.CollectCoStatus = "OK"
}

func (ipt *Input) setIptErrCOStatus() {
	ipt.CollectCoStatus = "NotOK"
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
		ipt.LastCustomerObject = m
		ms = append(ms, m)
	} else {
		ipt.setIptLastCOInfoByErr()
		ms = append(ms, ipt.LastCustomerObject)
	}
	pts := getPointsFromMeasurement(ms)
	return pts
}
