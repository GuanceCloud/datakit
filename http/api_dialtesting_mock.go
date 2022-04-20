package http

import (
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

var defDialtestingMock dialtestingMock = &prodDialtestingMock{}

type dialtestingMock interface {
	debugInit(dt.Task) error
	debugRun(dt.Task) error
	getResults(dt.Task) (tags map[string]string, fields map[string]interface{})
}

type prodDialtestingMock struct{}

func (*prodDialtestingMock) debugInit(task dt.Task) error {
	return task.InitDebug()
}

func (*prodDialtestingMock) debugRun(task dt.Task) error {
	return task.Run()
}

func (*prodDialtestingMock) getResults(task dt.Task) (tags map[string]string, fields map[string]interface{}) {
	return task.GetResults()
}
