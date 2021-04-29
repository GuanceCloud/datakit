package check

import (
	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"fmt"
)

var (
	l = logger.DefaultSLogger("check")
)

func CheckInputToml(name string, tomlcfg []byte) error {
	if c, ok := inputs.Inputs[name]; ok {
		dkinput := c()
		if err := toml.Unmarshal(tomlcfg, dkinput); err != nil {
			l.Errorf("toml.Unmarshal: %s", err.Error())
			return err
		}

		l.Debugf("toml %+#v", dkinput)
		return nil
		// TODO:
	}
	return fmt.Errorf("input %s not exist", name)
}
