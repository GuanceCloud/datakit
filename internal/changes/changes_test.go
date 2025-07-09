// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package changes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadManifest(t *testing.T) {
	_, err := loadManifest("manifests/k8s.toml")
	assert.NoError(t, err)
}

func TestRenderK8sTemplate(t *testing.T) {
	globalManifests = Manifests{
		K8sManifest: &Manifest{
			Version: "1.0",
			Changes: []Change{
				{
					ID:    "k8s_change_01_01",
					Title: I18n{En: "{{.OwnerKind}} Changed"},
					Message: I18n{En: `The image for {{.OwnerKind}} {{.OwnerName}} in Kubernetes namespace {{.Namespace}} have changed.
Container name: {{.ContainerName}}
Old Value: {{.OldValue}}
New Value: {{.NewValue}}`},
				},
			},
		},
	}

	data := struct {
		Namespace            string
		OwnerKind, OwnerName string
		ContainerName        string
		OldValue, NewValue   string
		ChangeValueList      []string
	}{
		Namespace:     "kube-system",
		OwnerKind:     "Deployment",
		OwnerName:     "etcd-abc",
		ContainerName: "etcd-container",
		OldValue:      "etcd:v1.1.1",
		NewValue:      "etcd:v1.2.2",
	}

	title, message, err := RenderK8sTemplate(LangEn, PodTemplateImage, data)
	assert.NoError(t, err)

	assert.Equal(t, "Deployment Changed", title)
	t.Logf(message)
}
