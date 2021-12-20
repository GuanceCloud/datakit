package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestUserAgent(t *testing.T) {
	cases := []struct {
		name     string
		pl, in   string
		expected map[string]interface{}
		fail     bool
	}{
		{
			name: "normal",
			pl:   `json(_, userAgent) user_agent(userAgent)`,
			in: `
{
   "userAgent" : "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
   "second"    : 2,
   "third"     : "abc",
   "forth"     : true
}
`,
			expected: map[string]interface{}{
				"isMobile":   false,
				"isBot":      false,
				"os":         "Windows 7",
				"browser":    "Chrome",
				"browserVer": "36.0.1985.125",
				"engine":     "AppleWebKit",
				"engineVer":  "537.36",
				"ua":         "Windows",
			},
			fail: false,
		},
		{
			name: "normal",
			pl:   `json(_, userAgent) user_agent(userAgent)`,
			in: `
{
    "userAgent" : "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15"
}
`,
			expected: map[string]interface{}{
				"isMobile":   false,
				"isBot":      false,
				"os":         "Intel Mac OS X 10_15_7",
				"browser":    "Safari",
				"browserVer": "15.1",
				"engine":     "AppleWebKit",
				"engineVer":  "605.1.15",
				"ua":         "Macintosh",
			},
			fail: false,
		},

		{
			name: "normal",
			pl:   `json(_, userAgent) user_agent(agent)`,
			in: `
{
    "userAgent" : "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15"
}
`,
			expected: map[string]interface{}{},
			fail:     false,
		},

		{
			name: "invalid arg type",
			pl:   `json(_, userAgent) user_agent("userAgent")`,
			in: `
		{
		   "userAgent" : "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15"
		}
		`,
			fail: true,
		},

		{
			name: "too many args",
			pl:   `json(_, userAgent) user_agent(userAgent, someArg)`,
			in: `
		{
		   "userAgent" : "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15"
		}
		`,
			fail: true,
		},
	}
	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}

			err = runner.Run(tc.in)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Error(err)
				}
			} else {
				t.Log(runner.Result())
				fieldsToCompare := make(map[string]interface{})
				for k := range tc.expected {
					fieldsToCompare[k], _ = runner.GetContent(k)
				}
				tu.Equals(t, tc.expected, fieldsToCompare)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
