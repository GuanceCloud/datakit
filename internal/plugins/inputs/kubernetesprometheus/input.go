// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetesprometheus wraps prometheus collect on kubernetes.
package kubernetesprometheus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Input struct {
	NodeLocal bool `toml:"node_local"`
	InstanceManager

	chPause chan bool
	pause   *atomic.Bool
	feeder  dkio.Feeder
	cancel  context.CancelFunc

	runOnce sync.Once
}

func (*Input) SampleConfig() string                    { return example }
func (*Input) Catalog() string                         { return inputName }
func (*Input) AvailableArchs() []string                { return []string{datakit.LabelK8s} }
func (*Input) Singleton()                              { /*nil*/ }
func (*Input) SampleMeasurement() []inputs.Measurement { return nil /* no measurement docs exported */ }
func (*Input) Terminate()                              { /* TODO */ }

func (ipt *Input) Run() {
	klog = logger.SLogger("kubernetesprometheus")

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for {
		// enable nodeLocal model or election success
		if ipt.NodeLocal || !ipt.pause.Load() {
			ipt.runOnce.Do(
				func() {
					managerGo.Go(func(_ context.Context) error {
						if err := ipt.start(); err != nil {
							klog.Warn(err)
						}
						return nil
					})
				})
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.stop()
			klog.Info("exit")
			return

		case pause := <-ipt.chPause:
			ipt.pause.Store(pause)

			// disable nodeLocal model and election defeat
			if !ipt.NodeLocal && pause {
				ipt.stop()
				ipt.runOnce = sync.Once{} // reset runOnce
			}

		case <-tick.C:
			// next
		}
	}
}

func (ipt *Input) start() error {
	klog.Info("start")

	apiClient, err := client.GetAPIClient()
	if err != nil {
		return err
	}

	nodeName, err := getLocalNodeName()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	ipt.cancel = cancel

	if ipt.NodeLocal {
		ctx = withNodeName(ctx, nodeName)
		ctx = withNodeLocal(ctx, ipt.NodeLocal)
	}
	ctx = withPause(ctx, ipt.pause)

	ipt.InstanceManager.Run(ctx, apiClient.Clientset, apiClient.InformerFactory, ipt.feeder)

	apiClient.InformerFactory.Start(ctx.Done())
	apiClient.InformerFactory.WaitForCacheSync(ctx.Done())

	<-ctx.Done()
	klog.Info("end")
	return nil
}

func (ipt *Input) stop() {
	if ipt.cancel != nil {
		ipt.cancel()
	}
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

func newPauseVar() *atomic.Bool {
	b := &atomic.Bool{}
	b.Store(true)
	return b
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			NodeLocal: true,
			chPause:   make(chan bool, inputs.ElectionPauseChannelLength),
			pause:     newPauseVar(),
			feeder:    dkio.DefaultFeeder(),
		}
	})
}
