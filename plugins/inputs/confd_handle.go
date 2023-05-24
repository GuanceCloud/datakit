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
				// Stop
				l.Info("before confd stopInput")

				stopInput(handleList, &errs)
			case 1:
				// Delete
				l.Info("before confd deleteInput")

				deleteInput(handleList, &errs)
			case 2:
				// Add
				l.Info("before confd addInputs")

				addInputs(confdInputs, handleList, &errs)

			case 3:
				l.Info("before set pipelines")

				plscript.LoadAllScripts2StoreFromPlStructPath(plscript.GitRepoScriptNS,
					filepath.Join(datakit.GitReposRepoFullPath, "pipeline"))
			}
		}

		round++
		if round > 6 {
			mtx.Lock()
			defer mtx.Unlock()

			// If this kind inputs is empty, delete the map
			for name, inputInfo := range InputsInfo {
				if len(inputInfo) == 0 {
					delete(InputsInfo, name)
				}
			}

			return errs
		}
	}
}

// Stop collector List.
func stopInput(handles []handle, errs *[]error) {
	mtx.RLock()
	defer mtx.RUnlock()

	// Stop when insertIndex and deleteIndex any not be -1
	for _, h := range handles {
		if h.deleteIndex == -1 && h.insertIndex == -1 {
			continue
		}

		// Make sure have this input kind
		if _, ok := InputsInfo[h.name]; !ok {
			l.Debugf("confd stop skip non-datakit-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd stop skip non-datakit-input-kind %s", h.name))
			continue
		}

		// Walk stop all this kind inputs
		for i := 0; i < len(InputsInfo[h.name]); i++ {
			ii := InputsInfo[h.name][i]
			if ii.input == nil {
				l.Debugf("confd stop skip datakit-input is nil %s, %d", h.name, i)
				continue
			}

			if v2, ok := ii.input.(InputV2); ok {
				v2.Terminate()
			}
		}
	}
}

// Delete the input kind.
func deleteInput(handles []handle, errs *[]error) {
	mtx.Lock()
	defer mtx.Unlock()

	// Delete this input kind when insertIndex and deleteIndex any not be -1
	for _, h := range handles {
		if h.deleteIndex == -1 && h.insertIndex == -1 {
			continue
		}

		// Make sure have this input kind
		if _, ok := InputsInfo[h.name]; !ok {
			l.Debugf("confd delete skip non-datakit-input-kind %s", h.name)
			*errs = append(*errs, fmt.Errorf("confd delete skip non-datakit-input-kind %s", h.name))
			continue
		}

		// Delete this input kind
		delete(InputsInfo, h.name)
	}
}

// Add collector list.
func addInputs(confdInputs map[string][]*ConfdInfo, handles []handle, errs *[]error) {
	envs := getEnvs()
	mtx.Lock()
	defer mtx.Unlock()
	g := datakit.G("confd_inputs")

	// Recreate this input kind, and append form confdInputs,
	// When insertIndex and deleteIndex any not be -1
	for _, h := range handles {
		if h.insertIndex == -1 {
			continue
		}

		// Make sure you have this input type
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

		// Append all confd data
		for i := 0; i < len(confdInputs[h.name]); i++ {
			newInput := &inputInfo{confdInputs[h.name][i].Input}

			if inp, ok := newInput.input.(HTTPInput); ok {
				inp.RegHTTPHandler()
			}

			if inp, ok := newInput.input.(PipelineInput); ok {
				inp.RunPipeline()
			}

			if inp, ok := newInput.input.(ReadEnv); ok && datakit.Docker {
				inp.ReadEnv(envs)
			}

			InputsInfo[h.name] = append(InputsInfo[h.name], newInput)

			func(name string, ii *inputInfo) {
				g.Go(func(ctx context.Context) error {
					// NOTE: 让每个采集器间歇运行，防止每个采集器扎堆启动，导致主机资源消耗出现规律性的峰值
					time.Sleep(time.Duration(rand.Int63n(int64(10 * time.Second)))) //nolint:gosec
					l.Infof("starting input %s ...", name)

					protectRunningInput(name, ii)
					l.Infof("input %s exited", name)
					return nil
				})
			}(h.name, newInput)
		}
	}
}
