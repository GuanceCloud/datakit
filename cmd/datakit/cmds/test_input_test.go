package cmds

/*
import (
	"io/ioutil"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

var (
	mConfigFiles = map[string]string{
		"nginx.conf":      nginxConf,
		"hostobject.conf": hostobjectConf,
		"mysql.conf":      mysqlConf,
	}
)

// go test -v -timeout 30s -run ^TestPrintInputOnce$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds
func TestPrintInputOnce(t *testing.T) {
	configFiles, err := ReleaseFiles()
	if err != nil {
		t.Error(err)
		return
	}

	for _, v := range configFiles {
		t.Log(v)
		if err := printInputOnceCore(t, v); err != nil {
			t.Error(err)
			return
		}
	}

	t.Log("PrintInputOnce test completed!")
}

func printInputOnceCore(t *testing.T, configPath string) error {
	t.Helper()

	FlagCmdLogPath := "/dev/null"

	tryLoadMainCfg()
	setCmdRootLog(FlagCmdLogPath)
	if err := inputDebugger(configPath); err != nil {
		t.Errorf("inputDebugger: %s", err)
		return err
	}

	return nil
}

// go test -v -timeout 30s -run ^TestReleaseFiles$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds
func TestReleaseFiles(t *testing.T) {
	arr, err := ReleaseFiles()
	if err != nil {
		t.Error(err)
		return
	}

	for _, v := range arr {
		t.Log(v)
	}

	t.Log("TestReleaseFiles completed!")
}

func ReleaseFiles() ([]string, error) {
	l = logger.SLogger(pkgName)

	var fullPaths []string

	for name, content := range mConfigFiles {
		pyTmpFle, err := ioutil.TempFile("", name)
		if err != nil {
			l.Errorf("ioutil.TempFile failed: %s", err.Error())
			return []string{}, err
		}

		if _, err = pyTmpFle.WriteString(content); err != nil {
			l.Errorf("TempFile.WriteString failed: %s", err.Error())
			return []string{}, err
		}

		fullPaths = append(fullPaths, pyTmpFle.Name())
		if err := pyTmpFle.Close(); err != nil {
			l.Errorf("pyTmpFle.Close failed: %s", err.Error())
		}
	}

	return fullPaths, nil
}

const (
	pkgName = "configs"

	nginxConf = `
# {"version": "1.1.9-rc2-48-g0f8a5e1ca5", "desc": "do NOT edit this line"}

[[inputs.nginx]]
	url = "http://localhost:8080/server_status"
	# ##(optional) collection interval, default is 30s
	# interval = "30s"
	use_vts = false
	## Optional TLS Config
	# tls_ca = "/xxx/ca.pem"
	# tls_cert = "/xxx/cert.cer"
	# tls_key = "/xxx/key.key"
	## Use TLS but skip chain & host verification
	insecure_skip_verify = false
	# HTTP response timeout (default: 5s)
	response_timeout = "20s"

	[inputs.nginx.log]
	files = ["/usr/local/var/log/nginx/access.log","/usr/local/var/log/nginx/error.log"]
	#	# grok pipeline script path
	pipeline = "nginx.p"
	[inputs.nginx.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
	# ...%
`  // nginxConf

	hostobjectConf = `
[inputs.hostobject]

#pipeline = '' # optional

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

## Ignore mount points by filesystem type. Default ignored following FS types
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]


[inputs.hostobject.tags] # (optional) custom tags
# cloud_provider = "aliyun" # aliyun/tencent/aws
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
`  // hostobjectConf

	mysqlConf = `
# {"version": "1.1.9-rc4.2-105-g02717ef041", "desc": "do NOT edit this line"}

[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  # sock = "<SOCK>"
  # charset = "utf8"

  ## @param connect_timeout - number - optional - default: 10s
  # connect_timeout = "10s"

  ## Deprecated
  # service = "<SERVICE>"

  interval = "10s"

  ## @param inno_db
  innodb = true

  ## table_schema
  tables = []

  ## user
  users = []

  ## 开启数据库性能指标采集
  # dbm = false

  # [inputs.mysql.log]
  # #required, glob logfiles
  # files = ["/var/log/mysql/*.log"]

  ## glob filteer
  #ignore = [""]

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  #character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  #match = '''^(# Time|\d{4}-\d{2}-\d{2}|\d{6}\s+\d{2}:\d{2}:\d{2}).*'''

  ## grok pipeline script path
  #pipeline = "mysql.p"

  # [[inputs.mysql.custom_queries]]
  #   sql = "SELECT foo, COUNT(*) FROM table.events GROUP BY foo"
  #   metric = "xxxx"
  #   tagKeys = ["column1", "column1"]
  #   fieldKeys = ["column3", "column1"]

  ## 监控指标配置
  [inputs.mysql.dbm_metric]
    enabled = true

  ## 监控采样配置
  [inputs.mysql.dbm_sample]
    enabled = true

  [inputs.mysql.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`  // mysqlConf

) // const
*/
