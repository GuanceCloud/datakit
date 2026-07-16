// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiappsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type changeTestFeeder struct {
	points []*point.Point
}

func (f *changeTestFeeder) Feed(_ point.Category, points []*point.Point, _ ...dkio.FeedOption) error {
	f.points = append(f.points, points...)
	return nil
}

func (*changeTestFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestProcessChangeAddsUnifiedDiff(t *testing.T) {
	require.NoError(t, changes.LoadK8sManifest())

	feeder := &changeTestFeeder{}
	cfg := &Config{Feeder: feeder}
	obj := &apiappsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Name:      "app",
		Namespace: "default",
		UID:       types.UID("deployment-uid"),
	}}
	diffText := scalarAsUnifiedDiff(
		"spec.template.spec.containers.app.image", "image", "example/app:v1", "example/app:v2",
	)
	diffs := []FieldDiff{
		{
			ChangeID:      changes.PodTemplateImage,
			OwnerKind:     deploymentType,
			OwnerName:     obj.Name,
			ContainerName: "app",
			OldValue:      "example/app:v1",
			NewValue:      "example/app:v2",
			DiffText:      diffText,
		},
	}

	processChange(cfg, deploymentObjectClass, deploymentObjectResourceKey, diffs, obj)
	require.Len(t, feeder.points, 1)
	assert.Equal(t, diffText, feeder.points[0].Get("diff"))
}

func TestProcessChangeTruncatesLongChangeListValues(t *testing.T) {
	require.NoError(t, changes.LoadK8sManifest())

	feeder := &changeTestFeeder{}
	cfg := &Config{Feeder: feeder}
	obj := &apiappsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Name:      "app",
		Namespace: "default",
		UID:       types.UID("deployment-uid"),
	}}
	oldValue := strings.Repeat("旧", 200)
	newValue := "4"
	diffs := compareAnnotations(
		changes.DeploymentAnnotations,
		&metav1.ObjectMeta{Annotations: map[string]string{"example.com/long": oldValue}},
		&metav1.ObjectMeta{Annotations: map[string]string{"example.com/long": newValue}},
	)
	require.Len(t, diffs, 1)
	fillOwnerInfoForDiffs(diffs, obj.Namespace, deploymentType, obj.Name)
	diffText := diffs[0].DiffText

	processChange(cfg, deploymentObjectClass, deploymentObjectResourceKey, diffs, obj)

	require.Len(t, feeder.points, 1)
	expectedValue := strings.Repeat("旧", 113) + "... (truncated)"
	assert.Equal(t, "Annotations for Deployment app have changed\nChange details:\n[- example.com/long: "+expectedValue+" -> 4]", feeder.points[0].Get("df_message"))
	assert.Equal(t, diffText, feeder.points[0].Get("diff"))
	assert.Contains(t, diffText, oldValue)
}
