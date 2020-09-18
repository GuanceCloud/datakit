package huaweicloud

import (
	"fmt"
)

const (
	cesMestricPath      = "/V1.0/%s/batch-query-metric-data"
	cesBatchMestricPath = "/V1.0/%s/metric-data"
)

func (c *HWClient) CESGetMetric(namespace, metricname string, filter string, period int, from, to int64, dims []string) ([]byte, error) {
	querys := map[string]string{
		"namespace":   namespace,
		"metric_name": metricname,
		"from":        fmt.Sprintf("%d", from),
		"to":          fmt.Sprintf("%d", to),
		"period":      fmt.Sprintf("%d", period),
		"filter":      filter,
	}

	for i, d := range dims {
		querys[fmt.Sprintf("dim.%d", i)] = d
	}
	resp, err := c.Request("GET", cesMestricPath, querys, nil)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}
	//c.logger.Debugf("%s", string(resp))

	return resp, nil
}

func (c *HWClient) CESGetBatchMetrics(jsonReq []byte) ([]byte, error) {
	resp, err := c.Request("POST", cesBatchMestricPath, nil, jsonReq)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}
	return resp, nil
}
