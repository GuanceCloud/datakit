// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/labels"
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

func TestAddTaskProcessesChangedAnnotationConfig(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	info := testContainerLogInfo()
	task := &containerLogTask{
		containerID:                  "cid-test",
		podUID:                       info.podUID,
		info:                         info,
		useAnnotationOrEnvLogConfigs: true,
		configStr:                    `[{"type":"file","path":"/old.log"}]`,
	}
	coordinator.containerTasks["cid-test"] = task

	newConfig := `[{"disable":true}]`
	coordinator.addTask("cid-test", info, newConfig, false)

	require.Equal(t, 1, taskCount(coordinator))
	require.Equal(t, newConfig, task.configStr)
	require.True(t, task.useAnnotationOrEnvLogConfigs)
}

func TestAddTaskKeepsExistingConfigOnChangedAnnotationParseError(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	info := testContainerLogInfo()
	oldConfig := `[{"type":"file","path":"/old.log"}]`
	task := &containerLogTask{
		containerID:                  "cid-test",
		podUID:                       info.podUID,
		info:                         info,
		useAnnotationOrEnvLogConfigs: true,
		configStr:                    oldConfig,
	}
	coordinator.containerTasks["cid-test"] = task

	coordinator.addTask("cid-test", info, "[", false)

	require.Equal(t, 1, taskCount(coordinator))
	require.Equal(t, oldConfig, task.configStr)
	require.True(t, task.useAnnotationOrEnvLogConfigs)
}

func TestAddTaskRefreshesDefaultConfigAfterPodMetadataBecomesAvailable(t *testing.T) {
	defaults := testLoggingDefaults()
	defaults.setLabelAsTags = func(labels map[string]string) map[string]string { return labels }
	coordinator := newContainerLogCoordinator(defaults)

	initialInfo := testContainerLogInfo()
	initialInfo.logPath = filepath.Join(t.TempDir(), "container.log")
	coordinator.addTask(initialInfo.containerID, initialInfo, "", false)
	t.Cleanup(func() { coordinator.removeTask(initialInfo.containerID) })
	task := coordinator.containerTasks[initialInfo.containerID]
	require.Len(t, task.tailers, 1)
	initialHash := task.tailers[0].configHash

	enrichedInfo := *initialInfo
	enrichedInfo.podLabels = map[string]string{"app": "example"}

	coordinator.addTask(initialInfo.containerID, &enrichedInfo, "", false)

	require.Same(t, &enrichedInfo, task.info)
	require.NotEqual(t, initialHash, task.tailers[0].configHash)
	require.Equal(t, task.configs[0].getStructHash(), task.tailers[0].configHash)
}

func TestCRDSelectorDistinguishesUnknownAndEmptyLabels(t *testing.T) {
	selector, err := labels.Parse("!app")
	require.NoError(t, err)
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	task := &containerLogTask{containerID: "cid-test", info: testContainerLogInfo()}
	crd := &crdLoggingConfig{podLabelSelector: "!app", podLabelSelectorMatch: selector}

	require.False(t, coordinator.matchesCRDConfig(task, crd))
	task.info.podLabels = map[string]string{}
	require.True(t, coordinator.matchesCRDConfig(task, crd))
}

func TestAddTaskDoesNotStoreConfigWhenTailerCreationFails(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	info := testContainerLogInfo()
	info.logPath = "["

	coordinator.addTask(info.containerID, info, "", false)
	coordinator.addTask(info.containerID, info, "", false)

	task := coordinator.containerTasks[info.containerID]
	require.NotNil(t, task)
	require.Empty(t, task.tailers)
	require.Empty(t, task.configs)
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

func TestRequestLoggingScanBroadcastAndMerge(t *testing.T) {
	coordinator := newContainerLogCoordinator(testLoggingDefaults())
	first := coordinator.registerLoggingScanSignal()
	second := coordinator.registerLoggingScanSignal()

	coordinator.requestLoggingScan()
	coordinator.requestLoggingScan()

	require.Len(t, first, 1)
	require.Len(t, second, 1)
}
