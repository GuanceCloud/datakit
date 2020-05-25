package config

import (
	//"log"
	"testing"
	//uuid "github.com/satori/go.uuid"
)

func TestLoadMainCfg(t *testing.T) {

	c := newDefaultCfg()
	if err := c.LoadMainConfig(); err != nil {
		t.Errorf("%s", err)
	}
}

func TestInitCfg(t *testing.T) {
	// TODO
}
