package kubernetes

import (
	"context"
	"testing"
)

func TestCollectIngress(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	i.lastErr = i.initCfg()
	ctx := context.Background()
	err := i.collectIngress(ctx)

	t.Log("error ---->", err)

	for _, m := range i.collectCache {
		point, err := m.LineProto()
		if err != nil {
			t.Log("error ->", err)
		} else {
			t.Log("point ->", point.String())
		}
	}
}
