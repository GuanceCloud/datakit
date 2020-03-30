package run

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	teleagent "github.com/influxdata/telegraf/agent"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type Agent struct {
	Config *config.Config
}

func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}
	return a, nil
}

func (a *Agent) Run(ctx context.Context) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	log.Printf("Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return err
	}

	log.Printf("Connecting outputs")
	err = a.connectOutputs(ctx)
	if err != nil {
		return err
	}

	inputC := make(chan telegraf.Metric, 100)
	//procC := make(chan telegraf.Metric, 100)
	//outputC := make(chan telegraf.Metric, 100)

	startTime := time.Now()

	err = a.startServiceInputs(ctx, inputC)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	src := inputC
	dst := inputC

	wg.Add(1)
	go func(dst chan telegraf.Metric) {
		defer wg.Done()

		err := a.runInputs(ctx, startTime, dst)
		if err != nil {
			log.Printf("E! Error running inputs: %v", err)
		}

		a.stopServiceInputs()

		close(dst)
		log.Printf("D! Input channel closed")
	}(dst)

	src = dst

	wg.Add(1)
	go func(src chan telegraf.Metric) {
		defer wg.Done()

		err := a.runOutputs(startTime, src)
		if err != nil {
			log.Printf("E! Error running outputs: %v", err)
		}
	}(src)

	wg.Wait()

	a.closeOutputs()

	log.Printf("datakit stopped successfully")
	return nil
}

func (a *Agent) runInputs(
	ctx context.Context,
	startTime time.Time,
	dst chan<- telegraf.Metric,
) error {
	var wg sync.WaitGroup
	for _, input := range a.Config.Inputs {
		interval := a.Config.MainCfg.Interval.Duration
		jitter := time.Duration(0) //a.Config.Agent.CollectionJitter.Duration

		// Overwrite agent interval if this plugin has its own.
		if input.Config.Interval != 0 {
			interval = input.Config.Interval
		}

		acc := teleagent.NewAccumulator(input, dst)
		acc.SetPrecision(a.Precision())

		wg.Add(1)
		go func(input *models.RunningInput) {
			defer wg.Done()

			if a.Config.MainCfg.RoundInterval {
				err := internal.SleepContext(
					ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			a.gatherOnInterval(ctx, acc, input, interval, jitter)
		}(input)
	}
	wg.Wait()

	return nil
}

func (a *Agent) runOutputs(
	startTime time.Time,
	src <-chan telegraf.Metric,
) error {
	interval := a.Config.MainCfg.FlushInterval.Duration
	jitter := time.Duration(0) // a.Config.MainCfg.FlushJitter.Duration

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for _, output := range a.Config.Outputs {
		interval := interval
		// Overwrite agent flush_interval if this plugin has its own.
		if output.Config.FlushInterval != 0 {
			interval = output.Config.FlushInterval
		}

		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()

			if a.Config.MainCfg.RoundInterval {
				err := internal.SleepContext(
					ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			a.flush(ctx, output, interval, jitter)
		}(output)
	}

	for metric := range src {
		for i, output := range a.Config.Outputs {
			if i == len(a.Config.Outputs)-1 {
				output.AddMetric(metric)
			} else {
				output.AddMetric(metric.Copy())
			}
		}
	}

	log.Println("I! Hang on, flushing any cached metrics before shutdown")
	cancel()
	wg.Wait()

	return nil
}

func (a *Agent) flush(
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
			logError(a.flushOnce(output, interval, output.Write))
			return
		default:
		}

		select {
		case <-ticker.C:
			logError(a.flushOnce(output, interval, output.Write))
		case <-output.BatchReady:
			// Favor the ticker over batch ready
			select {
			case <-ticker.C:
				logError(a.flushOnce(output, interval, output.Write))
			default:
				logError(a.flushOnce(output, interval, output.WriteBatch))
			}
		case <-ctx.Done():
			logError(a.flushOnce(output, interval, output.Write))
			return
		}
	}
}

func (a *Agent) flushOnce(
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
			log.Printf("W! [%q] did not complete within its flush interval",
				output.LogName())
			output.LogBufferStatus()
		}
	}

}

func (a *Agent) initPlugins() error {
	for _, input := range a.Config.Inputs {
		err := input.Init()
		if err != nil {
			return fmt.Errorf("could not initialize input %s: %v", input.LogName(), err)
		}
	}
	for _, output := range a.Config.Outputs {
		err := output.Init()
		if err != nil {
			return fmt.Errorf("could not initialize output %s: %v", output.Config.Name, err)
		}
	}
	return nil
}

func (a *Agent) closeOutputs() {
	for _, output := range a.Config.Outputs {
		log.Printf("D! closing output: %s", output.Config.Name)
		output.Close()
	}
}

func (a *Agent) connectOutputs(ctx context.Context) error {
	for _, output := range a.Config.Outputs {
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

func (a *Agent) startServiceInputs(
	ctx context.Context,
	dst chan<- telegraf.Metric,
) error {
	started := []telegraf.ServiceInput{}

	for _, input := range a.Config.Inputs {
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			// Service input plugins are not subject to timestamp rounding.
			// This only applies to the accumulator passed to Start(), the
			// Gather() accumulator does apply rounding according to the
			// precision agent setting.
			log.Printf("D! starting service input: %s", input.Config.Name)
			acc := teleagent.NewAccumulator(input, dst)
			acc.SetPrecision(time.Nanosecond)

			err := si.Start(acc)
			if err != nil {
				log.Printf("E! Service for [%s] failed to start: %v",
					input.LogName(), err)

				for _, si := range started {
					si.Stop()
				}

				return err
			}

			started = append(started, si)
		}
	}

	return nil
}

func (a *Agent) stopServiceInputs() {
	for _, input := range a.Config.Inputs {
		log.Printf("D! stopping service input: %s", input.Config.Name)
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			si.Stop()
		}
	}
}

func (a *Agent) Precision() time.Duration {
	// precision := a.Config.Agent.Precision.Duration
	// interval := a.Config.Agent.Interval.Duration

	// if precision > 0 {
	// 	return precision
	// }

	// switch {
	// case interval >= time.Second:
	// 	return time.Second
	// case interval >= time.Millisecond:
	// 	return time.Millisecond
	// case interval >= time.Microsecond:
	// 	return time.Microsecond
	// default:
	// 	return time.Nanosecond
	// }

	return time.Nanosecond

}

// panicRecover displays an error if an input panics.
func panicRecover(input *models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("E! FATAL: [%s] panicked: %s, Stack:\n%s",
			input.LogName(), err, trace)
		log.Println("E! PLEASE REPORT THIS PANIC ON GITHUB with " +
			"stack trace, configuration, and OS information: " +
			"https://github.com/influxdata/telegraf/issues/new/choose")
	}
}

func (a *Agent) gatherOnInterval(
	ctx context.Context,
	acc telegraf.Accumulator,
	input *models.RunningInput,
	interval time.Duration,
	jitter time.Duration,
) {
	defer panicRecover(input)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		err := internal.SleepContext(ctx, internal.RandomDuration(jitter))
		if err != nil {
			return
		}

		err = a.gatherOnce(acc, input, interval)
		if err != nil {
			acc.AddError(err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) gatherOnce(
	acc telegraf.Accumulator,
	input *models.RunningInput,
	timeout time.Duration,
) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := make(chan error)
	go func() {
		done <- input.Gather(acc)
	}()

	for {
		select {
		case err := <-done:
			return err
		case <-ticker.C:
			log.Printf("W! [%s] did not complete within its interval",
				input.LogName())
		}
	}
}
