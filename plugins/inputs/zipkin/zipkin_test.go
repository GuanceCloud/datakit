// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package zipkin

import (
	"fmt"
	"testing"
)

func TestInt2ip(t *testing.T) {
	ip := int2ip(3232235778)
	for _, b := range ip {
		fmt.Printf("%d ", b)
	}
}
