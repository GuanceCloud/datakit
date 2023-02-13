// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestXML(t *testing.T) {
	testCases := []struct {
		name     string
		in       string
		script   string
		expected string
		key      string
		fail     bool
	}{
		{
			name: "normal",
			in: `<entry>
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
			in: `<entry>
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
			in: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, '/OrderEvent/@actionCode', action_code)`,
			expected: "5",
			key:      "action_code",
		},
		{
			name: "node OrderNumber",
			in: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, '/OrderEvent/OrderNumber/text()', OrderNumber)`,
			expected: "ORD12345",
			key:      "OrderNumber",
		},
		{
			name: "node VendorNumber",
			in: ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`,
			script:   `xml(_, '/OrderEvent/VendorNumber/text()', VendorNumber)`,
			expected: "V11111",
			key:      "VendorNumber",
		},
		{
			name:     "not a valid XML",
			in:       `Not a valid XML`,
			script:   `xml(_, '/OrderEvent/VendorNumber/text()', VendorNumber)`,
			expected: "V11111",
			key:      "VendorNumber",
			fail:     true,
		},
		{
			name: "not a valid xpath expr",
			in: ` <OrderEvent actionCode = "5">
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

			pt := ptinput.GetPoint()
			ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				ptinput.PutPoint(pt)
				t.Fatal(errR)
			}

			r, ok := pt.Fields[tc.key]
			if !ok && tc.fail {
				t.Logf("[%d] failed as expected", idx)
				return
			}
			tu.Equals(t, true, ok)
			tu.Equals(t, tc.expected, r)
			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}
