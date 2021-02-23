package config

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestAddInput(t *testing.T) {
	// FIXME: 此处只支持测试 aliyunobject、host_processes 和 tailf 采集器
	// 正常情况下，所有 init() 函数都在单元测试执行前完成，但是实测采集器只有上述三个执行了 init()
	// 导致 inputs.Inputs 采集器列表中只有这三个
	// 原因未知，待查

	fp := filepath.Join("/tmp", "tailf")

	t.Log(inputs.Inputs)

	if err := addInput("tailf", fp); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fp)

	t.Log(inputs.InputsInfo)
}
