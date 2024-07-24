// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetesprometheus wraps prometheus collect on kubernetes.
package kubernetesprometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Input struct {
	InstanceManager

	chPause chan bool
	pause   bool
	feeder  dkio.Feeder

	cancel context.CancelFunc
}

func (*Input) SampleConfig() string                    { return example }
func (*Input) Catalog() string                         { return inputName }
func (*Input) AvailableArchs() []string                { return []string{datakit.LabelK8s} }
func (*Input) Singleton()                              { /*nil*/ }
func (*Input) SampleMeasurement() []inputs.Measurement { return nil /* no measurement docs exported */ }
func (*Input) Terminate()                              { /* TODO */ }

func (ipt *Input) Run() {
	klog = logger.SLogger("kubernetesprometheus")

	for {
		if !ipt.pause {
			if err := ipt.start(); err != nil {
				klog.Warn(err)
			}
		} else {
			ipt.stop()
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.stop()
			klog.Info("exit")
			return

		case ipt.pause = <-ipt.chPause:
			// nil
		}
	}
}

func (ipt *Input) start() error {
	klog.Info("start")

	apiClient, err := client.GetAPIClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	ipt.cancel = cancel

	ipt.InstanceManager.Run(ctx, apiClient.Clientset, apiClient.InformerFactory, ipt.feeder)

	apiClient.InformerFactory.Start(ctx.Done())
	apiClient.InformerFactory.WaitForCacheSync(ctx.Done())

	<-ctx.Done()
	return nil
}

func (ipt *Input) stop() {
	ipt.cancel()
	klog.Info("stop")
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			chPause: make(chan bool, inputs.ElectionPauseChannelLength),
			feeder:  dkio.DefaultFeeder(),
		}
	})
}
