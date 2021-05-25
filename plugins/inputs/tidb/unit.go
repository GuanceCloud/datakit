package tidb

import (
	"strconv"
	"strings"
)

var units = [...]string{"KiB", "MiB", "GiB", "TiB", "PiB"}

const (
	KiB = 1
	MiB = KiB << 10
	GiB = MiB << 10
	TiB = GiB << 10
	PiB = TiB << 10
)

func toKiB(s string) float64 {
	for _, unit := range units {
		if strings.HasSuffix(s, unit) {
			f, err := strconv.ParseFloat(strings.TrimSuffix(s, unit), 64)
			if err != nil {
				f = -1
			}

			switch unit {
			case "KiB":
				f *= KiB
			case "MiB":
				f *= MiB
			case "GiB":
				f *= GiB
			case "TiB":
				f *= TiB
			case "PiB":
				f *= PiB
			}

			return f
		}
	}
	return -1
}
