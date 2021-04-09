package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/godror/godror"
	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"golang.org/x/net/context/ctxhttp"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	fInterval        = flag.String("interval", "5m", "gather interval")
	fDataType        = flag.String("data-type", "metric", "data type, metric/logging, default metric")
	fMetric          = flag.String("metric-name", "oracle_monitor", "gathered metric name")
	fInstanceDesc    = flag.String("instance-desc", "", "oracle description")
	fHost            = flag.String("host", "", "oracle host")
	fPort            = flag.String("port", "1521", "oracle port")
	fUsername        = flag.String("username", "", "oracle username")
	fPassword        = flag.String("password", "", "oracle password")
	fServiceName     = flag.String("service-name", "", "oracle service name")
	fTags            = flag.String("tags", "", `additional tags in 'a=b,c=d,...' format`)
	fDatakitHTTPPort = flag.Int("datakit-http-port", 9529, "DataKit HTTP server port")

	fLog      = flag.String("log", filepath.Join(datakit.InstallDir, "externals", "oraclemonitor.log"), "log path")
	fLogLevel = flag.String("log-level", "info", "log file")

	l              *logger.Logger
	datakitPostURL = ""
)

type monitor struct {
	libPath       string
	metric        string
	interval      string
	instanceId    string
	user          string
	password      string
	desc          string
	host          string
	port          string
	serviceName   string
	clusterType   string
	tags          map[string]string
	oracleVersion string

	db               *sql.DB
	intervalDuration time.Duration
}

func buildMonitor() *monitor {
	m := &monitor{
		metric:      *fMetric,
		interval:    *fInterval,
		user:        *fUsername,
		password:    *fPassword,
		desc:        *fInstanceDesc,
		host:        *fHost,
		port:        *fPort,
		serviceName: *fServiceName,
	}

	if m.interval != "" {
		du, err := time.ParseDuration(m.interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", m.interval, err.Error())
			m.intervalDuration = 10 * time.Minute
		} else {
			m.intervalDuration = du
		}
	}

	for {
		db, err := sql.Open("godror", fmt.Sprintf("%s/%s@%s:%s/%s", m.user, m.password, m.host, m.port, m.serviceName))
		if err == nil {
			m.db = db
			break
		}

		l.Errorf("oracle connect faild %v, retry each 3 seconds...", err)
		time.Sleep(time.Second * 3)
		continue
	}

	return m
}

func main() {
	flag.Parse()

	datakitPostURL = fmt.Sprintf("http://0.0.0.0:%d/v1/write/metric?name=oraclemonitor", *fDatakitHTTPPort)

	logger.SetGlobalRootLogger(*fLog, *fLogLevel, logger.OPT_DEFAULT)
	if *fInstanceDesc != "" { // add description to logger
		l = logger.SLogger("oracle-" + *fInstanceDesc)
	} else {
		l = logger.SLogger("oracle")
	}

	m := buildMonitor()
	m.run()
}

func (m *monitor) run() {

	l.Info("start oraclemonitor...")

	tick := time.NewTicker(m.intervalDuration)
	defer tick.Stop()
	defer m.db.Close()

	wg := sync.WaitGroup{}

	for {
		select {
		case <-tick.C:
			for idx, _ := range execCfgs {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					m.handle(execCfgs[i])
				}(idx)
			}

			wg.Wait() // blocking
		}
	}
}

func (m *monitor) handle(ec *ExecCfg) {
	res, err := m.query(ec)
	if err != nil {
		l.Warnf("oracle query `%s' faild: %v, ignored", ec.sql, err)
		return
	}

	l.Debugf("get %d result from metric %s", len(res), ec.metricName)

	if res == nil {
		return
	}

	_ = handleResponse(m, ec.metricName, ec.tagsMap, res)
}

func handleResponse(m *monitor, metricName string, tagsKeys []string, response []map[string]interface{}) error {
	lines := [][]byte{}

	for _, item := range response {

		tags := map[string]string{}

		tags["oracle_service"] = m.serviceName
		tags["oracle_server"] = fmt.Spprintf("%s:%s", m.host, m.port)
		tags["host"] = m.host

		for _, tagKey := range tagsKeys {
			tags[tagKey] = String(item[tagKey])
			delete(item, tagKey)
		}

		// add user-added tags
		// XXX: this may overwrite tags within @tags
		for k, v := range m.tags {
			tags[k] = v
		}

		pt, err := ifxcli.NewPoint(metricName, tags, item, time.Now())
		if err != nil {
			l.Error("NewPoint(): %s", err.Error())
			return err
		}

		fmt.Println("point ======>", pt.String())

		lines = append(lines, []byte(pt.String()))
	}

	if len(lines) == 0 {
		l.Debugf("no metric collected on %s", metricName)
		return nil
	}

	// io 输出
	if err := WriteData(bytes.Join(lines, []byte("\n")), datakitPostURL); err != nil {
		return err
	}

	return nil
}

func (m *monitor) query(obj *ExecCfg) ([]map[string]interface{}, error) {
	rows, err := m.db.Query(obj.sql)
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
					str := strings.TrimSpace(val.(string))
					data, err := strconv.ParseFloat(str, 64)
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

// String converts <i> to string.
func String(i interface{}) string {
	if i == nil {
		return ""
	}
	switch value := i.(type) {
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(int64(value), 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(uint64(value), 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	case []rune:
		return string(value)
	default:
		// Finally we use json.Marshal to convert.
		jsonContent, _ := json.Marshal(value)
		return string(jsonContent)
	}
}

func WriteData(data []byte, urlPath string) error {
	// dataway path
	ctx, _ := context.WithCancel(context.Background())
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))

	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post to %s ok", urlPath)
		return nil
	default:
		l.Errorf("post to %s failed(HTTP: %d): %s", urlPath, resp.StatusCode, string(body))
		return fmt.Errorf("post datakit failed")
	}
	return nil
}

const (
	oracle_process_sql = `
    SELECT PROGRAM, PGA_USED_MEM, PGA_ALLOC_MEM, PGA_FREEABLE_MEM, PGA_MAX_MEM FROM GV$PROCESS
    `

	oracle_system_sql = `
    SELECT VALUE, METRIC_NAME FROM GV$SYSMETRIC ORDER BY BEGIN_TIME
    `

	oracle_tablespace_sql = `
    SELECT
	  m.tablespace_name,
	  NVL(m.used_space * t.block_size, 0),
	  m.tablespace_size * t.block_size,
	  NVL(m.used_percent, 0),
	  NVL2(m.used_space, 0, 1)
	FROM
	  dba_tablespace_usage_metrics m
	  join dba_tablespaces t on m.tablespace_name = t.tablespace_name
    `
)

type ExecCfg struct {
	sql        string
	metricName string
	tagsMap    []string
}

var execCfgs = []*ExecCfg{
	&ExecCfg{
		sql:        oracle_process_sql,
		metricName: "oracle_process",
		tagsMap:    []string{"program"},
	},
	&ExecCfg{
		sql:        oracle_system_sql,
		metricName: "oracle_system",
		tagsMap:    []string{},
	},
	&ExecCfg{
		sql:        oracle_tablespace_sql,
		metricName: "oracle_tablespace",
		tagsMap:    []string{"tablespace"},
	},
}
