package container

import (
	"testing"
)

func TestGetContainerPodName(t *testing.T) {
	mock := Kubernetes{
		URL: "http://127.0.0.1:10255",
	}

	if err := mock.Init(); err != nil {
		t.Fatal(err)
	}

	var cases = []struct {
		id string
	}{
		{
			"cf063df157fed1976dfff6b47e7ba55a1ce36385715b859bd03d4acc9b92690c",
		},
	}

	for idx, tc := range cases {
		name, err := mock.GetContainerPodName(containerIDPrefix + tc.id)
		if err != nil {
			t.Error(err)
		}
		t.Logf("[%d] container_id:%s pod_name:%s\n", idx, tc.id, name)
	}
}
