// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

var defDialtestingMock dialtestingMock = &prodDialtestingMock{}

type dialtestingMock interface {
	debugInit(*dt.HTTPTask) error
	debugRun(*dt.HTTPTask) error
	getResults(*dt.HTTPTask) (tags map[string]string, fields map[string]interface{})
}

type prodDialtestingMock struct{}

func (*prodDialtestingMock) debugInit(task *dt.HTTPTask) error {
	return task.InitDebug()
}

func (*prodDialtestingMock) debugRun(task *dt.HTTPTask) error {
	return task.Run()
}

func (*prodDialtestingMock) getResults(task *dt.HTTPTask) (tags map[string]string, fields map[string]interface{}) {
	return task.GetResults()
}
