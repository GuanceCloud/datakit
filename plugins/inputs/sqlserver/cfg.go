package sqlserver

import (
	"database/sql"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	sample = `
[[inputs.sqlserver]]
	# your sqlserver host ,example ip:port
	host = ""
	# your sqlserver user,password
	user = ""
	password = ""
	# ##(optional) collection interval, default is 10s
	# interval = "10s"
	[inputs.sqlserver.log]
	#	files = []
	#	# grok pipeline script path
	#	pipeline = "sqlserver.p"
	[inputs.sqlserver.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
	# ...`

	pipeline = `
grok(_,"%{TIMESTAMP_ISO8601:time} %{NOTSPACE:origin}\\s+%{GREEDYDATA:msg}")
default_time(time)
`

	inputName    = `sqlserver`
	catalogName  = "db"
	l            = logger.DefaultSLogger(inputName)
	collectCache []*io.Point
	minInterval  = time.Second * 5
	maxInterval  = time.Second * 30
	query        = []string{
		sqlServerPerformanceCounters,
		sqlServerWaitStatsCategorized,
		sqlServerDatabaseIO,
		sqlServerProperties,
		sqlServerSchedulers,
		sqlServerVolumeSpace,
	}
)

type Input struct {
	Host     string            `toml:"host"`
	User     string            `toml:"user"`
	Password string            `toml:"password"`
	Interval datakit.Duration  `toml:"interval"`
	Tags     map[string]string `toml:"tags"`
	Log      *sqlserverlog     `toml:"log"`

	lastErr error
	tail    *tailer.Tailer
	start   time.Time
	db      *sql.DB
}

type sqlserverlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	Match             string   `toml:"match"`
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newTimeFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeIByte,
		Desc:     desc,
	}
}

func newBoolFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Bool,
		Type:     inputs.Gauge,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}
