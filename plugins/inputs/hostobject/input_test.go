package hostobject

import "testing"

// go test -v -timeout 30s -run ^TestCollect$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject
func TestCollect(t *testing.T) {
	ho := DefaultHostObject()
	pts, err := ho.Collect()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(pts)

	if len(pts) == 0 {
		t.Error("collect empty!")
		return
	}

	for category := range pts {
		if category != "/v1/write/object" {
			t.Errorf("category not object: %s", category)
			return
		}
	}

	t.Log("TestCollect succeeded!")
}
