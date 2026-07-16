// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"

	apicorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/unifieddiff"
)

func compareAsUnifiedDiff(path string, oldVal, newVal interface{}) (equal bool, diffText string) {
	return compareAsUnifiedDiffWithRootKey(path, lastPathSegment(path), oldVal, newVal)
}

func compareAsUnifiedDiffWithRootKey(
	path, rootKey string, oldVal, newVal interface{},
) (equal bool, diffText string) {
	if reflect.DeepEqual(oldVal, newVal) {
		return true, ""
	}
	if oldMap, ok := oldVal.(map[string]string); ok {
		if newMap, ok := newVal.(map[string]string); ok {
			oldVal, newVal = changedStringMapEntries(oldMap, newMap)
		}
	}

	oldText, err := yaml.Marshal(map[string]interface{}{rootKey: oldVal})
	if err != nil {
		klog.Warnf("cannot marshal old value for Kubernetes change diff %q: %s", path, err)
		return false, ""
	}

	newText, err := yaml.Marshal(map[string]interface{}{rootKey: newVal})
	if err != nil {
		klog.Warnf("cannot marshal new value for Kubernetes change diff %q: %s", path, err)
		return false, ""
	}
	if bytes.Equal(oldText, newText) {
		return true, ""
	}

	return false, unifieddiff.Text(path, string(oldText), string(newText))
}

func scalarAsUnifiedDiff(path, key string, oldVal, newVal interface{}) string {
	contextPath := path
	if index := strings.LastIndexByte(path, '.'); index >= 0 {
		contextPath = path[:index]
	}

	_, diffText := compareAsUnifiedDiffWithRootKey(
		path,
		lastPathSegment(contextPath),
		map[string]interface{}{key: oldVal},
		map[string]interface{}{key: newVal},
	)
	return diffText
}

func lastPathSegment(path string) string {
	return path[strings.LastIndexByte(path, '.')+1:]
}

func changedStringMapEntries(oldMap, newMap map[string]string) (map[string]string, map[string]string) {
	oldChanged := make(map[string]string)
	newChanged := make(map[string]string)

	for key, oldValue := range oldMap {
		newValue, exists := newMap[key]
		if !exists || newValue != oldValue {
			oldChanged[key] = singleLineDiffValue(oldValue)
		}
	}
	for key, newValue := range newMap {
		oldValue, exists := oldMap[key]
		if !exists || oldValue != newValue {
			newChanged[key] = singleLineDiffValue(newValue)
		}
	}

	return oldChanged, newChanged
}

func singleLineDiffValue(value string) string {
	quoted := strconv.Quote(value)
	return quoted[1 : len(quoted)-1]
}

func changedEnvVars(oldEnvs, newEnvs []apicorev1.EnvVar) ([]apicorev1.EnvVar, []apicorev1.EnvVar) {
	sharedLength := len(oldEnvs)
	if len(newEnvs) < sharedLength {
		sharedLength = len(newEnvs)
	}
	var oldChanged, newChanged []apicorev1.EnvVar

	for index := 0; index < sharedLength; index++ {
		if !reflect.DeepEqual(oldEnvs[index], newEnvs[index]) {
			oldChanged = append(oldChanged, oldEnvs[index])
			newChanged = append(newChanged, newEnvs[index])
		}
	}
	oldChanged = append(oldChanged, oldEnvs[sharedLength:]...)
	newChanged = append(newChanged, newEnvs[sharedLength:]...)

	return oldChanged, newChanged
}

func envDiffValues(envs []apicorev1.EnvVar) []string {
	values := make([]string, 0, len(envs))
	for _, env := range envs {
		values = append(values, envDiffValue(env))
	}
	return values
}

func envDiffValue(env apicorev1.EnvVar) string {
	if env.ValueFrom != nil {
		return env.Name + "=valueFrom(" + strconv.Quote(env.ValueFrom.String()) + ")"
	}
	return env.Name + "=" + strconv.Quote(env.Value)
}

func containerDiffPath(containerName, field string) string {
	return "spec.template.spec.containers." + containerName + "." + field
}
