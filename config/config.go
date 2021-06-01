//nolint:gocyclo
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (

	// envVarRe is a regex to find environment variables in the config file
	envVarRe      = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

func LoadCfg(c *Config, mcp string) error {

	InitDirs()

	if Docker { // only accept configs from ENV under docker(or daemon-set) mode

		if runtime.GOOS != "linux" {
			return fmt.Errorf("docker mode not supported under %s", runtime.GOOS)
		}

		if err := c.LoadEnvs(); err != nil {
			return err
		}

		// 这里暂时用 hostname 当做 datakit ID, 后续肯定会移除掉, 即 datakit ID 基本已经废弃不用了,
		// 中心最终将通过统计主机个数作为 datakit 数量来收费.
		// 由于 datakit UUID 不再重要, 出错也不管了
		_ = c.SetUUID()

		_ = CreateSymlinks()

	} else {
		if err := c.LoadMainTOML(mcp); err != nil {
			return err
		}
	}

	if err := c.ApplyMainConfig(); err != nil {
		return err
	}

	if err := c.InitCfg(datakit.MainConfPath); err != nil {
		return err
	}

	l = logger.SLogger("config")

	l.Infof("main cfg: %+#v", c)

	initPluginSamples()
	if err := initPluginPipeline(); err != nil {
		l.Fatal(err)
	}

	initDefaultEnabledPlugins(c)

	if err := LoadInputsConfig(c); err != nil {
		l.Error(err)
		return err
	}

	return nil
}

func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}

func feedEnvs(data []byte) []byte {
	data = trimBOM(data)

	parameters := envVarRe.FindAllSubmatch(data, -1)

	l.Debugf("parameters: %s", parameters)

	for _, parameter := range parameters {
		if len(parameter) != 3 {
			continue
		}

		var envvar []byte
		if parameter[1] != nil {
			envvar = parameter[1]
		} else if parameter[2] != nil {
			envvar = parameter[2]
		} else {
			continue
		}

		envval, ok := os.LookupEnv(strings.TrimPrefix(string(envvar), "$"))
		if ok {
			envval = envVarEscaper.Replace(envval)
			data = bytes.Replace(data, parameter[0], []byte(envval), 1)
		} else {
			data = bytes.Replace(data, parameter[0], []byte("no-value"), 1)
		}
	}

	return data
}

func parseCfgFile(f string) (*ast.Table, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("read config %s failed: %s", f, err.Error())
	}

	data = feedEnvs(data)

	tbl, err := toml.Parse(data)
	if err != nil {
		l.Errorf("parse toml %s failed", string(data))
		return nil, err
	}

	return tbl, nil
}

func sliceContains(name string, list []string) bool {
	for _, b := range list {
		if b == name {
			return true
		}
	}
	return false
}
