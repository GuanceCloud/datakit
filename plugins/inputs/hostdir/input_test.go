package hostdir

import (
	"os"
	"runtime"
	"testing"
)

func TestInput_Collect(t *testing.T) {
	str, _ := os.Getwd()
	i := Input{Dir: str, platform: runtime.GOOS}
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
	t.Log(i.collectCache[0])
}
