// +build linux,amd64

package oraclemonitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"database/sql"

	_ "github.com/godror/godror"
	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	//_ "gopkg.in/rana/ora.v4"
)

type OracleMonitor struct {
	Oracle      []*Oracle
	ctx         context.Context
	cancelFun   context.CancelFunc
	accumulator telegraf.Accumulator
	logger      *models.Logger

	runningInstances []*runningInstance
}

type runningInstance struct {
	cfg        *Oracle
	agent      *OracleMonitor
	logger     *models.Logger
	db         *sql.DB
	metricName string
}

func (_ *OracleMonitor) SampleConfig() string {
	return configSample
}

func (_ *OracleMonitor) Description() string {
	return ""
}

func (_ *OracleMonitor) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *OracleMonitor) Init() error {
	return nil
}

func (o *OracleMonitor) Start(acc telegraf.Accumulator) error {
	o.logger = &models.Logger{
		Name: `oraclemonitor`,
	}

	if len(o.Oracle) == 0 {
		o.logger.Warnf("no configuration found")
		return nil
	}

	o.logger.Infof("starting...")

	o.accumulator = acc

	for _, instCfg := range o.Oracle {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  o,
			logger: o.logger,
		}

		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "oracle_monitor"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 5
		}

		connStr := fmt.Sprintf("%s/%s@%s/%s", instCfg.Username, instCfg.Password, instCfg.Host, instCfg.Server)
		db, err := sql.Open("godror", connStr)
		if err != nil {
			r.logger.Errorf("oracle connect faild %v", err)
		}
		r.db = db

		o.runningInstances = append(o.runningInstances, r)

		go r.run(o.ctx)
	}
	return nil
}

func (o *OracleMonitor) Stop() {
	o.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		go r.command()

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) command() {
	for key, item := range metricMap {
		resMap, err := r.Query(item)
		if err != nil {
			r.logger.Errorf("oracle connect faild %v", err)
		}

		r.handleResponse(key, resMap)
	}
}

func (r *runningInstance) handleResponse(m string, response []map[string]interface{}) error {
	for _, item := range response {
		tags := map[string]string{}

		tags["oracle_server"] = r.cfg.Server
		tags["oracle_port"] = r.cfg.Port
		tags["instance_id"] = r.cfg.InstanceId
		tags["instance_desc"] = r.cfg.InstanceDesc
		tags["product"] = "oracle"
		tags["host"] = r.cfg.Host
		tags["type"] = m

		if tagKeys, ok := tagsMap[m]; ok {
			for _, tagKey := range tagKeys {
				tags[tagKey] = item[tagKey].(string)
			}
		}

		r.agent.accumulator.AddFields(r.metricName, item, tags)
	}

	return nil
}

func (r *runningInstance) Query(sql string) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for idx, _ := range cache {
		var a interface{}
		cache[idx] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		_ = rows.Scan(cache...)

		item := make(map[string]interface{})
		for i, data := range cache {
			key := strings.ToLower(columns[i])
			item[key] = *data.(*interface{})
		}
		list = append(list, item)
	}
	return list, nil
}

func init() {
	inputs.Add("oraclemonitor", func() telegraf.Input {
		ac := &OracleMonitor{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
