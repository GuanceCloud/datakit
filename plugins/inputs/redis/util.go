package redis

import (
	"math"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func Round(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Cast(num*output)) / output
}

func Cast(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func Conv(val interface{}, datatype string) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	switch datatype {
	case inputs.Float:
		res, err = cast.ToFloat64E(val)
	case inputs.Int:
		res, err = cast.ToInt64E(val)
	case inputs.Bool:
		res, err = cast.ToBoolE(val)
	case inputs.String:
		res, err = cast.ToStringE(val)
	}

	return res, err
}

func IsSlicesHave(s []int, index int) bool {
	for _, i := range s {
		if i == index {
			return true
		}
	}
	return false
}
