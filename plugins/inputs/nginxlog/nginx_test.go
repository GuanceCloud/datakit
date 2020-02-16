package nginxlog

import (
	"log"
	"testing"
)

func TestNginxLogCfg(t *testing.T) {

	access_log := &NginxAccessLog{
		LogFile:     `/var/log/nginx/access.log`,
		Measurement: `nginx_access`,
	}

	access_log2 := &NginxAccessLog{
		LogFile:     `/var/log/nginx2/access.log`,
		Measurement: `nginx_access`,
	}

	error_log := &NginxErrorLog{
		LogFile:     `/var/log/nginx/error.log`,
		Measurement: `nginx_error`,
	}

	cfg := &NginxLogConfig{
		AccessLogs: []*NginxAccessLog{access_log, access_log2},
		ErrorLogs:  []*NginxErrorLog{error_log},
	}

	// cfgdata, err := toml.Marshal(cfg)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// ioutil.WriteFile("./nginxlog.toml", cfgdata, 0664)

	telcfg, err := cfg.ToTelegraf(``)
	if err == nil {
		log.Println(telcfg)
	}

	//log.Println(string(cfgdata))

}
