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

	"github.com/GuanceCloud/cliutils/testutil"
)

func TestLimitReader_Close(t *testing.T) {
	r := io.NopCloser(strings.NewReader("hello world!!!!!"))

	lr := newLimitReader(r, 10)

	c, err := io.ReadAll(lr)

	t.Log(err, len(c), string(c), errors.Is(err, errLimitReader))

	testutil.NotOk(t, err, "")
	testutil.Equals(t, 10, len(c))
	testutil.Equals(t, true, errors.Is(err, errLimitReader))
}

func TestLookupCDNName(t *testing.T) {
	ok := isDomainName("xxxxx.xxxxxx")
	testutil.Assert(t, ok, "")

	_, _, err := lookupCDNName("xxxxx.xxxxxx")
	t.Log(err)
	testutil.NotOk(t, err, "")

	{
		cname, cdn, err := lookupCDNName("www-static.qbox.me")
		t.Log(cname, cdn)
		testutil.Ok(t, err)
	}

	{
		cname, cdn, err := lookupCDNName("tbcache.com")
		t.Log(cname, cdn)
		testutil.Ok(t, err)
	}

	{
		cname, cdn, err := lookupCDNName("go.dev")
		t.Log(cname, cdn, err)
		testutil.NotOk(t, err, "")
	}

	{
		cname, cdn, err := lookupCDNName("ucc.alicdn.com")
		t.Log(cname, cdn)
		testutil.Ok(t, err)
	}

	{
		cname, cdn, err := lookupCDNName("static.guance.com")
		t.Log(cname, cdn)
		testutil.Ok(t, err)
	}

	{
		cname, cdn, err := lookupCDNName("res.vmallres.com")
		t.Log(cname, cdn)
		testutil.Ok(t, err)
	}

}
