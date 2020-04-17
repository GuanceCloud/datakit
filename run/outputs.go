package run

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type OutputsMgr struct {
	Outputs []*models.RunningOutput
}

func (r *OutputsMgr) Init() error {
	for _, o := range r.Outputs {
		if p, ok := o.Output.(telegraf.Initializer); ok {
			err := p.Init()
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (ro *OutputsMgr) ConnectOutputs(ctx context.Context) error {
	for _, output := range ro.Outputs {
		log.Printf("D! Attempting connection to [%s]", output.LogName())
		err := output.Output.Connect()
		if err != nil {
			log.Printf("E! Failed to connect to [%s], retrying in 15s, "+
				"error was '%s'", output.LogName(), err)

			err := internal.SleepContext(ctx, 15*time.Second)
			if err != nil {
				return err
			}

			err = output.Output.Connect()
			if err != nil {
				return err
			}
		}
		log.Printf("D! Successfully connected to [%s]", output.LogName())
	}
	return nil
}

func (ro *OutputsMgr) Start(src chan telegraf.Metric, startTime time.Time, interval time.Duration, jitter time.Duration, roundInterval bool) error {

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for _, output := range ro.Outputs {
		// Overwrite agent flush_interval if this plugin has its own.
		if output.Config.FlushInterval != 0 {
			interval = output.Config.FlushInterval
		}

		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()

			if roundInterval {
				err := internal.SleepContext(ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			ro.flush(ctx, output, interval, jitter)
		}(output)
	}

	for metric := range src {

		for i, output := range ro.Outputs {
			if i == len(ro.Outputs)-1 {
				output.AddMetric(metric)
			} else {
				output.AddMetric(metric.Copy())
			}
		}
	}

	log.Printf("I! Hang on, flushing any cached metrics before shutdown")
	cancel()
	wg.Wait()

	return nil
}

func (r *OutputsMgr) Close() {
	for _, o := range r.Outputs {
		err := o.Output.Close()
		if err != nil {
			log.Printf("E! closing output: %v", err)
		}
	}
}

func (ro *OutputsMgr) flush(
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

func (ro *OutputsMgr) flushOnce(
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
