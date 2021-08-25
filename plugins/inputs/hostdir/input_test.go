package hostdir

import (
	"fmt"
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
	fmt.Println(i.collectCache[0])
}
