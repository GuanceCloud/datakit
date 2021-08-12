package cmds

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type integration struct {
	Title  string   `yaml:"title"`
	Tags   []string `yaml:"tags"`
	DocURL string   `yaml:"url"`
}

var (
	titles = map[string]string{
		"apache":         "Apache",
		"cloudprober":    "Cloudprober",
		"container":      "容器",
		"cpu":            "CPU",
		"dataway":        "DataWay",
		"ddtrace":        "DDTrace",
		"dialtesting":    "网络拨测",
		"disk":           "Disk",
		"diskio":         "DiskIO",
		"docker":         "Docker",
		"elasticsearch":  "ElasticSearch",
		"gitlab":         "Gitlab",
		"host_processes": "Process",
		"hostobject":     "主机",
		"iis":            "IIS",
		"influxdb":       "InfluxDB",
		"jenkins":        "Jenkins",
		"jvm":            "JVM",
		"kafka":          "Kafka",
		"kubernetes":     "Kubernetes",
		"logging":        "日志",
		"mem":            "Memory",
		"memcached":      "Memcached",
		"mongodb":        "MongoDB",
		"mysql":          "MySQL",
		"net":            "Net",
		"nginx":          "Nginx",
		"oracle":         "Oracle",
		"postgresql":     "PostgreSQL",
		"proxy":          "DataKit代理",
		"rabbitmq":       "RabbitMQ",
		"redis":          "Redis",
		"rum":            "RUM",
		"sensors":        "硬件温度 Sensors",
		"smart":          "磁盘 S.M.A.R.T",
		"solr":           "Solr",
		"sqlserver":      "SQLServer",
		"ssh":            "SSH",
		"statsd":         "Statsd",
		"swap":           "Swap",
		"system":         "System",
		"tomcat":         "Tomcat",
		"prom":           "Prometheus 数据接入",
		"windows_event":  "Windows 事件",
	}
)

func exportIntegration(to, ignore string) error {
	if err := os.MkdirAll(to, 0600); err != nil {
		return err
	}

	arr := strings.Split(ignore, ",")
	skip := map[string]bool{}
	for _, x := range arr {
		skip[x] = true
	}

	for k, i := range inputs.Inputs {
		if _, ok := skip[k]; ok {
			continue
		}

		switch i().(type) {
		case inputs.InputV2:

			title := titles[k]
			if title == "" {
				title = k
			}

			x := integration{
				Title:  title,
				Tags:   []string{"IT运维", "指标采集"},
				DocURL: "https://www.yuque.com/dataflux/datakit/" + k,
			}

			if err := os.MkdirAll(filepath.Join(to, k), os.ModePerm); err != nil {
				return err
			}

			mf, err := yaml.Marshal(x)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(filepath.Join(to, k, "manifest.yaml"), mf, os.ModePerm); err != nil {
				return err
			}

		default:
			l.Debugf("ignore %s", k)
		}
	}
	return nil
}
