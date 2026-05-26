// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

func (c *Canary) CheckLast(category Category, feed *Feed, storageIndex string) (bool, error) {
	startTime := time.Now()
	dql := c.BuildDQL(category, storageIndex)

	// Track status for metrics
	status := "error"
	var found bool

	// Defer to record metrics on function exit (only if enabled)
	defer func() {
		if c.EnableMetrics {
			duration := time.Since(startTime).Seconds()
			dqlQueryDurationVec.WithLabelValues(string(category), storageIndex, status).Observe(duration)
			dqlQueryTotalVec.WithLabelValues(string(category), storageIndex, status).Inc()
		}
	}()

	queryReq := &httpapi.QueryRaw{
		Token: config.GetToken(),
		Queries: []*httpapi.SingleQuery{
			{
				Query: dql,
			},
		},
	}
	body, err := queryReq.JSON()
	if err != nil {
		return false, err
	}
	resp, err := config.Cfg.Dataway.DQLQuery(body)
	if err != nil {
		return false, fmt.Errorf("DQL query failed: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read response failed: %w", err)
	}
	var dqlResp DQLQueryResponse
	if err := json.Unmarshal(respBody, &dqlResp); err != nil {
		return false, fmt.Errorf("unmarshal response failed: %w", err)
	}

	result, ok := ParseLast(&dqlResp)
	if !ok {
		status = "not_found"
		return false, nil
	}

	found = feed.TimeMs == result.TimeMs && feed.Round == result.Round
	if found {
		status = "success"
	} else {
		status = "not_found"
	}

	return found, nil
}
