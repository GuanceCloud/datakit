package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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
	"github.com/jessevdk/go-flags"
	"golang.org/x/net/context/ctxhttp"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type Option struct {
	Interval        string `long:"interval" description:"gather interval" default:"10s"`
	Metric          string `long:"metric-name" description:"gathered metric name" default:"oracle_monitor"`
	InstanceDesc    string `long:"instance-desc" description:"oracle description"`
	Host            string `long:"host" description:"oracle host"`
	Port            string `long:"port" description:"oracle port" default:"1521"`
	Username        string `long:"username" description:"oracle username"`
	Password        string `long:"password" description:"oracle password"`
	ServiceName     string `long:"service-name" description:"oracle service name"`
	Tags            string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`
	DatakitHTTPPort int    `long:"datakit-http-port" description:"DataKit HTTP server port" default:"9529"`

	Log      string   `long:"log" description:"log path"`
	LogLevel string   `long:"log-level" description:"log file" default:"info"`
	Query    []string `long:"query" description:"custom query arrary"`
}

var (
	opt            Option
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
		metric:      opt.Metric,
		interval:    opt.Interval,
		user:        opt.Username,
		password:    opt.Password,
		desc:        opt.InstanceDesc,
		host:        opt.Host,
		port:        opt.Port,
		serviceName: opt.ServiceName,
		tags:        make(map[string]string),
	}

	items := strings.Split(opt.Tags, ";")
	for _, item := range items {
		tagArr := strings.Split(item, "=")

		if len(tagArr) == 2 {
			tagKey := strings.Trim(tagArr[0], " ")
			tagVal := strings.Trim(tagArr[1], " ")
			if tagKey != "" {
				m.tags[tagKey] = tagVal
			}
		}
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

	for _, query := range opt.Query {
		l.Info("custom query ======>", query)
		arr := strings.Split(query, ":")
		customCfg := &ExecCfg{
			sql:        arr[0],
			metricName: arr[1],
			tagsMap:    strings.Split(arr[2], ","),
		}
		execCfgs = append(execCfgs, customCfg)
	}

	return m
}

func main() {
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if opt.Log == "" {
		opt.Log = filepath.Join(datakit.InstallDir, "externals", "oracle.log")
	}

	datakitPostURL = fmt.Sprintf("http://0.0.0.0:%d/v1/write/metric?input=oracle", opt.DatakitHTTPPort)

	if err := logger.SetGlobalRootLogger(opt.Log, opt.LogLevel, logger.OPT_DEFAULT); err != nil {
		l.Errorf("set root log faile: %s", err.Error())
	}

	if opt.InstanceDesc != "" { // add description to logger
		l = logger.SLogger("oracle-" + opt.InstanceDesc)
	} else {
		l = logger.SLogger("oracle")
	}

	m := buildMonitor()
	m.run()
}

func (m *monitor) run() {
	l.Info("start oracle...")

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
		fmt.Printf("oracle query `%s' faild: %v, ignored \n", ec.sql, err)
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
	var (
		pt  *ifxcli.Point
		err error
	)

	if metricName == "oracle_system" {
		handleSystem(m, metricName, response)
		return nil
	}

	for _, item := range response {

		tags := map[string]string{}

		tags["oracle_service"] = m.serviceName
		tags["oracle_server"] = fmt.Sprintf("%s:%s", m.host, m.port)

		for _, tagKey := range tagsKeys {
			tags[tagKey] = strings.Replace(String(item[tagKey]), " ", "_", -1)
			delete(item, tagKey)
		}

		// add user-added tags
		// XXX: this may overwrite tags within @tags
		for k, v := range m.tags {
			tags[k] = v
		}

		pt, err = ifxcli.NewPoint(metricName, tags, item, time.Now())
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

func handleSystem(m *monitor, metricName string, response []map[string]interface{}) error {
	lines := [][]byte{}
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	// resData := make(map[string]interface{})
	for _, item := range response {
		tags["oracle_service"] = m.serviceName
		tags["oracle_server"] = fmt.Sprintf("%s:%s", m.host, m.port)

		// add user-added tags
		// XXX: this may overwrite tags within @tags
		for k, v := range m.tags {
			tags[k] = v
		}

		fieldName := String(item["metric_name"])
		value := item["value"]

		fieldName = strings.ToLower(strings.Replace(fieldName, " ", "_", -1))

		if newName, ok := dic[fieldName]; ok {
			fields[newName] = value
		}
	}

	pt, err := ifxcli.NewPoint(metricName, tags, fields, time.Now())
	if err != nil {
		l.Error("NewPoint(): %s", err.Error())
		return err
	}

	fmt.Println("point ======>", pt.String())

	lines = append(lines, []byte(pt.String()))

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
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))

	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer timeoutCancel()

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
    NVL(m.used_space * t.block_size, 0) as used_space,
    m.tablespace_size * t.block_size as ts_size,
    NVL(m.used_percent, 0) as in_use,
    NVL2(m.used_space, 0, 1) as off_use
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
		tagsMap:    []string{"tablespace_name"},
	},
}

var dic = map[string]string{
	"buffer_cache_hit_ratio":       "buffer_cachehit_ratio",
	"cursor_cache_hit_ratio":       "cursor_cachehit_ratio",
	"library_cache_hit_ratio":      "library_cachehit_ratio",
	"shared_pool_free_%":           "shared_pool_free",
	"physical_read_bytes_per_sec":  "physical_reads",
	"physical_write_bytes_per_sec": "physical_writes",
	"enqueue_timeouts_per_sec":     "enqueue_timeouts",

	"gc_cr_block_received_per_second": "gc_cr_block_received",
	"global_cache_blocks_corrupted":   "cache_blocks_corrupt",
	"global_cache_blocks_lost":        "cache_blocks_lost",
	"average_active_sessions":         "active_sessions",
	"sql_service_response_time":       "service_response_time",
	"user_rollbacks_per_sec":          "user_rollbacks",
	"total_sorts_per_user_call":       "sorts_per_user_call",
	"rows_per_sort":                   "rows_per_sort",
	"disk_sort_per_sec":               "disk_sorts",
	"memory_sorts_ratio":              "memory_sorts_ratio",
	"database_wait_time_ratio":        "database_wait_time_ratio",
	"session_limit_%":                 "session_limit_usage",
	"session_count":                   "session_count",
	"temp_space_used":                 "temp_space_used",
}
