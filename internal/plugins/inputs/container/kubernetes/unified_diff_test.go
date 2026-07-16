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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
)

func TestCompareAsUnifiedDiff(t *testing.T) {
	oldValue := map[string]string{
		"keep":  "same",
		"value": "same-1\nsame-2\nsame-3\nold",
	}
	newValue := map[string]string{
		"keep":  "same",
		"value": "same-1\nsame-2\nsame-3\nnew",
	}

	equal, diffText := compareAsUnifiedDiff("metadata.annotations", oldValue, newValue)
	require.False(t, equal)
	assert.Equal(t, `--- a/metadata.annotations
+++ b/metadata.annotations
@@ -1,2 +1,2 @@
 annotations:
-  value: same-1\nsame-2\nsame-3\nold
+  value: same-1\nsame-2\nsame-3\nnew
`, diffText)

	equal, diffText = compareAsUnifiedDiff("metadata.annotations", oldValue, oldValue)
	assert.True(t, equal)
	assert.Empty(t, diffText)
}

func TestCompareAsUnifiedDiffKeepsLongText(t *testing.T) {
	oldValue := "before-" + strings.Repeat("a", 600)
	newValue := "after-" + strings.Repeat("b", 600)

	diffs := compareAnnotations(
		changes.DeploymentAnnotations,
		&metav1.ObjectMeta{Annotations: map[string]string{"field.cattle.io/publicEndpoints": oldValue}},
		&metav1.ObjectMeta{Annotations: map[string]string{"field.cattle.io/publicEndpoints": newValue}},
	)
	require.Len(t, diffs, 1)
	diffText := diffs[0].DiffText
	assert.Contains(t, diffText, "-  field.cattle.io/publicEndpoints: "+oldValue+"\n")
	assert.Contains(t, diffText, "+  field.cattle.io/publicEndpoints: "+newValue+"\n")
}

func TestScalarAsUnifiedDiff(t *testing.T) {
	diffText := scalarAsUnifiedDiff("spec.replicas", "replicas", int32(1), int32(2))

	assert.Equal(t, `--- a/spec.replicas
+++ b/spec.replicas
@@ -1,2 +1,2 @@
 spec:
-  replicas: 1
+  replicas: 2
`, diffText)
}
