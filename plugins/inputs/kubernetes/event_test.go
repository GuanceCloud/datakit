package kubernetes

import (
	"testing"
)

func TestEvent(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	err := i.initCfg()
	t.Log("error ---->", err)

	i.collectEvents()
}
