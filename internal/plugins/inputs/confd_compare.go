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
	name        string
	insertIndex int // the index begin insert input, -1: need not insert
	deleteIndex int // the index begin delete input, -1: need not delete
}

// CompareInputs Sub-category comparison collector.
func CompareInputs(confdInputs map[string][]*ConfdInfo, defaultEnabledInputs []string) {
	handleList := make([]handle, 0)

	// Make sure confdInputs have all InputsInfo inputs kind
	for inputsName := range InputsInfo {
		if _, ok := confdInputs[inputsName]; !ok {
			confdInputs[inputsName] = []*ConfdInfo{}
			l.Debugf("confdInputs add non-datakit-input-kind %s", inputsName)
		}
	}

	// Traverse the collector types, len(confdInputs) must be >= len(Inputs)
	for confdInputName, confdConfigs := range confdInputs {
		handleList = append(handleList, handle{confdInputName, -1, -1})

		// Make sure InputsInfo have this collector kind
		var inputsInfo []*inputInfo
		var ok bool
		if inputsInfo, ok = InputsInfo[confdInputName]; !ok {
			InputsInfo[confdInputName] = []*inputInfo{}
			l.Debugf("confd add non-datakit-input-kind %s", confdInputName)
		}

		forLen := len(inputsInfo)
		if forLen > len(confdConfigs) {
			forLen = len(confdConfigs)
		}

		// Compare one kind inputs
		// Find index if need modify
		for i := 0; i < forLen; i++ {
			jsonConfd, _ := json.Marshal(confdConfigs[i].Input)
			jsonInput, _ := json.Marshal(InputsInfo[confdInputName][i].input)

			jsonConfdStr := string(jsonConfd)
			jsonInputStr := string(jsonInput)
			_ = jsonConfdStr
			_ = jsonInputStr

			if string(jsonConfd) != string(jsonInput) {
				l.Info("input configInfo modified, reset, inputName: ", confdInputName, ", index: ", i)
				// from here insert(append) by confd
				handleList[len(handleList)-1].insertIndex = i
				// from here delete, first delete then append so =i. If singleton len(confdConfigs) must be 1
				handleList[len(handleList)-1].deleteIndex = i
				break
			}
		}

		// Find index if need only insert
		if handleList[len(handleList)-1].insertIndex == -1 {
			// If insertIndex > -1 , need not these code
			if len(confdConfigs) > len(inputsInfo) {
				// confd is long
				handleList[len(handleList)-1].insertIndex = len(inputsInfo) // from here insert(append) by confd
			}
		}

		// Find index if need only delete. Be careful some input is default enabled and must singleton
		if handleList[len(handleList)-1].insertIndex == -1 {
			// If insertIndex > -1 , need not these code
			if len(confdConfigs) < len(inputsInfo) {
				// old inputsInfo is long

				// Some input is default enabled and must singleton
				// Default start collector + self
				i := 0
				for i = 0; i < len(defaultEnabledInputs); i++ {
					if confdInputName == defaultEnabledInputs[i] {
						break
					}
				}
				if i >= len(defaultEnabledInputs) && confdInputName != "self" {
					// Not default enabled input
					handleList[len(handleList)-1].deleteIndex = len(confdConfigs) // from here delete
				} else if len(InputsInfo[confdInputName]) > 1 {
					// Default enabled input and must singleton and len >1
					handleList[len(handleList)-1].deleteIndex = 1 // singleton
				}
			}
		}
	}

	// Start execute, use context timeout to prevent failure and cause memory leaks
	ctxNew, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if errs := handleInput(confdInputs, handleList, ctxNew); len(errs) != 0 {
		l.Errorf("modifyInput failed. err: %v", errs)
	}
}
