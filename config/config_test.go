package config

import (
	"log"
	"testing"

	uuid "github.com/satori/go.uuid"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/all"
)

func TestInitCfg(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	uid, err := uuid.NewV4()

	maincfg := &MainConfig{
		UUID:      "dkit_" + uid.String(),
		FtGateway: "http://localhost",
		Log:       "./datakit.log",
		LogLevel:  "debug",
		ConfigDir: "conf.d",
	}

	err = InitMainCfg(maincfg, `test.conf`)
	if err != nil {
		log.Fatalf("%s", err)
	}

	InitTelegrafSamples()

	if err = CreatePluginConfigs("./testconf.d", false); err != nil {
		log.Fatalf("%s", err)
	}
}

func TestLoadCfg(t *testing.T) {
	// c := NewConfig()
	// if err := c.LoadConfig(`test.conf`, `testconf.d`); err != nil {
	// 	log.Fatalf("%s", err)
	// }

	// log.Printf("ok: %#v", *c)
}
