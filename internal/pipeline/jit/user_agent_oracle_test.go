// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"reflect"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
)

func TestPipelineGoUserAgentOracle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]any
	}{
		{
			name:  "chrome",
			input: "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36",
			want: map[string]any{
				"isMobile": false, "isBot": false, "os": "Intel Mac OS X 11_1_0",
				"browser": "Chrome", "browserVer": "89.0.4389.72",
				"engine": "AppleWebKit", "engineVer": "537.36", "ua": "Macintosh",
			},
		},
		{
			name:  "iphone",
			input: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Version/17.0 Mobile/15E148 Safari/604.1",
			want: map[string]any{
				"isMobile": true, "isBot": false, "os": "CPU iPhone OS 17_0 like Mac OS X",
				"browser": "Safari", "browserVer": "17.0",
				"engine": "AppleWebKit", "engineVer": "605.1.15", "ua": "iPhone",
			},
		},
		{
			name:  "curl",
			input: "curl/8.0",
			want: map[string]any{
				"isMobile": false, "isBot": false, "os": "",
				"browser": "curl", "browserVer": "8.0",
				"engine": "", "engineVer": "", "ua": "",
			},
		},
		{
			name:  "okhttp",
			input: "okhttp/4.2.2",
			want: map[string]any{
				"isMobile": true, "isBot": false, "os": "",
				"browser": "OkHttp", "browserVer": "4.2.2",
				"engine": "", "engineVer": "", "ua": "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, types := funcs.UserAgentHandle(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("UserAgentHandle(%q) = %#v, want %#v", test.input, got, test.want)
			}
			if len(types) != len(test.want) {
				t.Fatalf("UserAgentHandle type count = %d, want %d", len(types), len(test.want))
			}
		})
	}
}
