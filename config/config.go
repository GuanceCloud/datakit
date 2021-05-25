//nolint:gocyclo
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l = logger.DefaultSLogger("config")

	// envVarRe is a regex to find environment variables in the config file
	envVarRe      = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

func LoadCfg(c *datakit.Config, mcp string) error {

	datakit.InitDirs()

	if err := c.LoadEnvs(mcp); err != nil {
		return err
	}

	if err := c.LoadMainConfig(mcp); err != nil {
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
