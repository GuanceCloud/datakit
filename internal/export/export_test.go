// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderBuf(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		md := []byte(`
{{ InstallCmd 4 (.WithPlatform "windows") }}
			`)

		p := &Params{}
		x, err := renderBuf(md, p)
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
		x, err := renderBuf(md, p)
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
		x, err := renderBuf(md, p)
		assert.NoError(t, err)
		t.Logf("\n%s", x)
	})
}
