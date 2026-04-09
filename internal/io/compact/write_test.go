// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	T "testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestIsGZip(t *T.T) {
	t.Run("is-gzip", func(t *T.T) {
		data := []byte("hello world")

		gz, err := datakit.GZip(data)
		assert.NoError(t, err)

		assert.Equal(t, GzipFlag(1), IsGzip(gz))
	})
}
