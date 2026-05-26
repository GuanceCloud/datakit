// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func testLoggingDefaults() *loggingDefaults {
	return &loggingDefaults{
		setLabelAsTags: func(map[string]string) map[string]string { return nil },
		insideFilepathFunc: func(hostDir, insideDir, path string) string {
			return ""
		},
	}
}

func testContainerLogInfo() *containerLogInfo {
	return &containerLogInfo{
		containerID:   "cid-test",
		containerName: "test-container",
		runtime:       "docker",
		image:         "repo/test:latest",
		logPath:       "/tmp/test.log",
		podNamespace:  "default",
		podName:       "pod-a",
	}
}

func taskCount(c *containerLogCoordinator) int {
	c.taskMutex.RLock()
	defer c.taskMutex.RUnlock()
	return len(c.containerTasks)
}

func TestAddTaskRollbackWhenFiltered(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())

	coordinator.addTask("cid-test", testContainerLogInfo(), "", true)

	require.Equal(t, 0, taskCount(coordinator))
}

func TestAddTaskRollbackOnParseError(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())

	coordinator.addTask("cid-test", testContainerLogInfo(), "[", false)

	require.Equal(t, 0, taskCount(coordinator))
}

func TestAddTaskRollbackWhenAllDisabled(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())

	coordinator.addTask("cid-test", testContainerLogInfo(), `[{"disable":true}]`, false)

	require.Equal(t, 0, taskCount(coordinator))
}

func TestRemoveTaskDeletesFromMap(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	coordinator.containerTasks["cid-test"] = &containerLogTask{
		containerID: "cid-test",
	}

	coordinator.removeTask("cid-test")

	require.Equal(t, 0, taskCount(coordinator))
}

func TestAddTaskConcurrentFilteredNoLeak(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	info := testContainerLogInfo()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			coordinator.addTask(fmt.Sprintf("cid-%d", i), info, "", true)
		}(i)
	}
	wg.Wait()

	require.Equal(t, 0, taskCount(coordinator))
}
