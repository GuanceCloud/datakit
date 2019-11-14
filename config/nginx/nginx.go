package nginx

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func init() {
	config.AddConfig("nginxlog", &NginxLogConfig{})
	config.AddConfig("nginx", &NginxConfig{})

}

const (
	nginxSample = `
# [[inputs.nginx]]
#   # An array of Nginx stub_status URI to gather stats.
#   urls = ["http://localhost/server_status"]
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

# # Read Nginx Plus' full status information (ngx_http_status_module)
# [[inputs.nginx_plus]]
#   ## An array of ngx_http_status_module or status URI to gather stats.
#   urls = ["http://localhost/status"]
#
#   # HTTP response timeout (default: 5s)
#   response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false

# # Read Nginx virtual host traffic status module information (nginx-module-vts)
# [[inputs.nginx_vts]]
#   ## An array of ngx_http_status_module or status URI to gather stats.
#   urls = ["http://localhost/status"]
#
#   ## HTTP response timeout (default: 5s)
#   response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`
)

type NginxStatus struct {
	Urls               string `toml:"urls"`
	TlsCa              string `toml:"tls_ca"`
	TlsCert            string `toml:"tls_cert"`
	TlsKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	ResponseTimeout    string `toml:"response_timeout"`
}

type NginxConfig struct {
	Status     []*NginxStatus `toml:"inputs.nginx"`
	PlusStatus []*NginxStatus `toml:"inputs.nginx_plus"`
}

func (c *NginxConfig) SampleConfig() string {
	return nginxSample
}

func (c *NginxConfig) FilePath(dir string) string {
	nginxdir := filepath.Join(dir, "nginx")
	return filepath.Join(nginxdir, "nginx.toml")
}

func (c *NginxConfig) ToTelegraf() (string, error) {
	d, err := toml.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

func (c *NginxConfig) Load(f string) error {

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
				if pluginName != "nginx" && pluginName != "nginx_plus" {
					continue
				}

				switch pluginSubTable := pluginVal.(type) {
				case []*ast.Table:
					for _, t := range pluginSubTable {
						cfg := &NginxStatus{}
						if err = toml.UnmarshalTable(t, cfg); err != nil {
							return err
						}
						if pluginName == "nginx" {
							c.Status = append(c.Status, cfg)
						} else {
							c.PlusStatus = append(c.PlusStatus, cfg)
						}
					}
				}
			}
			break
		}
	}

	return nil
}
