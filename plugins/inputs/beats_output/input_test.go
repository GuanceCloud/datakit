package beats_output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestParseListen$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/beats_output
func TestParseListen(t *testing.T) {
	cases := []struct {
		name        string
		listen      string
		expectError error
		expect      map[string]string
	}{
		{
			name:   "normal",
			listen: "tcp://0.0.0.0:5044",
			expect: map[string]string{
				"scheme": "tcp",
				"host":   "0.0.0.0:5044",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mVal, err := parseListen(tc.listen)
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.expect, mVal)
		})
	}
}
