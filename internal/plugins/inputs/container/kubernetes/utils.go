// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var annotationFilterPrefixes = []string{
	"argocd.argoproj.io/",
	"cattle.io/",
	"field.cattle.io/",
	"fluxcd.io/",
	"rancher.io/",
	"kubectl.kubernetes.io/",
	"nginx.ingress.kubernetes.io/",
}

// getCleanYAML converts a Kubernetes object to a clean YAML string.
// It removes unnecessary fields and filters annotations based on predefined prefixes.
// Returns an empty string if conversion fails.
func getCleanYAML(obj interface{}) string {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return ""
	}

	u := &unstructured.Unstructured{Object: unstructuredObj}
	return convertUnstructuredToYAML(u)
}

func convertUnstructuredToYAML(obj *unstructured.Unstructured) string {
	cleaned := obj.DeepCopy()

	// Remove status field
	unstructured.RemoveNestedField(cleaned.Object, "status")

	// Remove unnecessary metadata fields
	metadataFieldsToRemove := []string{
		"creationTimestamp",
		"resourceVersion",
		"uid",
		"generation",
		"managedFields",
	}
	for _, field := range metadataFieldsToRemove {
		unstructured.RemoveNestedField(cleaned.Object, "metadata", field)
	}

	annotations, found, err := unstructured.NestedStringMap(obj.Object, "metadata", "annotations")
	if found && err == nil {
		filteredAnnotations := filterAnnotations(annotations)
		_ = unstructured.SetNestedStringMap(cleaned.Object, filteredAnnotations, "metadata", "annotations")
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(cleaned.Object)
	if err != nil {
		return ""
	}

	return string(yamlBytes)
}

func filterAnnotations(annotations map[string]string) map[string]string {
	if len(annotations) == 0 {
		return annotations
	}
	filtered := make(map[string]string, len(annotations))
	for k, v := range annotations {
		if !shouldFilterAnnotation(k) {
			filtered[k] = v
		}
	}
	return filtered
}

func shouldFilterAnnotation(key string) bool {
	for _, prefix := range annotationFilterPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}
