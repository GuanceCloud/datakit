package run

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	teleagent "github.com/influxdata/telegraf/agent"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
)

type Agent struct {
	Config     *config.Config
	outputsMgr *OutputsMgr
}

func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config:     config,
		outputsMgr: &OutputsMgr{},
	}
	return a, nil
}

func (a *Agent) Run() error {

	log.Printf("Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return err
	}

	a.outputsMgr.Outputs, err = a.Config.LoadOutputs()
	if err != nil {
		return err
	}

	var outputsNames []string
	for _, p := range a.outputsMgr.Outputs {
		outputsNames = append(outputsNames, p.Config.Name)
	}

	log.Printf("available outputs: %s", strings.Join(outputsNames, ","))

	log.Printf("Connecting outputs")
	err = a.outputsMgr.Init()
	if err != nil {
		return err
	}

	err = a.outputsMgr.ConnectOutputs(ctx)
	if err != nil {
		return err
	}

	inputC := make(chan telegraf.Metric, 100)

	startTime := time.Now()

	ncInput := len(a.Config.Inputs)

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

		log.Printf("D! Interval Inputs done")
	}(dst)

	if ncInput > 0 {

		go func() {
			select {
			case <-ctx.Done():
				a.stopServiceInputs()
				close(dst)
			}
		}()

		src = dst

		wg.Add(1)
		go func(src chan telegraf.Metric) {
			defer wg.Done()

			if err := a.outputsMgr.Start(src,
				startTime,
				a.Config.MainCfg.FlushInterval.Duration,
				a.Config.MainCfg.FlushJitter.Duration,
				a.Config.MainCfg.RoundInterval); err != nil {
				log.Printf("E! Error running outputs: %v", err)
			}

			a.stopDatacleanService() //dataclean后面停
		}(src)

		wg.Wait()

		a.outputsMgr.Close()
	} else {
		//no input no output
		<-ctx.Done()
	}

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
		if _, ok := input.Input.(telegraf.ServiceInput); ok {
			continue
		}
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

func (a *Agent) initPlugins() error {
	for _, input := range a.Config.Inputs {
		err := input.Init()
		if err != nil {
			return fmt.Errorf("could not initialize input %s: %v", input.LogName(), err)
		}
	}
	// for _, output := range a.Config.Outputs {
	// 	err := output.Init()
	// 	if err != nil {
	// 		return fmt.Errorf("could not initialize output %s: %v", output.Config.Name, err)
	// 	}
	// }
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
		if _, ok := input.Input.(telegraf.ServiceInput); !ok {
			continue
		}
		if input.Config.Name == "dataclean" {
			continue
		}
		log.Printf("D! stopping service input: %s", input.Config.Name)
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			si.Stop()
		}
	}
}

func (a *Agent) stopDatacleanService() {
	for _, input := range a.Config.Inputs {
		if _, ok := input.Input.(telegraf.ServiceInput); !ok {
			continue
		}
		if input.Config.Name != "dataclean" {
			continue
		}
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
