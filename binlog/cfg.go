package binlog

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/influxdata/toml"
)

var (
	Cfg BinlogConfig

	binlogConfigSample = `
disable=true
null_int=0
null_float=0.0
[[sources]]
  #host of mysql, include port
  addr="localhost:3306"

  #username and password of mysql
  user="admin"
  password=""

  [[sources.databases]]
    db = "test"
	[[sources.databases.tables]]
#     ##the name of table
	  name="user"

#	  ##the name of metric, if empty use name as default
	  measurement=""

#	  ##specify the table's columns which will be taken as fields in metric, must be non-empty
	  fields=["column0"]

#	  ##specify the table's columns which will be taken as tags in metric, may empty
	  tags=["column1"]

#	  ##exlcude the events of binlog, there are 3 events: "insert","update","delete"
      exclude_events=[]
`
)

type BinlogTable struct {
	Name        string `toml:"name"`                  //表名
	Measurement string `toml:"measurement,omitempty"` //不配置则默认为表名

	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`

	//Columns             map[string]string `yaml:"columns"` //表中字段哪些作为tag，哪些字段作为field，至少配置一个field，忽略blob类型
	ExcludeListenEvents []string `toml:"exclude_events,omitempty"`
}

type BinlogDatabase struct {
	Database string         `toml:"db"`
	Tables   []*BinlogTable `toml:"tables"`
}

type BinlogDatasource struct {
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

	Databases []*BinlogDatabase `toml:"databases"`
}

type BinlogConfig struct {
	Disable     bool                `toml:"disable"`
	NullInt     int                 `toml:"null_int"`
	NullFloat   float64             `toml:"null_float"`
	Datasources []*BinlogDatasource `toml:"sources"`
}

func (c *BinlogConfig) SampleConfig() string {
	return binlogConfigSample
}

func (c *BinlogConfig) FilePath(d string) string {
	dir := filepath.Join(d, "binlog")
	return filepath.Join(dir, "binlog.conf")
}

func (c *BinlogConfig) ToTelegraf() (string, error) {
	return "", nil
}

func (c *BinlogConfig) Load(f string) error {

	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err = toml.Unmarshal(cfgdata, c); err != nil {
		return err
	}

	if len(c.Datasources) == 0 {
		return fmt.Errorf("binlog source not found")
	}

	//var tables []*BinlogTable

	for _, dt := range c.Datasources {
		if dt.Addr == "" || dt.User == "" {
			return fmt.Errorf("mysql host and username required")
		}

		if len(dt.Databases) == 0 {
			return fmt.Errorf("please specify at least one database")
		}

		for _, d := range dt.Databases {
			if d.Database == "" {
				return fmt.Errorf("database name must be non-empty")
			}

			if len(d.Tables) == 0 {
				return fmt.Errorf("please specify at least one table of %s", d.Database)
			}

			for _, t := range d.Tables {

				if t.Name == "" {
					return fmt.Errorf("table name must be non-empty")
				}

				// for k, v := range t.Columns {
				// 	if v == "tag" {
				// 		t.Tags = append(t.Tags, k)
				// 	}
				// 	if v == "field" {
				// 		t.Fields = append(t.Fields, k)
				// 	}
				// }

				if len(t.Fields) == 0 {
					return fmt.Errorf("please specify at least one column as field for table: %s.%s", d.Database, t.Name)
				}

				for _, fn := range t.Fields {
					for _, tn := range t.Tags {
						if tn == fn {
							return fmt.Errorf("column \"%s\" of %s.%s cannot be both as field and tag", tn, d.Database, t.Name)
						}
					}
				}

				// bexclude := false
				// for _, et := range d.ExcludeTables {
				// 	if et == t.Table {
				// 		bexclude = true
				// 		break
				// 	}
				// }

				// if !bexclude {
				// 	tables = append(tables, t)
				// }
			}
		}

	}

	// if len(tables) == 0 {
	// 	return fmt.Errorf("no table specified")
	// }

	return nil
}
