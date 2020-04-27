package mysqlmonitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type MysqlMonitor struct {
	Mysql       []*Mysql
	ctx         context.Context
	cancelFun   context.CancelFunc
	accumulator telegraf.Accumulator
	logger      *models.Logger

	runningInstances []*runningInstance
}

type runningInstance struct {
	cfg        *Mysql
	agent      *MysqlMonitor
	logger     *models.Logger
	db         *sql.DB
	metricName string
}

func (_ *MysqlMonitor) SampleConfig() string {
	return configSample
}

func (_ *MysqlMonitor) Description() string {
	return ""
}

func (_ *MysqlMonitor) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *MysqlMonitor) Init() error {
	return nil
}

func (o *MysqlMonitor) Start(acc telegraf.Accumulator) error {
	o.logger = &models.Logger{
		Name: `mysqlmonitor`,
	}

	if len(o.Mysql) == 0 {
		o.logger.Warnf("no configuration found")
		return nil
	}

	o.logger.Infof("starting...")

	o.accumulator = acc

	for _, instCfg := range o.Mysql {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  o,
			logger: o.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "mysql_monitor"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 5
		}

		connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", instCfg.Username, instCfg.Password, instCfg.Host, instCfg.Port, instCfg.Database)
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			r.logger.Errorf("mysql connect faild %v", err)
		}
		r.db = db

		o.runningInstances = append(o.runningInstances, r)

		go r.run(o.ctx)
	}
	return nil
}

func (o *MysqlMonitor) Stop() {
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
		// internal.SleepContext(ctx, 10*time.Second)
	}
}

func (r *runningInstance) command() {
	for key, item := range metricMap {
		resMap, err := r.Query(item)
		if err != nil {
			r.logger.Errorf("mysql query faild %v", err)
		}

		r.handleResponse(key, resMap)
	}
}

func (r *runningInstance) handleResponse(m string, response []map[string]interface{}) error {
	for _, item := range response {
		tags := map[string]string{}

		tags["db"] = r.cfg.Database
		tags["instanceId"] = r.cfg.InstanceId
		tags["host"] = r.cfg.Host
		tags["port"] = r.cfg.Port
		tags["product"] = "mysql"
		tags["type"] = m

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
	inputs.Add("mysqlmonitor", func() telegraf.Input {
		ac := &MysqlMonitor{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
