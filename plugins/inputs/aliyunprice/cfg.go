package aliyunprice

import "time"

const (
	defaultInterval = 5 * time.Minute
)

type (
	priceMod interface {
		handleTags(map[string]string) map[string]string
		handleFields(map[string]interface{}) map[string]interface{}
	}
)
