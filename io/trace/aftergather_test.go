package trace

import (
	"fmt"
	"testing"
)

func TestBuildPoint(t *testing.T) {
	for i := 0; i < 100; i++ {
		if pt, err := BuildPoint(randDatakitSpan(t), false); err != nil {
			t.Error(err.Error())
			t.FailNow()
		} else {
			fmt.Println(pt.String())
		}
	}
}
