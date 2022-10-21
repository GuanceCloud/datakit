// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package inputs manage all input's interfaces.
package inputs

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

func handleInput(confdInputs map[string][]*ConfdInfo, handleList []handle, ctx context.Context) (errs []error) {
	round := 0 // Loop index, exit after the steps slice is consumed
	for {
		select {
		case <-ctx.Done():
			tip := "confd handleInput timeout"
			l.Error(tip)
			errs = append(errs, fmt.Errorf(tip))
			return errs

		default:
			switch round {
			case 0:
				// 停止
				l.Info("before confd stopInput")

				stopInput(handleList, &errs)

			case 1:
				// 追加
				l.Info("before confd addInputs")

				addInputs(confdInputs, handleList, &errs)

			case 2:
				// 修改
				l.Info("before confd addInputs")

				modifyInput(confdInputs, handleList, &errs)

			case 3:
				l.Info("before set pipelines")

				plscript.LoadAllScripts2StoreFromPlStructPath(plscript.GitRepoScriptNS,
					filepath.Join(datakit.GitReposRepoFullPath, "pipeline"))

			case 4:
				// start and restart
				l.Info("before confd runInput")

				runInput(handleList, &errs)

			case 5:
				// delete
				l.Info("before confd deleteInput")

				deleteInput(handleList, &errs)
			}
		}

		round++
		if round > 6 {
			return errs
		}
	}
}

// stopInput stop collector List.
func stopInput(handles []handle, errs *[]error) {
	mtx.RLock()
	defer mtx.RUnlock()

	for _, h := range handles {
		if h.handleType == ADD {
			continue
		} // 0: add, 1: delete, 2: change

		// Make sure have this collector
		if _, ok := InputsInfo[h.name]; !ok {
			l.Debugf("confd stop skip non-datakit-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd stop skip non-datakit-input-kind %s", h.name))
			continue
		}
		if h.index >= len(InputsInfo[h.name]) {
			l.Debugf("confd stop datakit-input h.index out of bounds  %s, %d", h.name, h.index)
			*errs = append(*errs, fmt.Errorf("confd stop datakit-input h.index out of bounds %s, %d", h.name, h.index))
			continue
		}

		ii := InputsInfo[h.name][h.index]
		if ii.input == nil {
			l.Debugf("confd stop skip datakit-input is nil %s, %d", h.name, h.index)
			continue
		}

		if inp, ok := ii.input.(InputV2); ok {
			inp.Terminate()
		}
	}
}

// Add collector list.
func addInputs(confdInputs map[string][]*ConfdInfo, handles []handle, errs *[]error) {
	mtx.Lock()
	defer mtx.Unlock()

	for _, h := range handles {
		if h.handleType != ADD {
			continue
		} // 0: add, 1: delete, 2: change

		// Make sure you have this collector type
		if _, ok := InputsInfo[h.name]; !ok {
			InputsInfo[h.name] = []*inputInfo{}
			l.Debugf("confd add non-datakit-input-kind %s", h.name)
		}
		// Make sure confd has this collector
		if _, ok := confdInputs[h.name]; !ok {
			l.Debugf("confd add skip non-confd-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd add skip non-confd-input-kind %s", h.name))
			continue
		}

		// The new serial number should be at the end of InputsInfo[h.name]
		if h.index != len(InputsInfo[h.name]) {
			l.Debugf("confd add confd-input h.index wrong  %s, %d", h.name, h.index)
			*errs = append(*errs, fmt.Errorf("confd add confd-input h.index wrong %s, %d, %d. ", h.name, h.index, len(InputsInfo[h.name])))
			continue
		}

		InputsInfo[h.name] = append(InputsInfo[h.name], &inputInfo{confdInputs[h.name][h.index].Input})
	}
}

// Modify the Collector List.
func modifyInput(confdInputs map[string][]*ConfdInfo, handles []handle, errs *[]error) {
	mtx.Lock()
	defer mtx.Unlock()

	for _, h := range handles {
		if h.handleType != MODIFY {
			continue
		} // 0: add, 1: delete, 2: change

		// Make sure have this collector
		if _, ok := InputsInfo[h.name]; !ok {
			l.Debugf("confd modify skip non-datakit-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd modify skip non-datakit-input-kind %s", h.name))
			continue
		}
		if h.index >= len(InputsInfo[h.name]) {
			l.Debugf("confd modify datakit-input h.index out of bounds  %s, %d", h.name, h.index)
			*errs = append(*errs, fmt.Errorf("confd modify datakit-input h.index out of bounds  %s, %d", h.name, h.index))
			continue
		}

		// Make sure confd has this collector
		if _, ok := confdInputs[h.name]; !ok {
			l.Debugf("confd modify skip non-confd-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd modify skip non-confd-input-kind %s", h.name))
			continue
		}

		// The original location is a new collector instance.
		InputsInfo[h.name][h.index].input = confdInputs[h.name][h.index].Input
	}
}

// Start the collector.
func runInput(handles []handle, errs *[]error) {
	mtx.RLock()
	defer mtx.RUnlock()

	g := datakit.G("inputs")

	for _, h := range handles {
		if h.handleType == DELETE {
			continue
		} // 0: add, 1: delete, 2: change

		// Make sure have this collector
		if _, ok := InputsInfo[h.name]; !ok {
			l.Debugf("confd run skip non-datakit-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd run skip non-datakit-input-kind %s", h.name))
			continue
		}
		if h.index >= len(InputsInfo[h.name]) {
			l.Debugf("confd run datakit-input h.index out of bounds  %s, %d", h.name, h.index)
			*errs = append(*errs, fmt.Errorf("confd run datakit-input h.index out of bounds  %s, %d, %d. ",
				h.name, h.index, len(InputsInfo[h.name])))
			continue
		}

		ii := InputsInfo[h.name][h.index].input
		if ii == nil {
			l.Debugf("confd run skip non-datakit-input id nil %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd run skip non-datakit-input id nil %s", h.name))
		}

		if inp, ok := ii.(HTTPInput); ok {
			inp.RegHTTPHandler()
		}

		if inp, ok := ii.(PipelineInput); ok {
			inp.RunPipeline()
		}

		func(name string, ii *inputInfo) {
			g.Go(func(ctx context.Context) error {
				// NOTE: 让每个采集器间歇运行，防止每个采集器扎堆启动，导致主机资源消耗出现规律性的峰值
				time.Sleep(time.Duration(rand.Int63n(int64(10 * time.Second)))) //nolint:gosec
				l.Infof("starting input %s ...", name)

				protectRunningInput(name, ii)
				l.Infof("input %s exited", name)
				return nil
			})
		}(h.name, &inputInfo{ii})
	}
}

// Remove the collector.
func deleteInput(handles []handle, errs *[]error) {
	mtx.Lock()
	defer mtx.Unlock()

	// []handle is in reverse order
	for i := 0; i < len(handles); i++ {
		h := handles[i]

		if h.handleType != DELETE {
			continue
		} // 0: add, 1: delete, 2: change

		// Make sure have this collector
		if _, ok := InputsInfo[h.name]; !ok {
			l.Debugf("confd delete non-datakit-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd delete skip non-datakit-input-kind %s", h.name))
			continue
		}
		if h.index >= len(InputsInfo[h.name]) {
			l.Debugf("confd delete datakit-input h.index out of bounds  %s, %d", h.name, h.index)
			*errs = append(*errs, fmt.Errorf("confd delete datakit-input h.index out of bounds %s, %d", h.name, h.index))
			continue
		}

		InputsInfo[h.name] = append(InputsInfo[h.name][:h.index], InputsInfo[h.name][h.index+1:]...)

		// If this kind collector is empty, delete the map
		if len(InputsInfo[h.name]) == 0 {
			delete(InputsInfo, h.name)
		}
	}
}
