package mysqlmonitor

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l *zap.SugaredLogger
)

type MysqlMonitor struct {
	cfg *Mysql
	db  *sql.DB
}

func (_ *MysqlMonitor) Catalog() string {
	return "mysql"
}

func (_ *MysqlMonitor) SampleConfig() string {
	return configSample
}

func (_ *MysqlMonitor) Description() string {
	return ""
}

func (mysql *MysqlMonitor) Run() {
	l = logger.SLogger("harborMonitor")
	l.Info("harborMonitor input started...")

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysql.cfg.Username, mysql.cfg.Password, mysql.cfg.Host, mysql.cfg.Port, mysql.cfg.Database)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		l.Errorf("mysql connect faild %v", err)
	}

	mysql.db = db

	interval, err := time.ParseDuration(mysql.cfg.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			mysql.command()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (mysql *MysqlMonitor) command() {
	for key, item := range metricMap {
		resMap, err := mysql.Query(item)
		if err != nil {
			l.Errorf("mysql query faild %v", err)
		}

		mysql.handleResponse(key, resMap)
	}
}

func (mysql *MysqlMonitor) handleResponse(m string, response []map[string]interface{}) error {
	for _, item := range response {
		tags := map[string]string{}

		tags["dbName"] = mysql.cfg.Database
		tags["instanceId"] = mysql.cfg.InstanceId
		tags["instanceDesc"] = mysql.cfg.InstanceDesc
		tags["server"] = mysql.cfg.Host
		tags["port"] = mysql.cfg.Port
		tags["product"] = mysql.cfg.Product
		tags["type"] = m

		pt, err := io.MakeMetric(mysql.cfg.MetricName, tags, item, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		err = io.Feed([]byte(pt), io.Metric)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	return nil
}

func (r *MysqlMonitor) Query(sql string) ([]map[string]interface{}, error) {
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
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				switch vType.String() {
				case "int64":
					item[key] = val.(int64)
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					item[key] = val.(time.Time)
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					return nil, fmt.Errorf("unsupport data type '%s' now\n", vType)
				}
			}
		}

		list = append(list, item)
	}
	return list, nil
}

func init() {
	inputs.Add("mysqlmonitor", func() inputs.Input {
		return &MysqlMonitor{}
	})
}
