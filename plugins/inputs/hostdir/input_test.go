package hostdir

import (
	"os"
	"runtime"
	"testing"
)

func TestInput_Collect(t *testing.T) {
	str, _ := os.Getwd()
	i := Input{Dir: str, platform: runtime.GOOS}
	err := i.Collect()
	if err != nil {
		t.Log(err)
	}
	t.Log(i.collectCache[0])
}
