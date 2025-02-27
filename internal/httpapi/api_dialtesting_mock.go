// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"fmt"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
)

var defDialtestingMock dialtestingMock = &prodDialtestingMock{}

type dialtestingMock interface {
	debugInit(dt.ITask, map[string]dt.Variable) error
	debugRun(dt.ITask) error
	getResults(dt.ITask) (tags map[string]string, fields map[string]interface{})
	getVars(dt.ITask) dt.Vars
}

type prodDialtestingMock struct{}

func (*prodDialtestingMock) debugInit(task dt.ITask, variables map[string]dt.Variable) error {
	if err := task.CheckTask(); err != nil {
		return fmt.Errorf("task check failed: %w", err)
	}
	return task.RenderTemplateAndInit(variables)
}

func (*prodDialtestingMock) debugRun(task dt.ITask) error {
	return task.Run()
}

func (*prodDialtestingMock) getResults(task dt.ITask) (tags map[string]string, fields map[string]interface{}) {
	return task.GetResults()
}

func (*prodDialtestingMock) getVars(task dt.ITask) dt.Vars {
	return task.GetPostScriptVars()
}
