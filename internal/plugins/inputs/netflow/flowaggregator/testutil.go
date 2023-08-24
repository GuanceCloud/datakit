// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package flowaggregator

import (
	"fmt"
	"time"
)

func WaitForFlowsToBeFlushed(aggregator *FlowAggregator, timeoutDuration time.Duration, minEvents uint64) (uint64, error) {
	timeout := time.After(timeoutDuration)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return 0, fmt.Errorf("timeout error waiting for events")
		// Got a tick, we should check on doSomething()
		case <-ticker.C:
			events := aggregator.flushedFlowCount.Load()
			if events >= minEvents {
				return events, nil
			}
		}
	}
}
