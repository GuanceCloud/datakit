package nginx

import (
	"log"
	"testing"
)

func TestNginxCfg(t *testing.T) {

	// cfg := &NginxConfig{
	// 	Status: []*NginxStatus{
	// 		&NginxStatus{
	// 			Urls: "http://localhost",
	// 		},
	// 	},

	// 	PlusStatus: []*NginxStatus{
	// 		&NginxStatus{
	// 			Urls: "http://localhostplus",
	// 		},
	// 	},
	// }

	//ioutil.WriteFile("./nginx.toml", []byte(cfg.SampleConfig()), 0664)

	// dt, err := toml.Marshal(cfg)

	// if err == nil {
	// 	ioutil.WriteFile("./aa.toml", dt, 0664)
	// }

	var cfg NginxConfig

	if err := cfg.Load("./a.toml"); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("-- %#v", cfg)
		ts, e := cfg.ToTelegraf()
		if e == nil {
			log.Println(ts)
		}
	}

	// if err := cfg.Load("./aa.toml"); err != nil {
	// 	log.Fatalln(err)
	// } else {
	// 	log.Printf("%#v", cfg)
	// }
}

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

	telcfg, err := cfg.ToTelegraf()
	if err == nil {
		log.Println(telcfg)
	}

	//log.Println(string(cfgdata))

}
