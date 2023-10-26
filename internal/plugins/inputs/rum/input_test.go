// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitReaderClose(t *testing.T) {
	r := io.NopCloser(strings.NewReader("hello world!!!!!"))

	lr := newLimitReader(r, 10)

	c, err := io.ReadAll(lr)

	assert.Error(t, err)
	assert.Len(t, c, 10)
	assert.ErrorIs(t, err, errLimitReader)

	t.Log(err, len(c), string(c), errors.Is(err, errLimitReader))
}

func TestLookupCDNName(t *testing.T) {
	assert.True(t, isDomainName("xxxxx.xxxxxx"))

	_, _, err := lookupCDNName("xxxxx.xxxxxx")
	assert.Error(t, err)

	{
		_, _, err := lookupCDNName("www-static.qbox.me")
		assert.NoError(t, err)
	}

	{
		_, _, err := lookupCDNName("tbcache.com")
		assert.NoError(t, err)
	}

	{
		_, _, err := lookupCDNName("go.dev")
		assert.Error(t, err)
	}

	{
		_, _, err := lookupCDNName("ucc.alicdn.com")
		assert.NoError(t, err)
	}

	{
		_, _, err := lookupCDNName("static.guance.com")
		assert.NoError(t, err)
	}

	{
		_, _, err := lookupCDNName("res.vmallres.com")
		assert.NoError(t, err)
	}
}
