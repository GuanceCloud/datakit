// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package xfsquota

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuotaOutput(t *testing.T) {
	cases := []struct {
		in string
	}{
		/*{
					in: `
		User quota on /mnt/stu01 (/dev/mapper/class-stu01)
		                               Blocks                                          Inodes
		User ID          Used       Soft       Hard    Warn/Grace           Used       Soft       Hard    Warn/ Grace
		---------- -------------------------------------------------- --------------------------------------------------
		root                0          0          0     00 [--------]          3          0          0     00 [--------]
		zhangsan        81920      51200      81920     00  [6 days]          1          6          8     00 [--------]

		Group quota on /mnt/stu01 (/dev/mapper/class-stu01)
		                               Blocks                                          Inodes
		Group ID         Used       Soft       Hard    Warn/Grace           Used       Soft       Hard    Warn/ Grace
		---------- -------------------------------------------------- --------------------------------------------------
		root                0          0          0     00 [--------]          3          0          0     00 [--------]
		zhangsan        81920          0          0     00 [--------]          1          0          0     00 [--------]

		`,
				},*/
		{
			in: `
Project quota on /mnt/stu01 (/dev/mapper/class-stu01)
                               Blocks
Project ID          Used       Soft       Hard    Warn/Grace
---------- --------------------------------------------------
root                0          0          0     00 [--------]
zhangsan        81920      51200      81920     00  [6 days]
`,
		},
	}

	for _, tc := range cases {
		res, err := parseQuotaOutput(tc.in)
		assert.NoError(t, err)

		fmt.Println(res)
	}
}
