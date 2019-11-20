package apache

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func init() {
	config.AddConfig("apachelog", &ApacheLogConfig{})
	config.AddConfig("apache", &ApacheConfig{})
}

const (
	apacheSample = `
# [[inputs.apache]]
#   ## An array of URLs to gather from, must be directed at the machine.
#   ## readable version of the mod_status page including the auto query string.
#   ## Default is "http://localhost/server-status?auto".
#   urls = ["http://localhost/server_status?auto"]
#
#   ## Optional TLS Config
#   tls_ca = "/etc/telegraf/ca.pem"
#   tls_cert = "/etc/telegraf/cert.cer"
#   tls_key = "/etc/telegraf/key.key"
#   ## Use TLS but skip chain & host verification
#   insecure_skip_verify = false
#
#   # HTTP response timeout (default: 5s)
#   response_timeout = "5s"
`
)

type ApacheStatus struct {
	Urls               string `toml:"urls"`
	TlsCa              string `toml:"tls_ca"`
	TlsCert            string `toml:"tls_cert"`
	TlsKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	ResponseTimeout    string `toml:"response_timeout"`
}

type ApacheConfig struct {
	Status []*ApacheStatus `toml:"inputs.apache"`
}

func (c *ApacheConfig) SampleConfig() string {
	return apacheSample
}

func (c *ApacheConfig) FilePath(root string) string {
	d := filepath.Join(root, "apache")
	return filepath.Join(d, "apache.conf")
}

func (c *ApacheConfig) ToTelegraf() (string, error) {
	d, err := toml.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

func (c *ApacheConfig) Load(f string) error {

	cfgdata, err := ioutil.ReadFile(f)

	tbl, err := toml.Parse(cfgdata)
	if err != nil {
		return err
	}

	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("invalid configuration")
		}
		fmt.Println(name)

		if name == "inputs" {

			for pluginName, pluginVal := range subTable.Fields {
				if pluginName != "apache" {
					continue
				}

				switch pluginSubTable := pluginVal.(type) {
				case []*ast.Table:
					for _, t := range pluginSubTable {
						cfg := &ApacheStatus{}
						if err = toml.UnmarshalTable(t, cfg); err != nil {
							return err
						}
						c.Status = append(c.Status, cfg)
					}
				}
			}
			break
		}
	}

	return nil
}
