package kubernetes

import (
	"context"
	"testing"
)

func TestCollectPersistentVolumes(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	i.lastErr = i.initCfg()
	ctx := context.Background()
	err := i.collectPersistentVolumes(ctx)

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

func TestCollectPersistentVolumeClaims(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	i.lastErr = i.initCfg()
	ctx := context.Background()
	err := i.collectPersistentVolumeClaims(ctx)

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

func TestCollectServices(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	i.lastErr = i.initCfg()
	ctx := context.Background()
	err := i.collectServices(ctx)

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

func TestCollectStatefulSets(t *testing.T) {
	i := &Input{
		Tags:           make(map[string]string),
		KubeConfigPath: "/Users/liushaobo/.kube/config",
	}

	i.lastErr = i.initCfg()
	ctx := context.Background()
	err := i.collectStatefulSets(ctx)

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
