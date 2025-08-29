// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	T "testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestUnmarshalInputConf(t *T.T) {
	t.Run(`tags-rename`, func(t *T.T) {
		conf := `
[[inputs.prom]]
measurement_name = "guancedb"
url = "http://$IP:8480/metrics"
interval = "10s"
source = "guance-scopedb-insert"
[inputs.prom.tags]
	source = "scopedb"
	pod = "$PODNAME"
[inputs.prom.tags_rename.mapping]
	namespace = "index_namespace"`

		creators := map[string]inputs.Creator{
			"prom": func() inputs.Input {
				return &Input{}
			},
		}

		ret, err := config.LoadSingleConf(conf, creators)
		if err != nil {
			t.Logf("doLoadConf: %s", err)
		}

		for k, v := range ret {
			t.Logf("%q: %+#v", k, v[0])
			t.Logf("rename: %+#v", v[0].Input.(*Input).TagsRename)
		}
	})
}
