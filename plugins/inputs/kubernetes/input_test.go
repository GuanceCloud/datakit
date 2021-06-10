package kubernetes

import (
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/stretchr/testify/assert"
	// "gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"testing"
)

func TestInitCfg(t *testing.T) {
	i := &Input{
		Tags:        make(map[string]string),
		URL:         "https://172.16.2.41:6443",
		BearerToken: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/token",
		ClientConfig: tls.ClientConfig{
			TLSCA: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/ca_crt.pem",
		},
	}

	err := i.initCfg()
	if err != nil {
		t.Log("error ---->", err)
		return
	}

	// 通过 ServerVersion 方法来获取版本号
	versionInfo, err := i.client.ServerVersion()
	if err != nil {
		assert.Error(t, err, "")
	}

	t.Log("version ==>", versionInfo.String())
}

func TestInitCfgErr(t *testing.T) {
	i := &Input{
		Tags:        make(map[string]string),
		URL:         "",
		BearerToken: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/token",
		ClientConfig: tls.ClientConfig{
			TLSCA: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/ca_crt.pem",
		},
	}

	err := i.initCfg()
	if err != nil {
		t.Log("error ---->", err)
		return
	}

	// 通过 ServerVersion 方法来获取版本号
	versionInfo, err := i.client.ServerVersion()
	if err != nil {
		assert.Error(t, err, "")
	}

	t.Log("version ==>", versionInfo.String())
}

func TestLoadCfg(t *testing.T) {
	arr, err := config.LoadInputConfigFile("./cfg.conf", func() inputs.Input {
		return &Input{}
	})

	if err != nil {
		t.Fatalf("%s", err)
	}

	kube := arr[0].(*Input)

	t.Log("url ---->", kube.URL)
	t.Log("token ---->", kube.BearerToken)
	t.Log("ca ---->", kube.TLSCA)
}

func TestRun(t *testing.T) {
	arr, err := config.LoadInputConfigFile("./cfg.conf", func() inputs.Input {
		return &Input{}
	})

	if err != nil {
		t.Fatalf("%s", err)
	}

	kube := arr[0].(*Input)

	err = kube.initCfg()
	if err != nil {
		t.Log("init config error ---->", err)
		return
	}

	err = kube.Collect()
	t.Log("collect data error ---->", err)

	for k, ms := range kube.collectCache {
		t.Log("collect resource type", k)
		for _, m := range ms {
			point, err := m.LineProto()
			if err != nil {
				t.Log("error ->", err)
			} else {
				t.Log("point ->", point.String())
			}
		}
	}
}
