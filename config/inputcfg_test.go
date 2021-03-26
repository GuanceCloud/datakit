package config

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestAddInput(t *testing.T) {
	// FIXME: 此处只支持测试 aliyunobject、host_processes 和 tailf 采集器
	// 正常情况下，所有 init() 函数都在单元测试执行前完成，但是实测采集器只有上述三个执行了 init()
	// 导致 inputs.Inputs 采集器列表中只有这三个
	// 原因未知，待查

	// inputs.Inputs: map[aliyunobject:0x2464880 host_processes:0x24858b0 tailf:0x170d070]
	// t.Logf("inputs.Inputs: %v\n", inputs.Inputs)

	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	t.Log(inputs.Inputs)

	if err := addInput("tailf", file.Name()); err != nil {
		t.Fatal(err)
	}

	t.Log(inputs.InputsInfo)
}
