// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"

	kubev1guancebeta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type datakitCRD struct {
	client k8sClientX
	items  []kubev1guancebeta1.DataKit
}

func newDatakitCRD(client k8sClientX) *datakitCRD {
	return &datakitCRD{
		client: client,
	}
}

func (d *datakitCRD) pullItems() error {
	if len(d.items) != 0 {
		return nil
	}

	list, err := d.client.getDataKits().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get datakits resource: %w", err)
	}

	d.items = list.Items
	return nil
}

func (d *datakitCRD) forkInputs() error {
	if err := d.pullItems(); err != nil {
		return err
	}

	l.Debugf("fetch datakit resources, %#v", d.items)

	for _, item := range d.items {
		opt := metav1.ListOptions{LabelSelector: "app=" + item.Spec.K8sDeployment}
		pods, err := d.client.getPodsForNamespace(item.Spec.K8sNamespace).List(context.Background(), opt)
		if err != nil {
			l.Warn(err)
			continue
		}

		l.Debugf("crd pods, len %d", len(pods.Items))

		for idx, pod := range pods.Items {
			if _, ok := discoveryInputsMap[string(pod.UID)]; ok {
				continue
			}

			l.Debugf("fork crd input: %s, %s", pod.Name, item.Spec.InputConf)

			instance := discoveryInput{
				id:     string(pod.UID),
				source: pod.Name,
				name:   "datakitCRD",
				config: complatePromConfig(item.Spec.InputConf, &pods.Items[idx]),
				extraTags: map[string]string{
					"pod_name":  pod.Name,
					"node_name": pod.Spec.NodeName,
					"namespace": defaultNamespace(pod.Namespace),
				},
			}
			if err := instance.run(); err != nil {
				l.Warn(err)
			}
		}
	}
	return nil
}
