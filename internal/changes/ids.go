// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package changes

type ChangeID string

var (
	//
	// DO NOT MODIFY - Strongly linked to change manifest.
	//
	PodTemplateImage           ChangeID = "k8s_change_01_01" // Defined in the Container section.
	PodTemplateEnv             ChangeID = "k8s_change_01_02" // Defined in the Container section.
	PodTemplateCommand         ChangeID = "k8s_change_01_03" // Defined in the Container section.
	PodTemplateResources       ChangeID = "k8s_change_01_04" // Defined in the Container section.
	PodTemplateVolumeMounts    ChangeID = "k8s_change_01_05" // Defined in the Container section.
	PodTemplateVolumes         ChangeID = "k8s_change_01_06"
	PodTemplateSecurityContext ChangeID = "k8s_change_01_07"
	PodTemplateProbe           ChangeID = "k8s_change_01_08"
	PodTemplateNetworkPolilcy  ChangeID = "k8s_change_01_09"
	PodTemplateTolerations     ChangeID = "k8s_change_01_10"
	PodTemplateNodeSelector    ChangeID = "k8s_change_01_11"
	PodTemplateAffinity        ChangeID = "k8s_change_01_12"
	PodTemplateServiceAccount  ChangeID = "k8s_change_01_13"

	DeploymentCreate      ChangeID = "k8s_change_02_01"
	DeploymentDelete      ChangeID = "k8s_change_02_02"
	DeploymentLabels      ChangeID = "k8s_change_02_03"
	DeploymentAnnotations ChangeID = "k8s_change_02_04"
	DeploymentReplicas    ChangeID = "k8s_change_02_05"
	DeploymentStrategy    ChangeID = "k8s_change_02_06"

	DaemonSetCreate      ChangeID = "k8s_change_03_01"
	DaemonSetDelete      ChangeID = "k8s_change_03_02"
	DaemonSetLabels      ChangeID = "k8s_change_03_03"
	DaemonSetAnnotations ChangeID = "k8s_change_03_04"

	StatefulSetCreate      ChangeID = "k8s_change_04_01"
	StatefulSetDelete      ChangeID = "k8s_change_04_02"
	StatefulSetLabels      ChangeID = "k8s_change_04_03"
	StatefulSetAnnotations ChangeID = "k8s_change_04_04"
	StatefulSetReplicas    ChangeID = "k8s_change_04_05"
)
