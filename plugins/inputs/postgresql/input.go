package postgresql

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dk "gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkInputs "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = "postgresql"
	catalogName = "db"
	l           = logger.DefaultSLogger(inputName)
)

const sampleConfig = `
[[inputs.postgresql]]
## 服务器地址
# url格式
# postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
# 简单字符串格式
# host=localhost user=pqgotest password=... sslmode=... dbname=app_production

address = "postgres://postgres@localhost/test?sslmode=disable"

## 配置采集的数据库，默认会采集所有的数据库，当同时设置ignored_databases和databases会忽略databases
# ignored_databases = ["db1"]
# databases = ["db1"]

## 设置服务器Tag，默认是基于服务器地址生成
# outputaddress = "db01"

## 采集间隔
# 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
interval = "10s"

## 日志采集
# [inputs.postgresql.log]
# files = []
# pipeline = "postgresql.p"

## 自定义Tag
[inputs.postgresql.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
`

const pipelineCfg = `
add_pattern("log_date", "%{YEAR}-%{MONTHNUM}-%{MONTHDAY}%{SPACE}%{HOUR}:%{MINUTE}:%{SECOND}%{SPACE}(?:CST|UTC)")
add_pattern("log_level", "(LOG|ERROR|FATAL|PANIC|WARNING|NOTICE|INFO)")
add_pattern("session_id", "([.0-9a-z]*)")
add_pattern("application_name", "(\\[%{GREEDYDATA:application_name}?\\])")
add_pattern("remote_host", "(\\[\\[?%{HOST:remote_host}?\\]?\\])")
grok(_, "%{log_date:time}%{SPACE}\\[%{INT:process_id}\\]%{SPACE}(%{WORD:db_name}?%{SPACE}%{application_name}%{SPACE}%{USER:user}?%{SPACE}%{remote_host}%{SPACE})?%{session_id:session_id}%{SPACE}(%{log_level:log_level}:)?")

# default
grok(_, "%{log_date:time}%{SPACE}\\[%{INT:process_id}\\]%{SPACE}%{log_level:log_level}")

nullif(remote_host, "")
nullif(session_id, "")
nullif(application_name, "")
nullif(user, "")
nullif(db_name, "")

group_in(log_level, [""], "INFO")

default_time(time)
`

type Rows interface {
	Close() error
	Columns() ([]string, error)
	Next() bool
	Scan(...interface{}) error
}

type Service interface {
	Start() error
	Stop() error
	Query(string) (Rows, error)
	SetAddress(string)
	GetColumnMap(scanner, []string) (map[string]*interface{}, error)
}

type scanner interface {
	Scan(dest ...interface{}) error
}

type Input struct {
	Address          string                 `toml:"address"`
	Outputaddress    string                 `toml:"outputaddress"`
	IgnoredDatabases []string               `toml:"ignored_databases"`
	Databases        []string               `toml:"databases"`
	Interval         string                 `toml:"interval"`
	Tags             map[string]string      `toml:"tags"`
	Log              *dkInputs.TailerOption `toml:"log"`

	service      Service
	tail         Tailer
	duration     time.Duration
	collectCache []dkInputs.Measurement
}

type inputMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m inputMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m inputMeasurement) Info() *dkInputs.MeasurementInfo {
	return &dkInputs.MeasurementInfo{
		Name:   inputName,
		Fields: postgreFields,
		Tags: map[string]interface{}{
			"server": dkInputs.NewTagInfo("The server address"),
			"db":     dkInputs.NewTagInfo("The database name"),
		},
	}
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return dk.AllArch
}

func (i *Input) SampleMeasurement() []dkInputs.Measurement {
	return []dkInputs.Measurement{
		&inputMeasurement{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{
		"postgresql": pipelineCfg,
	}
}

func (p *Input) SanitizedAddress() (sanitizedAddress string, err error) {
	var (
		canonicalizedAddress string
	)

	var kvMatcher, _ = regexp.Compile(`(password|sslcert|sslkey|sslmode|sslrootcert)=\S+ ?`)

	if p.Outputaddress != "" {
		return p.Outputaddress, nil
	}

	if strings.HasPrefix(p.Address, "postgres://") || strings.HasPrefix(p.Address, "postgresql://") {
		if canonicalizedAddress, err = parseURL(p.Address); err != nil {
			return sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = p.Address
	}

	sanitizedAddress = kvMatcher.ReplaceAllString(canonicalizedAddress, "")

	return sanitizedAddress, err
}

func (i *Input) executeQuery(query string) error {
	var (
		columns []string
		err     error
	)

	rows, err := i.service.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	if columns, err = rows.Columns(); err != nil {
		return err
	}

	for rows.Next() {
		columnMap, err := i.service.GetColumnMap(rows, columns)
		if err != nil {
			return err
		}
		err = i.accRow(columnMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Input) getDbMetrics() error {
	query := `
	SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size 
	FROM pg_stat_database psd 
	JOIN pg_database pd ON psd.datname = pd.datname 
	WHERE psd.datname not ilike 'template%'   AND psd.datname not ilike 'rdsadmin'   
	AND psd.datname not ilike 'azure_maintenance'   AND psd.datname not ilike 'postgres'
	`
	if len(i.IgnoredDatabases) != 0 {
		query += fmt.Sprintf(` AND psd.datname NOT IN ('%s')`, strings.Join(i.IgnoredDatabases, "','"))
	} else if len(i.Databases) != 0 {
		query += fmt.Sprintf(` AND psd.datname IN ('%s')`, strings.Join(i.Databases, "','"))
	}

	err := i.executeQuery(query)

	return err
}

func (i *Input) getBgwMetrics() error {
	query := `
		select * FROM pg_stat_bgwriter
	`
	err := i.executeQuery(query)
	return err
}

func (i *Input) getConnectionMetrics() error {
	query := `
		WITH max_con AS (SELECT setting::float FROM pg_settings WHERE name = 'max_connections') 
		SELECT MAX(setting) AS max_connections, SUM(numbackends)/MAX(setting) AS percent_usage_connections
		FROM pg_stat_database, max_con
	`

	err := i.executeQuery(query)
	return err
}

func (i *Input) Collect() error {
	var (
		err error
	)

	i.service.SetAddress(i.Address)
	defer i.service.Stop()
	err = i.service.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(3)
	// collect db metrics
	go func() {
		defer wg.Done()
		if e := i.getDbMetrics(); e != nil {
			l.Error(e.Error())
			err = e
		}
	}()

	// collect bgwriter
	go func() {
		defer wg.Done()
		if e := i.getBgwMetrics(); e != nil {
			l.Error(e.Error())
			err = e
		}
	}()

	// connection
	go func() {
		defer wg.Done()
		if e := i.getConnectionMetrics(); e != nil {
			l.Error(e.Error())
			err = e
		}
	}()

	wg.Wait()
	return err
}

func (i *Input) accRow(columnMap map[string]*interface{}) error {
	var tagAddress string
	tagAddress, err := i.SanitizedAddress()
	if err != nil {
		return err
	}

	tags := map[string]string{"server": tagAddress, "db": "postgres"}

	if i.Tags != nil {
		for k, v := range i.Tags {
			tags[k] = v
		}
	}

	fields := make(map[string]interface{})
	for col, val := range columnMap {
		if col != "datname" {
			if _, isValidCol := postgreFields[col]; !isValidCol {
				continue
			}
		}

		if *val != nil {
			value := *val
			switch trueVal := value.(type) {
			case []uint8:
				if col == "datname" {
					tags["db"] = string(trueVal)
				} else {
					fields[col] = string(trueVal)
				}
			default:
				fields[col] = trueVal
			}
		}
	}
	if len(fields) > 0 {
		i.collectCache = append(i.collectCache, &inputMeasurement{
			name:   inputName,
			fields: fields,
			tags:   tags,
			ts:     time.Now(),
		})
	}

	return nil

}

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (i *Input) runService(inputs Inputs, datakit Datakit) {
	if i.Log != nil {
		go func() {
			inputs.JoinPipelinePath(i.Log, "postgresql.p")
			i.Log.Source = inputName
			i.Log.Tags = make(map[string]string)
			for k, v := range i.Tags {
				i.Log.Tags[k] = v
			}
			tail, err := inputs.NewTailer(i.Log)
			if err != nil {
				l.Errorf("init tailf err: %s", err.Error())
				return
			}
			i.tail = tail
			tail.Run()
		}()
	}

	duration, err := time.ParseDuration(i.Interval)
	if err != nil {
		l.Error(fmt.Errorf("invalid interval, %s", err.Error()))
	} else if duration <= 0 {
		l.Error(fmt.Errorf("invalid interval, cannot be less than zero"))
	}

	i.duration = datakit.ProtectedInterval(minInterval, maxInterval, duration)

	tick := time.NewTicker(i.duration)

	for {
		select {
		case <-datakit.Exit():
			l.Infof("%s exit", inputName)
			return
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			} else {
				if len(i.collectCache) > 0 {
					err := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache, &io.Option{CollectCost: time.Since(start)})
					if err != nil {
						io.FeedLastError(inputName, err.Error())
						l.Error(err.Error())
					}
					i.collectCache = i.collectCache[:0]
				}
			}
		}
	}
}

type Datakit struct {
	ch                chan interface{}
	Metric            string
	ProtectedInterval func(min, max, cur time.Duration) time.Duration
	Exit              func() <-chan interface{}
}

func (d Datakit) Close() {
	close(d.ch)
}

type Tailer interface {
	Run()
}

type Inputs interface {
	JoinPipelinePath(interface{}, string)
	NewTailer(interface{}) (Tailer, error)
	FeedMeasurement(string, string, interface{}, interface{}) error
}

type DkInputs struct{}

func (DkInputs) FeedMeasurement(name, category string, measurements interface{}, opt interface{}) error {
	return dkInputs.FeedMeasurement(name, category, measurements.([]dkInputs.Measurement), opt.(*io.Option))
}

func (DkInputs) NewTailer(opt interface{}) (Tailer, error) {
	tailer, error := dkInputs.NewTailer(opt.(*dkInputs.TailerOption))
	return tailer, error
}

func (DkInputs) JoinPipelinePath(op interface{}, defaultPipeline string) {
	dkInputs.JoinPipelinePath(op.(*dkInputs.TailerOption), defaultPipeline)
}

func (i *Input) Run() {
	inputs := DkInputs{}
	datakit := Datakit{
		ch:     make(chan interface{}),
		Metric: dk.Metric,
	}
	datakit.ProtectedInterval = func(min, max, cur time.Duration) time.Duration {
		return dk.ProtectedInterval(min, max, cur)
	}
	datakit.Exit = func() <-chan interface{} {
		return dk.Exit.Wait()
	}

	i.runService(inputs, datakit)
}

func parseURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	var kvs []string
	escaper := strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
	accrue := func(k, v string) {
		if v != "" {
			kvs = append(kvs, k+"="+escaper.Replace(v))
		}
	}

	if u.User != nil {
		v := u.User.Username()
		accrue("user", v)

		v, _ = u.User.Password()
		accrue("password", v)
	}

	if host, port, err := net.SplitHostPort(u.Host); err != nil {
		accrue("host", u.Host)
	} else {
		accrue("host", host)
		accrue("port", port)
	}

	if u.Path != "" {
		accrue("dbname", u.Path[1:])
	}

	q := u.Query()
	for k := range q {
		accrue(k, q.Get(k))
	}

	sort.Strings(kvs)
	return strings.Join(kvs, " "), nil
}

func NewInput(service Service) *Input {
	input := &Input{
		Interval: "10s",
	}
	input.service = service
	return input
}

func init() {
	dkInputs.Add(inputName, func() dkInputs.Input {
		service := &SqlService{
			MaxIdle:     1,
			MaxOpen:     1,
			MaxLifetime: time.Duration(0),
		}
		input := NewInput(service)
		return input
	})
}
