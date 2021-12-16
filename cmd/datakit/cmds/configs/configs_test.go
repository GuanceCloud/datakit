package configs

import "testing"

// go test -v -timeout 30s -run ^TestReleaseFiles$ gitlab.jiagouyun.com/cloudcare-tools/datakit/includefiles/configs
func TestReleaseFiles(t *testing.T) {
	arr, err := ReleaseFiles()
	if err != nil {
		t.Error(err)
		return
	}

	for _, v := range arr {
		t.Log(v)
	}

	t.Log("TestReleaseFiles completed!")
}
