package tailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreFiles(t *testing.T) {
	var testcases = []struct {
		ignore, in, out []string
		fail            bool
	}{
		{
			ignore: []string{"/tmp/abc"},
			in:     []string{"/tmp/123"},
			out:    []string{"/tmp/123"},
			fail:   false,
		},
		{
			ignore: []string{"/tmp/*"},
			in:     []string{"/tmp/123"},
			out:    []string{},
			fail:   false,
		},
		{
			ignore: []string{"C:/Users/admin/Desktop/tmp/*"},
			in:     []string{"C:/Users/admin/Desktop/tmp/123"},
			out:    []string{},
			fail:   false,
		},
	}

	for _, tc := range testcases {
		p := NewProvider()
		p.list = tc.in

		result, err := p.IgnoreFiles(tc.ignore).Result()
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, tc.out, result)
	}
}
