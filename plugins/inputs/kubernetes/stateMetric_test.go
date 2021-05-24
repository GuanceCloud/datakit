package kubernetes

import (
	"testing"
)

func TestCollectStateMetric(t *testing.T) {
	i := &Input{
		StateUrl: "http://172.16.2.42:32136/metrics",
	}

	i.lastErr = i.initCfg()
	err := i.collectStateMetric()

	t.Log("collect error -->", err)

	// for _, obj := range i.collectCache {
	// 	point, _ := obj.LineProto()
	// 	t.Log("point ->", point.String())
	// 	// if err != nil {
	// 	// 	t.Log("error ->", err)
	// 	// } else {
	// 	// 	t.Log("point ->", point.String())
	// 	// }
	// }
}
