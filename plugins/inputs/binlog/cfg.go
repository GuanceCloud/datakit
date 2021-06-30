// +build !386,!arm

package binlog

import (
	"math/rand"
	"time"

	"github.com/siddontang/go-mysql/mysql"
)

const (
	binlogConfigSample = `
#[[inputs.binlog]]

# ##(optional)
#null_int=0
#null_float=0.0

# ##(required)
#[[inputs.binlog.sources]]

#  ##(required) host of mysql, include port
#  addr='localhost:3306'

#  ##(required) username and password of mysql
#  user="admin"
#  password=""

#  ##(required)
#  [[inputs.binlog.sources.databases]]
#    ##(required) database name
#    db = 'test'
#	[[inputs.binlog.sources.databases.tables]]
#     ##(required) the name of table
#	  name='user'

#	  ##(optional) the name of metric, if empty use name as default
#	  measurement=""

#	  ##(required) specify the table's columns which will be taken as fields in metric, must be non-empty
#	  fields=['column0']

#	  ##(optional) specify the table's columns which will be taken as tags in metric, may empty
#	  tags=['column1']

#	  ##(optional)exlcude the events of binlog, there are 3 events: "insert","update","delete"
#      exclude_events=[]
`
)

var (
	HeartbeatPeriod = 60 * time.Second
	ReadTimeout     = 90 * time.Second
)

type TableConfig struct {
	Name        string `toml:"name"`                  //表名
	Measurement string `toml:"measurement,omitempty"` //不配置则默认为表名

	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`

	ExcludeListenEvents []string `toml:"exclude_events,omitempty"`
}

type DatabaseConfig struct {
	Database string         `toml:"db"`
	Tables   []*TableConfig `toml:"tables"`
}

type InstanceConfig struct {
	Addr     string `toml:"addr"`
	User     string `toml:"user"`
	Password string `toml:"password,omitempty"`
	Pwd      string `toml:"-"`

	Flavor   string `toml:"-"`
	Charset  string `toml:"-"`
	ServerID uint32 `toml:"-"`

	HeartbeatPeriod time.Duration `toml:"-"`
	ReadTimeout     time.Duration `toml:"-"`

	// discard row event without table meta
	DiscardNoMetaRowEvent bool `toml:"-"`

	UseDecimal bool `toml:"-"`
	ParseTime  bool `toml:"-"`

	TimestampStringLocation *time.Location `toml:"-"`

	// SemiSyncEnabled enables semi-sync or not.
	SemiSyncEnabled bool `toml:"-"`

	// Set to change the maximum number of attempts to re-establish a broken
	// connection
	MaxReconnectAttempts int `toml:"-"`

	Databases []*DatabaseConfig `toml:"databases"`
}

func (inst *InstanceConfig) applyDefault() {
	//一下选项是否做成可配置?
	inst.ServerID = uint32(rand.New(rand.NewSource(time.Now().Unix())).Intn(1000)) + 1001
	inst.Charset = mysql.DEFAULT_CHARSET
	inst.Flavor = mysql.MySQLFlavor
	inst.HeartbeatPeriod = HeartbeatPeriod
	inst.DiscardNoMetaRowEvent = true
	inst.ReadTimeout = ReadTimeout
	inst.UseDecimal = true
	inst.ParseTime = true
	inst.SemiSyncEnabled = false
}
