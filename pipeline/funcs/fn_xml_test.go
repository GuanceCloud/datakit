package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestXML(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		script   string
		expected string
		key      string
		fail     bool
	}{
		{
			name: "normal",
			input: `<entry>
        <fieldx>valuex</fieldx>
        <fieldy>...</fieldy>
        <fieldz>...</fieldz>
        <fieldarray>
            <fielda>element_a_1</fielda>
            <fielda>element_a_2</fielda>
        </fieldarray>
    </entry>`,
			script:   `xml(_, '/entry/fieldx/text()', new_name)`,
			expected: "valuex",
			key:      "new_name",
		},
		{
			name: "select array field",
			input: `<entry>
        <fieldx>valuex</fieldx>
        <fieldy>...</fieldy>
        <fieldz>...</fieldz>
        <fieldarray>
            <fielda>element_a_1</fielda>
            <fielda>element_a_2</fielda>
        </fieldarray>
    </entry>`,
			script:   `xml(_, '/entry/fieldarray//fielda[1]/text()', field_a_1)`,
			expected: "element_a_1",
			key:      "field_a_1",
		},
		{
			name: "select attribute",
			input: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, '/OrderEvent/@actionCode', action_code)`,
			expected: "5",
			key:      "action_code",
		},
		{
			name: "node OrderNumber",
			input: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, '/OrderEvent/OrderNumber/text()', OrderNumber)`,
			expected: "ORD12345",
			key:      "OrderNumber",
		},
		{
			name: "node VendorNumber",
			input: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, '/OrderEvent/VendorNumber/text()', VendorNumber)`,
			expected: "V11111",
			key:      "VendorNumber",
		},
		{
			name:     "not a valid XML",
			input:    `Not a valid XML`,
			script:   `xml(_, '/OrderEvent/VendorNumber/text()', VendorNumber)`,
			expected: "V11111",
			key:      "VendorNumber",
			fail:     true,
		},
		{
			name: "not a valid xpath expr",
			input: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, 'invalid xpath expr', VendorNumber)`,
			expected: "V11111",
			key:      "VendorNumber",
			fail:     true,
		},
	}

	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.input,
				}, "message", time.Now())

			tu.Equals(t, nil, err)
			tu.Equals(t, nil, ret.Error)

			r, ok := ret.Fields[tc.key]
			if !ok && tc.fail {
				t.Logf("[%d] failed as expected", idx)
				return
			}
			tu.Equals(t, true, ok)
			tu.Equals(t, tc.expected, r)
			t.Logf("[%d] PASS", idx)
		})
	}
}
