package mongodboplog

import (
	"testing"
)

func TestRewriteCategory(t *testing.T) {
	testcase := []string{
		"metric",
		"logging",
		"invalid",
		"",
	}

	for _, tc := range testcase {
		m := &Mongodboplog{
			Category: tc,
		}
		t.Logf("source: %v\n", m.Category)

		m.rewriteCategory()
		t.Logf("valid:  %v\n\n", m.Category)
	}
}
