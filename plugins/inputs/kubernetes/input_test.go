package kubernetes

import (
	"testing"
)

func TestMain(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	err := i.initCfg()
	t.Log("error ---->", err)

	i.Collect()

	for _, obj := range i.collectCache {
		point, err := obj.LineProto()
		if err != nil {
			t.Log("error ->", err)
		} else {
			t.Log("point ->", point.String())
		}
	}
}
