package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestReporter(t *testing.T) {
	m := Reporter{
		Status:  "info",
		Message: "demo",
	}

	tags := m.Tags()
	fields := m.Fields()

	assert.Equal(t, tags["source"], "datakit")
	assert.Equal(t, tags["status"], "info")
	assert.Equal(t, fields["message"], "demo")

	assert.Equal(t, m.escape("http://url?token=token_sdsf8sfsfsk"), "http://url?token=xxxxxx")

	// injection feed function
	m.feed = func(s1, s2 string, p []*Point, o *Option) error {
		assert.Equal(t, s1, "datakit")
		assert.Equal(t, s2, datakit.Logging)
		assert.Equal(t, len(p), 1)
		return nil
	}

	addReporter(m)
}
