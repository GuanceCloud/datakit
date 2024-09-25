package mathtoolkit

import "math"

func Trunc(x float64) int64 {
	return int64(math.Trunc(x))
}

func Floor(x float64) int64 {
	return int64(math.Floor(x))
}

func Ceil(x float64) int64 {
	return int64(math.Ceil(x))
}
