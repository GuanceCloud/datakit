// Package sqlserver collects SQL Server metrics.
package sqlserver

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipeline,
	}
	return pipelineMap
}

func (n *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:   inputName,
			Service:  inputName,
			Pipeline: n.Log.Pipeline,
		},
	}
}

func (n *Input) initDB() error {
	db, err := sql.Open("sqlserver", fmt.Sprintf("sqlserver://%s:%s@%s?dial+timeout=3", n.User, n.Password, n.Host))
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	n.db = db
	return nil
}

func (n *Input) RunPipeline() {
	if n.Log == nil || len(n.Log.Files) == 0 {
		return
	}

	if n.Log.Pipeline == "" {
		n.Log.Pipeline = "sqlserver.p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		GlobalTags:        n.Tags,
		IgnoreStatus:      n.Log.IgnoreStatus,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilineMatch:    `^\d{4}-\d{2}-\d{2}`,
	}

	pl, err := config.GetPipelinePath(n.Log.Pipeline)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	n.tail, err = tailer.NewTailer(n.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go n.tail.Start()
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("sqlserver start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	// Init DB until OK.
	for {
		if err := n.initDB(); err != nil {
			l.Errorf("initDB: %s", err.Error())

			io.FeedLastError(inputName, err.Error())
		} else {
			break
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("sqlserver exit")
			return
		}
	}

	defer func() {
		if err := n.db.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}

		if n.tail != nil {
			n.tail.Close()
		}
	}()

	for {
		n.getMetric()
		if len(collectCache) > 0 {
			err := io.Feed(inputName, datakit.Metric, collectCache, &io.Option{CollectCost: time.Since(n.start)})
			collectCache = collectCache[:0]
			if err != nil {
				n.lastErr = err
				l.Errorf(err.Error())
				continue
			}
		}

		if n.lastErr != nil {
			io.FeedLastError(inputName, n.lastErr.Error())
			n.lastErr = nil
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("sqlserver exit")
			return

		case <-n.semStop.Wait():
			n.exit()
			l.Info("sqlserver return")
			return

		}
	}
}

func (n *Input) exit() {
	if n.tail != nil {
		n.tail.Close()
		l.Info("sqlserver log exit")
	}
}

func (n *Input) Terminate() {
	if n.semStop != nil {
		n.semStop.Close()
	}
}

func (n *Input) getMetric() {
	start := time.Now()
	n.start = start
	for _, v := range query {
		n.handRow(v, start)
	}
}

func (n *Input) handRow(query string, ts time.Time) {
	rows, err := n.db.Query(query)
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}
	defer rows.Close() //nolint:errcheck

	if err := rows.Err(); err != nil {
		l.Errorf("rows.Err: %s", err)
		return
	}

	OrderedColumns, err := rows.Columns()
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}

	for rows.Next() {
		var columnVars []interface{}
		// var fields = make(map[string]interface{})
		// store the column name with its *interface{}
		columnMap := make(map[string]*interface{})

		for _, column := range OrderedColumns {
			columnMap[column] = new(interface{})
		}
		// populate the array of interface{} with the pointers in the right order
		for i := 0; i < len(columnMap); i++ {
			columnVars = append(columnVars, columnMap[OrderedColumns[i]])
		}
		// deconstruct array of variables and send to Scan
		err := rows.Scan(columnVars...)
		if err != nil {
			l.Error(err.Error())
			n.lastErr = err
			return
		}
		measurement := ""
		tags := make(map[string]string)
		for k, v := range n.Tags {
			tags[k] = v
		}
		fields := make(map[string]interface{})
		for header, val := range columnMap {
			if str, ok := (*val).(string); ok {
				if header == "measurement" {
					measurement = str
					continue
				}
				tags[header] = strings.TrimSuffix(str, "\\")
			} else {
				if *val == nil {
					continue
				}
				fields[header] = *val
			}
		}
		if len(fields) == 0 {
			continue
		}

		point, err := io.MakePoint(measurement, tags, fields, ts)
		if err != nil {
			l.Errorf("make point err:%s", err.Error())
			n.lastErr = err
			continue
		}
		collectCache = append(collectCache, point)
	}
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ServerProperties{},
		&Performance{},
		&WaitStatsCategorized{},
		&DatabaseIO{},
		&Schedulers{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},

			semStop: cliutils.NewSem(),
		}
		return s
	})
}
