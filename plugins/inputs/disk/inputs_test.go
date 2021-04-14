package disk

import (
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	i := &Input{
		MountPoints: []string{},
		IgnoreFS: []string{
			"tmpfs", "devtmpfs",
			"devfs", "iso9660",
			"overlay", "aufs", "squashfs",
		},
		Tags: map[string]string{
			"tag1": "a",
		},
		diskStats: &PSDisk{},
	}
	for x := 0; x < 5; x++ {
		i.Collect()
		time.Sleep(time.Second * 1)
	}
	if len(i.collectCache) < 1 {
		t.Error("Failed to collect, no data returned")
	}
	tmap := map[string]bool{}
	for _, v := range i.collectCache {
		m := v.(*diskMeasurement)
		tmap[m.ts.String()] = true
	}
	if len(tmap) != 1 {
		t.Error("Need to clear collectCache.")
	}

}
