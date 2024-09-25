package units

import (
	"fmt"
	"time"
)

func ToTime(quantity IQuantity) (time.Time, error) {
	var t time.Time
	if quantity.Unit() == nil {
		return t, fmt.Errorf("nil unit")
	}
	if kind := quantity.Unit().Kind; kind != TimeStamp {
		return t, fmt.Errorf("not kind of timestamp: %q", kind.String())
	}

	quantity, err := quantity.In(UnixNano)
	if err != nil {
		return t, fmt.Errorf("unable to be converted to unixnano: %w", err)
	}

	return time.Unix(0, quantity.IntValue()), nil
}
