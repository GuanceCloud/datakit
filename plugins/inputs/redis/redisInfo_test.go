package redis

import (
	"testing"
	// "github.com/stretchr/testify/assert"
)

func TestCollectInfoMetrics(t *testing.T) {
	info := CollectInfoMeasurement(cli)

	fmt.Println(info.fields)
}