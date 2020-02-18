package config

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/file"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/http"
)

var (
	ServiceName = "datakit"

	MainCfgPath   string
	ExecutableDir string

	ConvertedCfg []string
)

const (
	mainConfigTemplate = `uuid='{{.UUID}}'
ftdataway='{{.FtGateway}}'
log='{{.Log}}'
log_level='{{.LogLevel}}'
config_dir='{{.ConfigDir}}'

## Override default hostname, if empty use os.Hostname()
hostname = ""
## If set to true, do no set the "host" tag.
omit_hostname = false

## Global tags can be specified here in key="value" format.
#[global_tags]
# name = 'admin'
`
)

type Config struct {
	MainCfg          *MainConfig
	TelegrafAgentCfg *TelegrafAgentConfig
	Inputs           []*models.RunningInput
	Outputs          []*models.RunningOutput
}

func NewConfig() *Config {
	c := &Config{
		TelegrafAgentCfg: defaultTelegrafAgentCfg(),
		MainCfg: &MainConfig{
			FlushInterval: internal.Duration{Duration: 10 * time.Second},
			Interval:      internal.Duration{Duration: 10 * time.Second},
			LogLevel:      "info",
			RoundInterval: false,
		},
		Inputs:  make([]*models.RunningInput, 0),
		Outputs: make([]*models.RunningOutput, 0),
	}
	return c
}

type MainConfig struct {
	UUID      string `toml:"uuid"`
	FtGateway string `toml:"ftdataway"`

	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`

	ConfigDir string `toml:"config_dir,omitempty"`

	GlobalTags map[string]string `toml:"global_tags"`

	Interval      internal.Duration `toml:"interval"`
	RoundInterval bool
	FlushInterval internal.Duration

	OutputsFile string `toml:"output_file,omitempty"`

	Hostname     string
	OmitHostname bool
}

type ConvertTelegrafConfig interface {
	Load(f string) error
	ToTelegraf(f string) (string, error)
	FilePath(cfgdir string) string
}

func UserAgent() string {
	return fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)
}

func (c *Config) LoadMainConfig(ctx context.Context, maincfg string) error {
	data, err := ioutil.ReadFile(maincfg)
	if err != nil {
		return err
	}

	if tbl, err := parseConfig(data); err != nil {
		return err
	} else {

		//telegraf config
		bAgentSetLogLevel := false
		bAgentSetOmitHost := false
		bAgentSetHostname := false
		if val, ok := tbl.Fields["agent"]; ok {
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("invalid agent configuration")
			}

			if _, ok := subTable.Fields["debug"]; ok {
				bAgentSetLogLevel = true
			}

			if _, ok := subTable.Fields["omit_hostname"]; ok {
				bAgentSetOmitHost = true
			}

			if _, ok := subTable.Fields["hostname"]; ok {
				bAgentSetHostname = true
			}

			if err = toml.UnmarshalTable(subTable, c.TelegrafAgentCfg); err != nil {
				return fmt.Errorf("invalid agent configuration, %s", err)
			}

			delete(tbl.Fields, "agent")
		}

		if err := toml.UnmarshalTable(tbl, c.MainCfg); err != nil {
			return err
		}

		if !c.MainCfg.OmitHostname {
			if c.MainCfg.Hostname == "" {
				hostname, err := os.Hostname()
				if err != nil {
					return err
				}

				c.MainCfg.Hostname = hostname
			}

			c.MainCfg.GlobalTags["host"] = c.MainCfg.Hostname
		}

		if c.TelegrafAgentCfg.LogTarget == "file" && c.TelegrafAgentCfg.Logfile == "" {
			c.TelegrafAgentCfg.Logfile = filepath.Join(ExecutableDir, "agent.log")
		}

		if !bAgentSetLogLevel {
			c.TelegrafAgentCfg.Debug = (strings.ToLower(c.MainCfg.LogLevel) == "debug")
		}

		if !bAgentSetOmitHost {
			c.TelegrafAgentCfg.OmitHostname = c.MainCfg.OmitHostname
		}

		if !bAgentSetHostname {
			c.TelegrafAgentCfg.Hostname = c.MainCfg.Hostname
		}
	}

	return nil
}

func (c *Config) LoadConfig(ctx context.Context) error {

	for name, creator := range inputs.Inputs {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		if name == "apachelog" || name == "nginxlog" {
			if theip, ok := creator().(ConvertTelegrafConfig); ok {
				thepath := theip.FilePath(c.MainCfg.ConfigDir)
				_, err := os.Stat(thepath)

				if err != nil && os.IsNotExist(err) {
					continue
				}
				if err := theip.Load(thepath); err != nil {
					if err == ErrNoTelegrafConf {
						continue
					} else {
						return fmt.Errorf("fail to load %s, %s", thepath, err)
					}
				}
				if telestr, err := theip.ToTelegraf(thepath); err == nil {
					ConvertedCfg = append(ConvertedCfg, telestr)
				} else {
					return fmt.Errorf("convert %s to telegraf failed, %s", name, err)
				}
			}
			continue
		}

		path := filepath.Join(c.MainCfg.ConfigDir, name, fmt.Sprintf("%s.conf", name))

		_, err := os.Stat(path)
		if err != nil && os.IsNotExist(err) {
			continue
		}

		input := creator()
		data, err := ioutil.ReadFile(path)
		if err != nil {
			if err != nil {
				return fmt.Errorf("Error loading config file %s, %s", path, err)
			}
		}

		tbl, err := parseConfig(data)
		if err != nil {
			return fmt.Errorf("Error loading config file %s, %s", path, err)
		}

		if err := c.addInput(name, input, tbl); err != nil {
			return err
		}
	}

	// for name, creator := range outputs.Outputs {
	// 	output := creator()
	// 	if cfgdata, ok := embbed.EmbeddedOutputsConfs[name]; ok {
	// 		tbl, err := parseConfig([]byte(cfgdata))
	// 		if err != nil {
	// 			return fmt.Errorf("Error loading embbed config %s, %s", name, err)
	// 		}
	// 		if err := c.addOutput(name, output, tbl); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	if c.MainCfg.OutputsFile != "" {
		fileOutput := file.NewFileOutput()
		fileOutput.Files = []string{c.MainCfg.OutputsFile}
		if err := c.addOutputDirectly("file", fileOutput); err != nil {
			return err
		}
	}

	if c.MainCfg.FtGateway != "" {
		httpOutput := http.NewHttpOutput()
		if httpOutput.Headers == nil {
			httpOutput.Headers = map[string]string{}
		}
		httpOutput.Headers[`X-Datakit-UUID`] = c.MainCfg.UUID
		httpOutput.Headers[`X-Version`] = git.Version
		httpOutput.Headers[`User-Agent`] = UserAgent()
		httpOutput.ContentEncoding = "gzip"
		httpOutput.URL = c.MainCfg.FtGateway
		if err := c.addOutputDirectly("http", httpOutput); err != nil {
			return err
		}
	}

	return LoadTelegrafConfigs(ctx, c.MainCfg.ConfigDir)
}

func InitMainCfg(cfg *MainConfig, path string) error {

	var err error
	tm := template.New("")
	tm, err = tm.Parse(mainConfigTemplate)
	if err != nil {
		return fmt.Errorf("Error creating %s, %s", path, err)
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tm.Execute(buf, cfg); err != nil {
		return fmt.Errorf("Error creating %s, %s", path, err)
	}

	if err := ioutil.WriteFile(path, []byte(buf.Bytes()), 0664); err != nil {
		return fmt.Errorf("Error creating %s, %s", path, err)
	}
	return nil
}

func CreatePluginConfigs(cfgdir string, upgrade bool) error {

	//datakit定义的插件的配置文件
	for name, creator := range inputs.Inputs {
		plugindir := ""
		if name == "apachelog" {
			plugindir = filepath.Join(cfgdir, "apache")
		} else if name == "nginxlog" {
			plugindir = filepath.Join(cfgdir, "nginx")
		} else {
			plugindir = filepath.Join(cfgdir, name)
		}
		cfgpath := filepath.Join(plugindir, fmt.Sprintf(`%s.conf`, name))

		if upgrade {
			//更新时，不动它
			_, err := os.Stat(cfgpath)
			if err == nil {
				continue
			}
		}
		if err := os.MkdirAll(plugindir, 0775); err != nil {
			return fmt.Errorf("Error create %s, %s", plugindir, err)
		}
		input := creator()
		if err := ioutil.WriteFile(cfgpath, []byte(input.SampleConfig()), 0664); err != nil {
			return fmt.Errorf("Error create %s, %s", cfgpath, err)
		}
	}

	//创建各个telegraf插件的配置文件
	for _, name := range SupportsTelegrafMetraicNames {

		plugindir := filepath.Join(cfgdir, name)
		cfgpath := filepath.Join(plugindir, fmt.Sprintf(`%s.conf`, name))
		if upgrade {
			//更新时，不动它
			_, err := os.Stat(cfgpath)
			if err == nil {
				continue
			}
		}

		if err := os.MkdirAll(plugindir, 0775); err != nil {
			return fmt.Errorf("Error create %s, %s", plugindir, err)
		}
		if samp, ok := telegrafCfgSamples[name]; ok {

			if err := ioutil.WriteFile(cfgpath, []byte(samp), 0664); err != nil {
				return fmt.Errorf("Error create %s, %s", cfgpath, err)
			}
		}
	}

	return nil
}

func parseConfig(contents []byte) (*ast.Table, error) {
	// contents = trimBOM(contents)

	// parameters := envVarRe.FindAllSubmatch(contents, -1)
	// for _, parameter := range parameters {
	// 	if len(parameter) != 3 {
	// 		continue
	// 	}

	// 	var env_var []byte
	// 	if parameter[1] != nil {
	// 		env_var = parameter[1]
	// 	} else if parameter[2] != nil {
	// 		env_var = parameter[2]
	// 	} else {
	// 		continue
	// 	}

	// 	env_val, ok := os.LookupEnv(strings.TrimPrefix(string(env_var), "$"))
	// 	if ok {
	// 		env_val = escapeEnv(env_val)
	// 		contents = bytes.Replace(contents, parameter[0], []byte(env_val), 1)
	// 	}
	// }

	return toml.Parse(contents)
}

func (c *Config) addInput(name string, input telegraf.Input, table *ast.Table) error {

	pluginConfig, err := buildInput(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, input); err != nil {
		return err
	}

	rp := models.NewRunningInput(input, pluginConfig)
	rp.SetDefaultTags(c.MainCfg.GlobalTags)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

func (c *Config) addOutputDirectly(name string, output telegraf.Output) error {

	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializer(name, nil)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	}

	outputConfig, err := buildOutput(name, nil)
	if err != nil {
		return err
	}

	ro := models.NewRunningOutput(name, output, outputConfig, 0, 0)
	c.Outputs = append(c.Outputs, ro)
	return nil
}

func (c *Config) addOutput(name string, output telegraf.Output, table *ast.Table) error {

	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializer(name, table)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	}

	outputConfig, err := buildOutput(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, output); err != nil {
		return err
	}

	ro := models.NewRunningOutput(name, output, outputConfig, 0, 0)
	c.Outputs = append(c.Outputs, ro)
	return nil
}

func buildSerializer(name string, tbl *ast.Table) (serializers.Serializer, error) {
	c := &serializers.Config{TimestampUnits: time.Duration(1 * time.Second)}

	if tbl != nil {
		if node, ok := tbl.Fields["data_format"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					c.DataFormat = str.Value
				}
			}
		}
	}

	if c.DataFormat == "" {
		c.DataFormat = "influx"
	}

	if tbl != nil {
		delete(tbl.Fields, "influx_max_line_bytes")
		delete(tbl.Fields, "influx_sort_fields")
		delete(tbl.Fields, "influx_uint_support")
		delete(tbl.Fields, "graphite_tag_support")
		delete(tbl.Fields, "data_format")
		delete(tbl.Fields, "prefix")
		delete(tbl.Fields, "template")
		delete(tbl.Fields, "json_timestamp_units")
		delete(tbl.Fields, "splunkmetric_hec_routing")
		delete(tbl.Fields, "wavefront_source_override")
		delete(tbl.Fields, "wavefront_use_strict")
	}

	return serializers.NewSerializer(c)
}

func buildOutput(name string, tbl *ast.Table) (*models.OutputConfig, error) {
	var filter models.Filter
	var err error

	if tbl != nil {
		filter, err = buildFilter(tbl)
		if err != nil {
			return nil, err
		}
	}

	oc := &models.OutputConfig{
		Name:   name,
		Filter: filter,
	}

	// TODO
	// Outputs don't support FieldDrop/FieldPass, so set to NameDrop/NamePass
	if len(oc.Filter.FieldDrop) > 0 {
		oc.Filter.NameDrop = oc.Filter.FieldDrop
	}
	if len(oc.Filter.FieldPass) > 0 {
		oc.Filter.NamePass = oc.Filter.FieldPass
	}

	if tbl != nil {
		if node, ok := tbl.Fields["flush_interval"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					dur, err := time.ParseDuration(str.Value)
					if err != nil {
						return nil, err
					}

					oc.FlushInterval = dur
				}
			}
		}

		if node, ok := tbl.Fields["metric_buffer_limit"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if integer, ok := kv.Value.(*ast.Integer); ok {
					v, err := integer.Int()
					if err != nil {
						return nil, err
					}
					oc.MetricBufferLimit = int(v)
				}
			}
		}

		if node, ok := tbl.Fields["metric_batch_size"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if integer, ok := kv.Value.(*ast.Integer); ok {
					v, err := integer.Int()
					if err != nil {
						return nil, err
					}
					oc.MetricBatchSize = int(v)
				}
			}
		}

		if node, ok := tbl.Fields["alias"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					oc.Alias = str.Value
				}
			}
		}

		delete(tbl.Fields, "flush_interval")
		delete(tbl.Fields, "metric_buffer_limit")
		delete(tbl.Fields, "metric_batch_size")
		delete(tbl.Fields, "alias")
	}

	return oc, nil
}

func buildInput(name string, tbl *ast.Table) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}
	if node, ok := tbl.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err := time.ParseDuration(str.Value)
				if err != nil {
					return nil, err
				}

				cp.Interval = dur
			}
		}
	}

	if node, ok := tbl.Fields["name_prefix"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.MeasurementPrefix = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["name_suffix"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.MeasurementSuffix = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["name_override"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.NameOverride = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["alias"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.Alias = str.Value
			}
		}
	}

	cp.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := toml.UnmarshalTable(subtbl, cp.Tags); err != nil {
				log.Printf("E! Could not parse tags for input %s\n", name)
			}
		}
	}

	delete(tbl.Fields, "name_prefix")
	delete(tbl.Fields, "name_suffix")
	delete(tbl.Fields, "name_override")
	delete(tbl.Fields, "alias")
	delete(tbl.Fields, "interval")
	delete(tbl.Fields, "tags")
	var err error
	cp.Filter, err = buildFilter(tbl)
	if err != nil {
		return cp, err
	}
	return cp, nil
}

func buildFilter(tbl *ast.Table) (models.Filter, error) {
	f := models.Filter{}

	if node, ok := tbl.Fields["namepass"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.NamePass = append(f.NamePass, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["namedrop"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.NameDrop = append(f.NameDrop, str.Value)
					}
				}
			}
		}
	}

	fields := []string{"pass", "fieldpass"}
	for _, field := range fields {
		if node, ok := tbl.Fields[field]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if ary, ok := kv.Value.(*ast.Array); ok {
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							f.FieldPass = append(f.FieldPass, str.Value)
						}
					}
				}
			}
		}
	}

	fields = []string{"drop", "fielddrop"}
	for _, field := range fields {
		if node, ok := tbl.Fields[field]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if ary, ok := kv.Value.(*ast.Array); ok {
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							f.FieldDrop = append(f.FieldDrop, str.Value)
						}
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["tagpass"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					tagfilter := &models.TagFilter{Name: name}
					if ary, ok := kv.Value.(*ast.Array); ok {
						for _, elem := range ary.Value {
							if str, ok := elem.(*ast.String); ok {
								tagfilter.Filter = append(tagfilter.Filter, str.Value)
							}
						}
					}
					f.TagPass = append(f.TagPass, *tagfilter)
				}
			}
		}
	}

	if node, ok := tbl.Fields["tagdrop"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					tagfilter := &models.TagFilter{Name: name}
					if ary, ok := kv.Value.(*ast.Array); ok {
						for _, elem := range ary.Value {
							if str, ok := elem.(*ast.String); ok {
								tagfilter.Filter = append(tagfilter.Filter, str.Value)
							}
						}
					}
					f.TagDrop = append(f.TagDrop, *tagfilter)
				}
			}
		}
	}

	if node, ok := tbl.Fields["tagexclude"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.TagExclude = append(f.TagExclude, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["taginclude"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.TagInclude = append(f.TagInclude, str.Value)
					}
				}
			}
		}
	}
	if err := f.Compile(); err != nil {
		return f, err
	}

	delete(tbl.Fields, "namedrop")
	delete(tbl.Fields, "namepass")
	delete(tbl.Fields, "fielddrop")
	delete(tbl.Fields, "fieldpass")
	delete(tbl.Fields, "drop")
	delete(tbl.Fields, "pass")
	delete(tbl.Fields, "tagdrop")
	delete(tbl.Fields, "tagpass")
	delete(tbl.Fields, "tagexclude")
	delete(tbl.Fields, "taginclude")
	return f, nil
}
