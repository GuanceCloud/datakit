// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package inputs manage all input's interfaces.
package inputs

import (
	"context"
	"encoding/json"
	"time"
)

type ConfdInfo struct {
	Input Input
}
type handle struct {
	name       string
	handleType handleType // 0: add, 1: delete, 2: change
	index      int
}

type handleType int

const (
	ADD    handleType = 0
	DELETE handleType = 1
	MODIFY handleType = 2
)

// CompareInputs Sub-category comparison collector.
func CompareInputs(confdInputs map[string][]*ConfdInfo) {
	handleList := make([]handle, 0)

	// Traverse the collector types, len(confdInputs) must be >= len(Inputs)
	for confdInputName, confdConfigs := range confdInputs {
		// Processing within a single collector kind
		inputsInfoLen := 0
		inputsInfo, ok := InputsInfo[confdInputName]
		if ok {
			inputsInfoLen = len(inputsInfo)
		}
		forLen := inputsInfoLen
		if inputsInfoLen > len(confdConfigs) {
			forLen = len(confdConfigs)
		}

		// region Forward comparison, traverse inputsInfo, and discover modifications and deletions
		for i := 0; i < forLen; i++ {
			jsonConfd, _ := json.Marshal(confdConfigs[i].Input)
			jsonInput, _ := json.Marshal(InputsInfo[confdInputName][i].input)

			if string(jsonConfd) != string(jsonInput) {
				l.Info("input configInfo modified, reset, inputName: ", confdInputName, ", index: ", i)
				handleList = append(handleList, handle{confdInputName, MODIFY, i})
			}
		}

		// Find possible collectors that need to be deleted (reverse order, deleteInput needs to be deleted in reverse order)
		if inputsInfoLen > len(confdConfigs) {
			for i := inputsInfoLen - 1; i >= forLen; i-- {
				// Reverse order, delete from the back, to avoid possible misalignment.
				l.Info("input configInfo delete, inputName: ", confdInputName, ", index: ", i)
				handleList = append(handleList, handle{confdInputName, DELETE, i})
			}
		}
		// endregion

		// region Reverse comparison, traverse the redundant confdConfigs, you can find the increase.
		if inputsInfoLen < len(confdConfigs) {
			for i := forLen; i < len(confdConfigs); i++ {
				l.Info("input configInfo creat new, inputName: ", confdInputName, ", index: ", i)
				handleList = append(handleList, handle{confdInputName, ADD, i})
			}
		}
		// endregion
	}

	if len(handleList) == 0 {
		return
	}

	// Start execute, use context timeout to prevent failure and cause memory leaks
	ctxNew, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if errs := handleInput(confdInputs, handleList, ctxNew); len(errs) != 0 {
		l.Errorf("modifyInput failed. err: %v", errs)
	}
}
