package config

import (
	"context"
	"log"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestLoadMainCfg(t *testing.T) {

	c := NewConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := c.LoadMainConfig(ctx, `test.conf`); err != nil {
		t.Errorf("%s", err)
	}
}

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
