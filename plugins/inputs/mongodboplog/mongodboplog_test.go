package mongodboplog

import (
	"testing"
)

func TestRewriteCategories(t *testing.T) {
	testcase := [][]string{
		{"metric", "logging", "invalid", "metric"},
		{},
		{"invalid"},
		{"metric"},
		{"logging"},
	}

	for _, tc := range testcase {
		m := &Mongodboplog{
			Categories: tc,
		}
		t.Logf("source: %v\n", m.Categories)

		m.rewriteCategories()
		t.Logf("valid:  %v\n\n", m.Categories)
	}
}
