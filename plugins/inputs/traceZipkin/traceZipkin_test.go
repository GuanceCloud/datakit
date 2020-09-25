package traceZipkin

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
