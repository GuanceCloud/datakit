package kubernetes

import (
	"testing"
)

func TestCollectObject(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	i.lastErr = i.initCfg()
	i.collectPodObject("podObject")

	for _, obj := range i.collectObjectCache {
		point, err := obj.LineProto()
		if err != nil {
			t.Log("error ->", err)
		} else {
			t.Log("point ->", point.String())
		}
	}
}
