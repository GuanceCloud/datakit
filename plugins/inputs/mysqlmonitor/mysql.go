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

	_ "github.com/go-sql-driver/mysql"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l    *logger.Logger
	name = "mysqlMonitor"
)

func (_ *Mysql) Catalog() string {
	return "mysql"
}

func (_ *Mysql) SampleConfig() string {
	return configSample
}

func (_ *Mysql) Description() string {
	return ""
}

func (mysql *Mysql) Run() {
	l = logger.SLogger("mysqlMonitor")
	l.Info("mysqlMonitor input started...")

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysql.Username, mysql.Password, mysql.Host, mysql.Port, mysql.Database)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		l.Errorf("mysql connect faild %v", err)
	}

	mysql.db = db

	interval, err := time.ParseDuration(mysql.Interval)
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

func (mysql *Mysql) command() {
	for key, item := range metricMap {
		resMap, err := mysql.Query(item)
		if err != nil {
			l.Errorf("mysql query faild %v", err)
		}

		mysql.handleResponse(key, resMap)
	}
}

func (mysql *Mysql) handleResponse(m string, response []map[string]interface{}) error {
	for _, item := range response {
		tags := map[string]string{}

		tags["dbName"] = mysql.Database
		tags["instanceId"] = mysql.InstanceId
		tags["instanceDesc"] = mysql.InstanceDesc
		tags["server"] = mysql.Host
		tags["port"] = mysql.Port
		tags["product"] = mysql.Product
		tags["type"] = m

		pt, err := io.MakeMetric(mysql.MetricName, tags, item, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	return nil
}

func (r *Mysql) Query(sql string) ([]map[string]interface{}, error) {
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
	inputs.Add(name, func() inputs.Input {
		return &Mysql{}
	})
}
