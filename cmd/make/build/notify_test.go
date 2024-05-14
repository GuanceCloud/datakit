// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	T "testing"
)

func Test_buildNotifyContent(t *T.T) {
	t.Run("basic", func(t *T.T) {
		// verify output text
		t.Logf("content:\n%s",
			buildNotifyContent("1.2.3", "testing.cdn.com", ReleaseLocal,
				[]string{
					`linux/386`,
					`linux/amd64`,
					`linux/arm`,
					`linux/arm64`,
					`darwin/amd64`,
					"windows/amd64",
					"windows/386",
				}),
		)
	})

	t.Run("testing", func(t *T.T) {
		// verify output text
		t.Logf("content:\n%s",
			buildNotifyContent("1.2.3", "testing.cdn.com", ReleaseTesting,
				[]string{
					`linux/386`,
					`linux/amd64`,
					`linux/arm`,
					`linux/arm64`,
					`darwin/amd64`,
					"windows/amd64",
					"windows/386",
				}),
		)
	})
}
