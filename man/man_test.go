package man

import (
	"testing"
)

func TestBuildMarkdownManual(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			md, err := BuildMarkdownManual(tc.doc, &Option{})
			if err != nil {
				t.Error(err)
			}

			t.Log(string(md))
		})
	}
}
