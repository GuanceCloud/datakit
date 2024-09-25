package parsetoolkit

import (
	"fmt"
	"math"
	"strconv"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/pprof"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
	"github.com/google/pprof/profile"
)

func GetLabel(smp *profile.Sample, key string) string {
	if labels, ok := smp.Label[key]; ok {
		for _, label := range labels {
			if label != "" {
				return label
			}
		}
	}
	return ""
}

func GetStringLabel(smp *profile.Sample, key string) string {
	if span := GetLabel(smp, key); span != "" {
		return span
	}
	if span, ok := GetNumLabel(smp, key); ok {
		return strconv.FormatUint(uint64(span), 10)
	}
	return ""
}

func GetNumLabel(smp *profile.Sample, key string) (int64, bool) {
	if values, ok := smp.NumLabel[key]; ok {
		if len(values) > 0 {
			// Find none zero value at first
			for _, v := range values {
				if v != 0 {
					return v, true
				}
			}
			return values[0], true
		}
	}
	return 0, false
}

func CalcPercentOfAggregator() {

}

func CalcPercentAndQuantity(frame *pprof.Frame, total int64) {
	if frame == nil {
		return
	}

	if total <= 0 {
		frame.Percent = "100"
	} else {
		frame.Percent = fmt.Sprintf("%.2f", float64(frame.Value)/float64(total)*100)
	}

	if frame.Unit != nil {
		frame.Quantity = frame.Unit.Quantity(frame.Value)

		// 转成默认单位
		if frame.Unit.Kind == quantity.Memory && frame.Unit != quantity.Byte {
			frame.Value, _ = frame.Quantity.IntValueIn(quantity.Byte)
			frame.Unit = quantity.Byte
		} else if frame.Unit.Kind == quantity.Duration && frame.Unit != quantity.MicroSecond {
			frame.Value, _ = frame.Quantity.IntValueIn(quantity.MicroSecond)
			frame.Unit = quantity.MicroSecond
		}
	}

	for _, subFrame := range frame.SubFrames {
		CalcPercentAndQuantity(subFrame, total)
	}
}

func FormatDuration(nanoseconds int64) string {
	ms := int64(math.Round(float64(nanoseconds) / 1000_000))
	if ms < 1000 {
		return fmt.Sprintf("%d%s", ms, "ms")
	}
	if ms < 60_000 {
		seconds := int64(math.Round(float64(ms) / 1000))
		return fmt.Sprintf("%d%s", seconds, "s")
	}

	minutes := int64(math.Round(float64(ms) / 60_000))
	return fmt.Sprintf("%d%s", minutes, "minute")
}
