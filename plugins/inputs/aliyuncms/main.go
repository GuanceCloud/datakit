// +build ignore

package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"

	teleagent "github.com/influxdata/telegraf/agent"
	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/file"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/run"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyuncms"
)

var (
	flagCfgFile    = flag.String("cfg", ``, `configure file`)
	flagOutputFile = flag.String("output", ``, `output metric file`)
	flagDataway    = flag.String("dataway", ``, `url of dataway`)

	GlobalTags = map[string]string{}

	inputName = "aliyuncms"
)

type Source struct {
	url        string
	template   string
	dataFormat string
	tags       []string
	fields     []string
}

func main() {

	flag.Parse()

	input, err := loadInput(*flagCfgFile)
	if err != nil {
		os.Exit(1)
	}

	outputs, err := loadOutputs()
	if err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("input: %s", input.Config.Name)

	inputC := make(chan telegraf.Metric, 100)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case sig := <-signals:
			_ = sig
			// if sig == syscall.SIGHUP {
			// 	log.Printf("I! Reloading config")
			// }
			cancel()
		}
	}()

	log.Printf("Connecting outputs")
	if err := connectOutputs(ctx, outputs); err != nil {
		log.Fatalf("%s", err)
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		runServiceInputs(ctx, []*models.RunningInput{input}, inputC)
		time.Sleep(20 * time.Millisecond)
		close(inputC)
		log.Printf("D! Input channel closed")
	}()

	wg.Add(1)
	go func(src chan telegraf.Metric) {
		defer wg.Done()
		runOutputs(outputs, time.Now(), src, false)
	}(inputC)

	wg.Wait()

	closeOutputs(outputs)

	log.Printf("%s stopped successfully", inputName)

}

func runServiceInputs(
	ctx context.Context,
	inputs []*models.RunningInput,
	dst chan<- telegraf.Metric,
) error {

	started := []telegraf.ServiceInput{}

	for _, input := range inputs {
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

	select {
	case <-ctx.Done():
		for _, svr := range started {
			svr.Stop()
		}
	}

	return nil
}

func runOutputs(
	outputs []*models.RunningOutput,
	startTime time.Time,
	src <-chan telegraf.Metric,
	roundInterval bool,
) error {
	interval := time.Second * 10
	jitter := time.Duration(0)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for _, output := range outputs {

		interval := interval
		// if output.Config.FlushInterval != 0 {
		// 	interval = output.Config.FlushInterval
		// }

		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()

			if roundInterval {
				err := internal.SleepContext(
					ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			flush(ctx, output, interval, jitter)
		}(output)
	}

	for metric := range src {
		for i, output := range outputs {
			if i == len(outputs)-1 {
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

func connectOutputs(ctx context.Context, outputs []*models.RunningOutput) error {
	for _, output := range outputs {
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

func closeOutputs(outputs []*models.RunningOutput) {
	for _, output := range outputs {
		log.Printf("D! closing output: %s", output.Config.Name)
		output.Close()
	}
}

func flush(
	ctx context.Context,
	output *models.RunningOutput,
	interval time.Duration,
	jitter time.Duration,
) {
	// since we are watching two channels we need a ticker with the jitter
	// integrated.
	ticker := run.NewTicker(interval, jitter)
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
			logError(flushOnce(output, interval, output.Write))
			return
		default:
		}

		select {
		case <-ticker.C:
			logError(flushOnce(output, interval, output.Write))
		case <-output.BatchReady:
			// Favor the ticker over batch ready
			select {
			case <-ticker.C:
				logError(flushOnce(output, interval, output.Write))
			default:
				logError(flushOnce(output, interval, output.WriteBatch))
			}
		case <-ctx.Done():
			logError(flushOnce(output, interval, output.Write))
			return
		}
	}
}

func flushOnce(
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
			//output.LogBufferStatus()
			return err
		case <-ticker.C:
			log.Printf("W! [%q] did not complete within its flush interval", output.LogName())
			//output.LogBufferStatus()
		}
	}

}

func loadInput(cfgpath string) (*models.RunningInput, error) {

	data, err := ioutil.ReadFile(cfgpath)
	if err != nil {
		log.Printf("Error loading config file %s, %s", cfgpath, err)
		return nil, err
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		log.Printf("Error parsing config file %s, %s", cfgpath, err)
		return nil, err
	}
	if tbl != nil {
		delete(tbl.Fields, "interval")
	}

	input := aliyuncms.NewAgent()

	if err := toml.UnmarshalTable(tbl, input); err != nil {
		log.Printf("Error parsing config file %s, %s", cfgpath, err)
		return nil, err
	}

	pluginConfig := &models.InputConfig{Name: inputName}
	rp := models.NewRunningInput(input, pluginConfig)
	rp.SetDefaultTags(GlobalTags)

	return rp, nil
}

func buildSerializer(format string) (serializers.Serializer, error) {
	c := &serializers.Config{TimestampUnits: time.Duration(1 * time.Second)}
	c.DataFormat = format // "influx"
	return serializers.NewSerializer(c)
}

func genRunningOutput(name string, output telegraf.Output) (*models.RunningOutput, error) {

	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializer("influx")
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

func loadOutputs() ([]*models.RunningOutput, error) {

	outputs := []*models.RunningOutput{}

	fileOutput := file.NewFileOutput()
	fileOutput.Files = []string{"stdout"}

	ro, err := genRunningOutput("file", fileOutput)
	if err != nil {
		return nil, err
	}

	outputs = append(outputs, ro)

	return outputs, nil
}
