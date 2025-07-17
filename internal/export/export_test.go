// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	T "testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
)

func TestRenderBuf(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		md := []byte(`
{{ InstallCmd 4 (.WithPlatform "windows") }}
			`)

		p := &Params{}
		x, err := p.renderBuf(md)
		assert.NoError(t, err)
		t.Logf("\n%s", x)
	})

	t.Run(`multiline`, func(t *T.T) {
		md := []byte(`
{{ InstallCmd 4
(.WithPlatform "windows")
(.WithVersion "-1.2.3")
}}
			`)

		p := &Params{}
		x, err := p.renderBuf(md)
		assert.NoError(t, err)
		t.Logf("\n%s", x)
	})

	t.Run(`multiline-without-indent`, func(t *T.T) {
		md := []byte(`
{{ InstallCmd 0
(.WithPlatform "windows")
(.WithVersion "-1.2.3")
}}
			`)

		p := &Params{}
		x, err := p.renderBuf(md)
		assert.NoError(t, err)
		t.Logf("\n%s", x)
	})

	t.Run(`render-ui-steps`, func(t *T.T) {
		md := []byte(`{{ UISteps "1,2,3" ","}}`)

		p := &Params{}
		x, err := p.renderBuf(md)
		assert.NoError(t, err)
		assert.Equal(t, "**1** ➔ **2** ➔ **3**", string(x))
		t.Logf("%s", x)

		md = []byte(`{{ UISteps "   1,2   ,3   " "," }}`)

		x, err = p.renderBuf(md)
		assert.NoError(t, err)
		assert.Equal(t, "**1** ➔ **2** ➔ **3**", string(x))
		t.Logf("%s", x)
	})

	t.Run("ChangeManifests", func(t *T.T) {
		p := &Params{
			ChangeManifests: changes.MustLoadAllManifest(),
		}

		assert.NotNil(t, p.ChangeManifests)

		t.Logf("%+#v", p.ChangeManifests)

		md := []byte(`{{ if eq .ChangeManifests.K8sManifest nil }}
{{ else }}
## Kubernetes {#k8s}

当前版本： {{  .ChangeManifests.K8sManifest.Version }}

{{.ChangeManifests.K8sManifest.MDTable "zh" }}

{{ end }}

{{ if eq .ChangeManifests.HostManifest nil }}
{{ else }}
## 主机 {#host}

当前版本： {{  .ChangeManifests.HostManifest.Version }}

{{.ChangeManifests.HostManifest.MDTable "zh" }}

{{ end }}
		`)

		x, err := p.renderBuf(md)
		assert.NoError(t, err)
		t.Logf("%s", x)
	})
}
