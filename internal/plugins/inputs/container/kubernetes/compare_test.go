// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiappsv1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
)

func TestCompareDeploymentUsesUnifiedDiff(t *testing.T) {
	oldReplicas, newReplicas := int32(1), int32(2)
	oldDeployment := &apiappsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "app",
			Namespace:   "default",
			Labels:      map[string]string{"version": "v1"},
			Annotations: map[string]string{"example.com/config": "old"},
		},
		Spec: apiappsv1.DeploymentSpec{
			Replicas: &oldReplicas,
			Strategy: apiappsv1.DeploymentStrategy{
				Type: apiappsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &apiappsv1.RollingUpdateDeployment{
					MaxUnavailable: intOrStringPointer(intstr.FromString("25%")),
					MaxSurge:       intOrStringPointer(intstr.FromString("25%")),
				},
			},
			Template: apicorev1.PodTemplateSpec{
				Spec: apicorev1.PodSpec{
					ServiceAccountName: "old-account",
					Containers: []apicorev1.Container{
						{
							Name:  "app",
							Image: "example/app:v1",
							Env: []apicorev1.EnvVar{
								{Name: "MODE", Value: "old"},
							},
						},
					},
				},
			},
		},
	}

	newDeployment := oldDeployment.DeepCopy()
	newDeployment.Spec.Replicas = &newReplicas
	newDeployment.Labels["version"] = "v2"
	newDeployment.Annotations["example.com/config"] = "new"
	newDeployment.Spec.Template.Spec.ServiceAccountName = "new-account"
	newDeployment.Spec.Template.Spec.Containers[0].Image = "example/app:v2"
	newDeployment.Spec.Template.Spec.Containers[0].Env[0].Value = "new"
	newDeployment.Spec.Strategy.RollingUpdate.MaxUnavailable = intOrStringPointer(intstr.FromInt(0))
	newDeployment.Spec.Strategy.RollingUpdate.MaxSurge = intOrStringPointer(intstr.FromString("50%"))

	diffs := compareDeployment(oldDeployment, newDeployment)
	require.NotEmpty(t, diffs)
	for _, fieldDiff := range diffs {
		assert.Truef(t, strings.HasPrefix(fieldDiff.DiffText, "--- a/"), "change %s is not unified", fieldDiff.ChangeID)
		assert.Contains(t, fieldDiff.DiffText, "\n+++ b/")
		assert.Contains(t, fieldDiff.DiffText, "\n@@ -")
	}

	strategyDiff := findFieldDiff(t, diffs, changes.DeploymentStrategy)
	assert.Equal(t, "Type=RollingUpdate,{MaxUnavailable=25%,MaxSurge=25%}", strategyDiff.OldValue)
	assert.Equal(t, "Type=RollingUpdate,{MaxUnavailable=0,MaxSurge=50%}", strategyDiff.NewValue)
}

func TestCompareAnnotationsTruncatesOnlyLongValues(t *testing.T) {
	tests := []struct {
		name     string
		oldValue string
		newValue string
		want     string
	}{
		{
			name:     "short values",
			oldValue: "3",
			newValue: "4",
			want:     "- example.com/value: 3 -> 4",
		},
		{
			name:     "value at limit",
			oldValue: strings.Repeat("旧", 128),
			newValue: "4",
			want:     "- example.com/value: " + strings.Repeat("旧", 128) + " -> 4",
		},
		{
			name:     "value over limit",
			oldValue: "4",
			newValue: strings.Repeat("新", 129),
			want:     "- example.com/value: 4 -> " + strings.Repeat("新", 113) + "... (truncated)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diffs := compareAnnotations(
				changes.DeploymentAnnotations,
				&metav1.ObjectMeta{Annotations: map[string]string{"example.com/value": test.oldValue}},
				&metav1.ObjectMeta{Annotations: map[string]string{"example.com/value": test.newValue}},
			)

			require.Len(t, diffs, 1)
			assert.Equal(t, []string{test.want}, diffs[0].ChangeValueList)
		})
	}
}

func TestCompareContainersAddsEnvContext(t *testing.T) {
	oldPod := &apicorev1.PodTemplateSpec{
		Spec: apicorev1.PodSpec{Containers: []apicorev1.Container{
			{
				Name: "app",
				Env: []apicorev1.EnvVar{
					{Name: "UNCHANGED_1", Value: "same"},
					{Name: "UNCHANGED_2", Value: "same"},
					{Name: "UNCHANGED_3", Value: "same"},
					{Name: "UNCHANGED_4", Value: "same"},
					{Name: "MODE", Value: "old"},
				},
			},
		}},
	}
	newPod := oldPod.DeepCopy()
	newPod.Spec.Containers[0].Env[4].Value = "new"

	envDiff := findFieldDiff(t, compareContainers(oldPod, newPod), changes.PodTemplateEnv)
	assert.Equal(t, `--- a/spec.template.spec.containers.app.env
+++ b/spec.template.spec.containers.app.env
@@ -1,2 +1,2 @@
 env:
-- MODE="old"
+- MODE="new"
`, envDiff.DiffText)
}

func TestCompareContainersDistinguishesLiteralAndValueFromEnv(t *testing.T) {
	source := &apicorev1.EnvVarSource{
		FieldRef: &apicorev1.ObjectFieldSelector{FieldPath: "metadata.name"},
	}
	oldPod := &apicorev1.PodTemplateSpec{
		Spec: apicorev1.PodSpec{Containers: []apicorev1.Container{
			{
				Name: "app",
				Env: []apicorev1.EnvVar{
					{Name: "POD_NAME", Value: "Ref:" + source.String()},
				},
			},
		}},
	}
	newPod := oldPod.DeepCopy()
	newPod.Spec.Containers[0].Env[0].Value = ""
	newPod.Spec.Containers[0].Env[0].ValueFrom = source

	envDiff := findFieldDiff(t, compareContainers(oldPod, newPod), changes.PodTemplateEnv)
	assert.Contains(t, envDiff.DiffText, "valueFrom(")
}

func TestCompareContainersCombinesProbeDiff(t *testing.T) {
	oldPod := &apicorev1.PodTemplateSpec{
		Spec: apicorev1.PodSpec{Containers: []apicorev1.Container{
			{
				Name: "app",
				LivenessProbe: &apicorev1.Probe{
					InitialDelaySeconds: 1,
				},
				ReadinessProbe: &apicorev1.Probe{
					TimeoutSeconds: 1,
				},
			},
		}},
	}
	newPod := oldPod.DeepCopy()
	newPod.Spec.Containers[0].LivenessProbe.InitialDelaySeconds = 2
	newPod.Spec.Containers[0].ReadinessProbe.TimeoutSeconds = 2

	probeDiff := findFieldDiff(t, compareContainers(oldPod, newPod), changes.PodTemplateProbe)
	assert.Equal(t, 1, strings.Count(probeDiff.DiffText, "--- "))
	assert.Contains(t, probeDiff.DiffText, "a/spec.template.spec.containers.app.probes")
	assert.Contains(t, probeDiff.DiffText, "\n probes:\n")
	assert.Contains(t, probeDiff.DiffText, "livenessProbe:")
	assert.Contains(t, probeDiff.DiffText, "readinessProbe:")
}

func findFieldDiff(t *testing.T, diffs []FieldDiff, changeID changes.ChangeID) FieldDiff {
	t.Helper()
	for _, fieldDiff := range diffs {
		if fieldDiff.ChangeID == changeID {
			return fieldDiff
		}
	}
	require.FailNowf(t, "missing field diff", "change ID %s", changeID)
	return FieldDiff{}
}

func intOrStringPointer(value intstr.IntOrString) *intstr.IntOrString {
	return &value
}
