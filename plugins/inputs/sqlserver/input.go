package sqlserver

import (
	"database/sql"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"strings"
	"time"
)

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return catalogName
}

func (_ *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (_ *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipeline,
	}
	return pipelineMap
}

func (n *Input) initDb() error {
	db, err := sql.Open("sqlserver", fmt.Sprintf("sqlserver://%s:%s@%s", n.User, n.Password, n.Host))
	if err != nil {
		return err
	}
	n.db = db
	return nil
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("sqlserver start")
	n.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	if n.Log != nil {
		go func() {
			inputs.JoinPipelinePath(n.Log, "sqlserver.p")
			n.Log.Source = inputName
			n.Log.Match = `^\d{4}-\d{2}-\d{2}`
			n.Log.FromBeginning = true
			n.Log.Tags = map[string]string{}
			for k, v := range n.Tags {
				n.Log.Tags[k] = v
			}
			tail, err := inputs.NewTailer(n.Log)
			if err != nil {
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			n.tail = tail
			tail.Run()
		}()
	}

	if err := n.initDb(); err != nil {
		l.Error(err.Error())
		io.FeedLastError(inputName, n.lastErr.Error())
		return
	}

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
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
		case <-datakit.Exit.Wait():
			if n.tail != nil {
				n.tail.Close()
				l.Info("sqlserver log exit")
			}
			l.Info("sqlserver exit")
			return
		}
	}
}

func (n *Input) getMetric() {
	start := time.Now()
	n.start = start
	n.wg.Add(len(query))
	for _, v := range query {
		go func(q string, ts time.Time) {
			defer n.wg.Done()
			n.handRow(q, ts)
		}(v, start)
	}
	n.wg.Wait()
}

func (n *Input) handRow(query string, ts time.Time) {
	rows, err := n.db.Query(query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	OrderedColumns, err := rows.Columns()
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	for rows.Next() {
		var columnVars []interface{}
		//var fields = make(map[string]interface{})

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
			return
		}
		measurement := ""
		var tags = make(map[string]string)
		var fields = make(map[string]interface{})
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
			return
		}

		point, err := io.MakePoint(measurement, tags, fields, ts)
		if err != nil {
			l.Errorf("make point err:%s", err.Error())
			n.lastErr = err
			continue
		}
		metricAppend(point)
	}
	defer rows.Close()
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Performance{},
		&WaitStatsCategorized{},
		&DatabaseIO{},
		&ServerProperties{},
		&Schedulers{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
		return s
	})
}
