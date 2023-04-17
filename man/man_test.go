// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package man

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMarkdownManual(t *T.T) {
	cases := []struct {
		name string
		doc  string
	}{
		{
			name: "input-doc-logfwdserver",
			doc:  "logfwdserver",
		},
		{
			name: "input-doc-cpu",
			doc:  "cpu",
		},

		{
			name: "non-input-doc",
			doc:  "datakit-conf",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			docs, err := BuildMarkdownManual(tc.doc, &Option{})
			if err != nil {
				t.Error(err)
			}

			for k, md := range docs {
				t.Logf("%s:\n%s", k, string(md))
			}
		})
	}
}

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
