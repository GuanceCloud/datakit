package config

import (
	"errors"
	"fmt"
	"time"
)

var (
	binlogCfgTemplate = `  disable: false
  jobs:
  - ft_gatway: ""
    addr: ""
    user: ""
    password: ""

    inputs:
    - database: test
      tables:
      - table: table1
        measurement: table1
        columns:
          field1: field
          tag1: tag
      #  exclude_events:
      #  - delete
      #- table: table2
      #  measurement: table2
      #  columns:
      #    field1: field
      #    tag1: tag
      #exclude_tables:
      #- table3
`
)

type BinlogTable struct {
	Table       string `yaml:"table"`                 //表名
	Measurement string `yaml:"measurement,omitempty"` //不配置则默认为表名

	Tags   []string `yaml:"-"`
	Fields []string `yaml:"-"`

	Columns             map[string]string `yaml:"columns"` //表中字段哪些作为tag，哪些字段作为field，至少配置一个field，忽略blob类型
	ExcludeListenEvents []string          `yaml:"exclude_events,omitempty"`
}

type BinlogInput struct {
	Database      string         `yaml:"database"`
	Tables        []*BinlogTable `yaml:"tables,omitempty"`
	ExcludeTables []string       `yaml:"exclude_tables,omitempty"`
}

type BinlogDatasource struct {
	FTGateway string `yaml:"ft_gateway"`

	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password,omitempty"`
	Pwd      string `yaml:"-"`

	Flavor   string `yaml:"-"`
	Charset  string `yaml:"-"`
	ServerID uint32 `yaml:"-"`

	HeartbeatPeriod time.Duration `yaml:"-"`
	ReadTimeout     time.Duration `yaml:"-"`

	// discard row event without table meta
	DiscardNoMetaRowEvent bool `yaml:"-"`

	UseDecimal bool `yaml:"-"`
	ParseTime  bool `yaml:"-"`

	TimestampStringLocation *time.Location `yaml:"-"`

	// SemiSyncEnabled enables semi-sync or not.
	SemiSyncEnabled bool `yaml:"-"`

	// Set to change the maximum number of attempts to re-establish a broken
	// connection
	MaxReconnectAttempts int `yaml:"-"`

	Inputs []*BinlogInput `yaml:"inputs"`

	//CustomTags map[string]string `yaml:"custom_tags,omitempty"` //自定义的tag
}

type BinlogConfig struct {
	Disable     bool                `yaml:"disable"`
	Datasources []*BinlogDatasource `yaml:"jobs"`
}

func GenBinlogInitCfg() string {

	return binlogCfgTemplate
}

func (c *BinlogConfig) Check() error {
	if len(c.Datasources) == 0 {
		return errors.New("please config at least one mysql server")
	}

	var tables []*BinlogTable

	for _, dt := range c.Datasources {
		if dt.Addr == "" || dt.User == "" {
			return errors.New("mysql host and username required")
		}

		if dt.FTGateway == "" {
			return errors.New("ftagent addr required")
		}

		if len(dt.Inputs) == 0 {
			return errors.New("please specify at least one database")
		}

		for _, d := range dt.Inputs {
			if len(d.Tables) == 0 {
				return fmt.Errorf("please specify at least one table of %s", d.Database)
			}

			for _, t := range d.Tables {

				//fmt.Printf("cols: %#v\n", t.Columns)

				for k, v := range t.Columns {
					if v == "tag" {
						t.Tags = append(t.Tags, k)
					}
					if v == "field" {
						t.Fields = append(t.Fields, k)
					}
				}

				if len(t.Fields) == 0 {
					return fmt.Errorf("please specify at least one column as field of table: %s.%s", d.Database, t.Table)
				}

				bexclude := false
				for _, et := range d.ExcludeTables {
					if et == t.Table {
						bexclude = true
						break
					}
				}

				if !bexclude {
					tables = append(tables, t)
				}
			}
		}

	}

	if len(tables) == 0 {
		return errors.New("no table specified from configuration")
	}

	return nil
}
