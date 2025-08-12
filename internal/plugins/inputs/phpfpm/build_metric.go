// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package phpfpm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	pfpm "github.com/hipages/php-fpm_exporter/phpfpm"
)

func fetchPoolData(statusURL string) (*pfpm.Pool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := statusURL
	if !strings.Contains(url, "?") {
		url += "?json&full"
	} else if !strings.Contains(url, "json") {
		url += "&json&full"
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Errorf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	content = pfpm.JSONResponseFixer(content)

	var pool pfpm.Pool
	if err := json.Unmarshal(content, &pool); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w, content: %s", err, string(content))
	}

	pool.Address = statusURL

	return &pool, nil
}

func (ipt *Input) collectPoolsPts() ([]*point.Point, error) {
	ipt.collectCache = make([]*point.Point, 0)
	opts := append(point.DefaultMetricOptions(), point.WithTime(ipt.ptsTime))

	if ipt.StatusURL == "" {
		l.Error("status_url is empty.")
		return nil, errors.New("status_url is empty")
	}
	var pools []*pfpm.Pool
	if ipt.UseFastCGI {
		pm := pfpm.PoolManager{}
		pm.Add(ipt.StatusURL)

		if err := pm.Update(); err != nil {
			l.Errorf("Could not update pool: %v", err)
		}

		for i := range pm.Pools {
			if pm.Pools[i].ScrapeError == nil && pm.Pools[i].Name != "" {
				pools = append(pools, &pm.Pools[i])
			} else {
				l.Warnf("Skipping pool %s due to scrape error: %w", pm.Pools[i].Address, pm.Pools[i].ScrapeError)
			}
		}

		if len(pools) == 0 {
			l.Warn("No valid FastCGI pools collected, returning empty points")
			return []*point.Point{}, nil
		}
	} else {
		// use http
		pool, err := fetchPoolData(ipt.StatusURL)
		if err != nil {
			l.Errorf("Could not fetch pool data from %s: %v", ipt.StatusURL, err)
			return nil, err
		}
		pools = []*pfpm.Pool{pool}
	}

	var kvs point.KVs
	for _, pool := range pools {
		// pool metric
		kvs = kvs.AddTag("pool", pool.Name)
		kvs = kvs.AddTag("address", pool.Address)
		kvs = kvs.AddTag("process_manager", pool.ProcessManager)

		kvs = kvs.Set("accepted_connections", pool.AcceptedConnections)
		kvs = kvs.Set("active_processes", pool.ActiveProcesses)
		kvs = kvs.Set("idle_processes", pool.IdleProcesses)
		kvs = kvs.Set("listen_queue", pool.ListenQueue)
		kvs = kvs.Set("listen_queue_length", pool.ListenQueueLength)
		kvs = kvs.Set("max_active_processes", pool.MaxActiveProcesses)
		kvs = kvs.Set("max_children_reached", pool.MaxChildrenReached)
		kvs = kvs.Set("max_listen_queue", pool.MaxListenQueue)
		kvs = kvs.Set("slow_requests", pool.SlowRequests)
		kvs = kvs.Set("total_processes", pool.TotalProcesses)

		for key, value := range ipt.Tags {
			kvs = kvs.AddTag(key, value)
		}

		// process metric
		for _, process := range pool.Processes {
			kvs = kvs.AddTag("pid", fmt.Sprintf("%d", process.PID))
			kvs = kvs.AddTag("process_state", process.State)

			kvs = kvs.Set("process_requests", process.Requests)
			kvs = kvs.Set("process_last_requestMemory", process.LastRequestMemory)
			kvs = kvs.Set("process_last_request_cpu", process.LastRequestCPU)
			kvs = kvs.Set("process_request_duration", process.RequestDuration)
		}
	}
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(inputName, kvs, opts...))

	return ipt.collectCache, nil
}
