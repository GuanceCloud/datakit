package configtemplate

import (
	"log"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func TestCfgTemp(t *testing.T) {

	logger.DefaultSLogger("main")

	ct := NewCfgTemplate("./result")
	err := ct.InstallConfigs("file://./conf.d/test.tar.gz")
	if err != nil {
		log.Fatalf("%s", err)
	}
}
