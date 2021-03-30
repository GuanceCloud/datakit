package redis

import (
	"math"
)

func Round(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Cast(num*output)) / output
}

func Cast(num float64) int {
	return int(num + math.Copysign(0.5, num))
}