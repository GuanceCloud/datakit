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
