package config

import (
	"log"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestLoadMainCfg(t *testing.T) {

	c := newDefaultCfg()
	if err := c.LoadMainConfig(); err != nil {
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

	err = initMainCfg(maincfg, `test.conf`)
	if err != nil {
		log.Fatalf("%s", err)
	}

	initTelegrafSamples()
}
