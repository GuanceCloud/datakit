package run

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/file"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/http"
)

const (
	outputChannelSize = 100
)

type outputChannel struct {
	name    string
	outputs []*models.RunningOutput
	ch      chan telegraf.Metric
}

type outputsMgr struct {
	outputChannels map[string]*outputChannel
}

func (om *outputsMgr) findMetricChannel(name string) *outputChannel {
	if oc, ok := om.outputChannels[name]; ok {
		return oc
	}
	return om.outputChannels["*"]
}

func newHttpOutput(name string, catalog string, c *config.Config, ftdataway string, maxPostInterval time.Duration) (*models.RunningOutput, error) {
	httpOutput := http.NewHttpOutput()
	httpOutput.Catalog = catalog
	if httpOutput.Headers == nil {
		httpOutput.Headers = map[string]string{}
	}
	httpOutput.Headers[`X-Datakit-UUID`] = c.MainCfg.UUID
	httpOutput.Headers[`X-Version`] = git.Version
	httpOutput.Headers[`User-Agent`] = config.DKUserAgent
	if maxPostInterval > 0 {
		httpOutput.Headers[`X-Max-POST-Interval`] = internal.IntervalString(maxPostInterval)
	}
	httpOutput.ContentEncoding = "gzip"
	httpOutput.URL = ftdataway
	return newRunningOutput(name, catalog, httpOutput)
}

func newFileOutput(name string, catalog string, outputFile string) (*models.RunningOutput, error) {
	fileOutput := file.NewFileOutput()
	fileOutput.Catalog = catalog
	fileOutput.Files = []string{outputFile}
	return newRunningOutput(name, catalog, fileOutput)
}

func (om *outputsMgr) LoadOutputs(cfg *config.Config) error {

	globalChannel := &outputChannel{
		name: "*",
		ch:   make(chan telegraf.Metric, outputChannelSize),
	}

	if cfg.MainCfg.OutputsFile != "" {
		if ro, err := newFileOutput("file", "", cfg.MainCfg.OutputsFile); err != nil {
			return err
		} else {
			globalChannel.outputs = append(globalChannel.outputs, ro)
		}
	}

	if cfg.MainCfg.DataWay != nil {
		if ro, err := newHttpOutput("http", "", cfg, cfg.MainCfg.DataWayRequestURL, config.MaxLifeCheckInterval); err != nil {
			return err
		} else {
			globalChannel.outputs = append(globalChannel.outputs, ro)
		}
	}
	om.outputChannels[globalChannel.name] = globalChannel

	for _, input := range cfg.Inputs {

		var oc *outputChannel

		catalog := ""
		if cip, ok := input.Input.(inputs.Input); ok {
			catalog = cip.Catalog()
		}

		if input.Config.DataWayRequestURL != "" {

			if ro, err := newHttpOutput(input.Config.Name, catalog, cfg, input.Config.DataWayRequestURL, 0); err != nil {
				return err
			} else {
				oc = &outputChannel{
					name:    ro.Config.Name,
					outputs: []*models.RunningOutput{ro},
					ch:      make(chan telegraf.Metric, outputChannelSize),
				}
			}
		}

		if input.Config.OutputFile != "" {

			if ro, err := newFileOutput(input.Config.Name, catalog, input.Config.OutputFile); err != nil {
				return err
			} else {
				if oc != nil {
					oc.outputs = append(oc.outputs, ro)
				} else {
					oc = &outputChannel{
						name:    ro.Config.Name,
						outputs: []*models.RunningOutput{ro},
						ch:      make(chan telegraf.Metric, outputChannelSize),
					}
				}

			}
		}

		if oc != nil {
			om.outputChannels[input.Config.Name] = oc
		}

	}

	return nil
}

func (om *outputsMgr) ConnectOutputs(ctx context.Context) error {
	for _, oc := range om.outputChannels {
		for _, op := range oc.outputs {
			log.Printf("D! Attempting connection to [%s]", op.LogName())
			err := op.Output.Connect()
			if err != nil {
				log.Printf("E! Failed to connect to [%s], retrying in 15s, "+
					"error was '%s'", op.LogName(), err)

				err := internal.SleepContext(ctx, 15*time.Second)
				if err != nil {
					return err
				}

				err = op.Output.Connect()
				if err != nil {
					return err
				}
			}
			log.Printf("D! Successfully connected to [%s]", op.LogName())
		}
	}
	return nil
}

func (om *outputsMgr) Start(startTime time.Time, interval time.Duration, jitter time.Duration, roundInterval bool) error {

	var wg sync.WaitGroup
	flushCtx, flushCancel := context.WithCancel(context.Background())

	//各个output定时刷新
	for _, oc := range om.outputChannels {
		for _, op := range oc.outputs {
			// Overwrite agent flush_interval if this plugin has its own.
			if op.Config.FlushInterval != 0 {
				interval = op.Config.FlushInterval
			}

			wg.Add(1)
			go func(output *models.RunningOutput) {
				defer wg.Done()

				if roundInterval {
					err := internal.SleepContext(flushCtx, internal.AlignDuration(startTime, interval))
					if err != nil {
						return
					}
				}

				om.flush(flushCtx, output, interval, jitter)
			}(op)
		}
	}

	//将不断接收到的metric添加到output中
	var chWg sync.WaitGroup
	for _, oc := range om.outputChannels {
		chWg.Add(1)
		go func(oc *outputChannel) {
			defer chWg.Done()
			for metric := range oc.ch {
				for _, op := range oc.outputs {
					op.AddMetric(metric)
				}
			}
		}(oc)
	}
	chWg.Wait()
	log.Printf("I! Hang on, flushing any cached metrics before shutdown")

	flushCancel()
	wg.Wait()

	return nil
}

func (om *outputsMgr) Stop() {
	for _, oc := range om.outputChannels {
		if oc.ch != nil {
			close(oc.ch)
		}
	}
}

func (om *outputsMgr) Close() {
	for _, oc := range om.outputChannels {
		for _, op := range oc.outputs {
			err := op.Output.Close()
			if err != nil {
				log.Printf("E! closing output: %v", err)
			}
		}
	}
}

func (ro *outputsMgr) flush(
	ctx context.Context,
	output *models.RunningOutput,
	interval time.Duration,
	jitter time.Duration,
) {
	// since we are watching two channels we need a ticker with the jitter
	// integrated.
	ticker := NewTicker(interval, jitter)
	defer ticker.Stop()

	logError := func(err error) {
		if err != nil {
			log.Printf("E! Error writing to %s: %v", output.LogName(), err)
		}
	}

	for {
		// Favor shutdown over other methods.
		select {
		case <-ctx.Done():
			logError(ro.flushOnce(output, interval, output.Write))
			return
		default:
		}

		select {
		case <-ticker.C:
			logError(ro.flushOnce(output, interval, output.Write))
		case <-output.BatchReady:
			// Favor the ticker over batch ready
			select {
			case <-ticker.C:
				logError(ro.flushOnce(output, interval, output.Write))
			default:
				logError(ro.flushOnce(output, interval, output.WriteBatch))
			}
		case <-ctx.Done():
			logError(ro.flushOnce(output, interval, output.Write))
			return
		}
	}
}

func (ro *outputsMgr) flushOnce(
	output *models.RunningOutput,
	timeout time.Duration,
	writeFunc func() error,
) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := make(chan error)
	go func() {
		done <- writeFunc()
	}()

	for {
		select {
		case err := <-done:
			output.LogBufferStatus()
			return err
		case <-ticker.C:
			log.Printf("W! [%q] did not complete within its flush interval", output.LogName())
			output.LogBufferStatus()
		}
	}

}

func newRunningOutput(name string, catalog string, output telegraf.Output) (*models.RunningOutput, error) {

	switch t := output.(type) {
	case serializers.SerializerOutput:
		c := &serializers.Config{}
		c.DataFormat = "influx"

		serializer, err := serializers.NewSerializer(c)
		if err != nil {
			return nil, err
		}

		t.SetSerializer(serializer)
	}

	outputConfig := &models.OutputConfig{
		Name: name,
	}

	ro := models.NewRunningOutput(name, output, outputConfig, 0, 0)
	return ro, nil
}
