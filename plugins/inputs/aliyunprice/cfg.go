package aliyunprice

import "time"

const (
	defaultInterval = 5 * time.Minute

	globalConfig = `
#access_key_id = ''
#access_key_secret = ''
#region_id = ''
`
)

type (
	priceMod interface {
		handleTags(map[string]string) map[string]string
		handleFields(map[string]interface{}) map[string]interface{}
	}
)
