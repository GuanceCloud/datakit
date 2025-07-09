// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var filterAnnotationPatterns = []string{
	"control-plane.alpha.kubernetes.io/leader",
	"controller.kubernetes.io/pod-deletion-cost",
	"deployment.revision",
	"deployment.kubernetes.io/revision",
	"deployment.kubernetes.io/revision-history",
	"deprecated.daemonset.template.generation",
	"kubectl.kubernetes.io/last-applied-configuration",
	"kubernetes.io/change-cause",
	"kubernetes.io/updated-by",
	"statefulset.kubernetes.io/pod-name",
	"pv.kubernetes.io/bind-completed",
	"pv.kubernetes.io/bound-by-controller",
}

type FieldDiff struct {
	ChangeID             changes.ChangeID
	Namespace            string
	OwnerKind, OwnerName string
	ContainerName        string

	OldValue, NewValue string
	ChangeValueList    []string
	DiffText           string
}

func createNoChangedFieldDiffs(changeID changes.ChangeID, namespace, kind, name string) (res []FieldDiff) {
	return []FieldDiff{
		{
			ChangeID:  changeID,
			Namespace: namespace,
			OwnerKind: kind,
			OwnerName: name,
		},
	}
}

func comparePodTemplate(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	res = append(res, compareContainers(oldPod, newPod)...)
	res = append(res, compareServiceAccount(oldPod, newPod)...)
	res = append(res, compareVolumes(oldPod, newPod)...)
	res = append(res, compareNetworkPolicy(oldPod, newPod)...)
	res = append(res, compareTolerations(oldPod, newPod)...)
	res = append(res, compareNodeSelector(oldPod, newPod)...)
	res = append(res, compareAffinity(oldPod, newPod)...)
	return
}

func compareContainers(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	for _, oldContainer := range oldPod.Spec.Containers {
		for _, newContainer := range newPod.Spec.Containers {
			if oldContainer.Name != newContainer.Name {
				continue
			}
			{
				if oldContainer.Image != newContainer.Image {
					res = append(res, FieldDiff{
						ChangeID:      changes.PodTemplateImage,
						ContainerName: oldContainer.Name,
						OldValue:      oldContainer.Image,
						NewValue:      newContainer.Image,
						DiffText:      formatAsDiffLines("image", oldContainer.Image, newContainer.Image),
					})
				}
			}
			{
				if equal, difftext := diff.Compare(oldContainer.Env, newContainer.Env); !equal {
					oldEnvMap := envSliceToMap(oldContainer.Env)
					newEnvMap := envSliceToMap(newContainer.Env)
					res = append(res, FieldDiff{
						ChangeID:        changes.PodTemplateEnv,
						ContainerName:   oldContainer.Name,
						ChangeValueList: diffMaps(oldEnvMap, newEnvMap),
						DiffText:        difftext,
					})
				}
			}
			{
				if equal, difftext := diff.Compare(oldContainer.Command, newContainer.Command); !equal {
					oldStr := stringSliceToString(oldContainer.Command)
					newStr := stringSliceToString(newContainer.Command)
					res = append(res, FieldDiff{
						ChangeID:      changes.PodTemplateCommand,
						ContainerName: oldContainer.Name,
						OldValue:      oldStr,
						NewValue:      newStr,
						DiffText:      difftext,
					})
				}
			}
			{
				if equal, difftext := diff.Compare(oldContainer.Resources.Limits, newContainer.Resources.Limits); !equal {
					oldLimits := resourceToStringSlice(oldContainer.Resources.Limits)
					newLimits := resourceToStringSlice(newContainer.Resources.Limits)
					res = append(res, FieldDiff{
						ChangeID:      changes.PodTemplateResources,
						ContainerName: oldContainer.Name,
						OldValue:      strings.Join(oldLimits, ","),
						NewValue:      strings.Join(newLimits, ","),
						DiffText:      difftext,
					})
				}
			}
			{
				if equal, difftext := diff.Compare(oldContainer.VolumeMounts, newContainer.VolumeMounts); !equal {
					oldVolumeMountMap := volumeMountsToMap(oldContainer.VolumeMounts)
					newVolumeMountMap := volumeMountsToMap(newContainer.VolumeMounts)
					res = append(res, FieldDiff{
						ChangeID:        changes.PodTemplateVolumeMounts,
						ContainerName:   oldContainer.Name,
						ChangeValueList: diffMaps(oldVolumeMountMap, newVolumeMountMap),
						DiffText:        difftext,
					})
				}
			}
			{
				if equal, difftext := diff.Compare(oldContainer.SecurityContext, newContainer.SecurityContext); !equal {
					oldSecurityContextMap := securityContextToMap(oldContainer.SecurityContext)
					newSecurityContextMap := securityContextToMap(newContainer.SecurityContext)
					res = append(res, FieldDiff{
						ChangeID:        changes.PodTemplateSecurityContext,
						ContainerName:   oldContainer.Name,
						ChangeValueList: diffMaps(oldSecurityContextMap, newSecurityContextMap),
						DiffText:        difftext,
					})
				}
			}
			{
				changeValueList := []string{}
				difftextList := []string{}

				if equal, difftext := diff.Compare(oldContainer.LivenessProbe, newContainer.LivenessProbe); !equal {
					oldLivenessProbeMap := probeToMap(oldContainer.LivenessProbe)
					newLivenessProbeMap := probeToMap(newContainer.LivenessProbe)
					for _, key := range diffMaps(oldLivenessProbeMap, newLivenessProbeMap) {
						changeValueList = append(changeValueList, "- LivenessProbe."+key)
					}
					difftextList = append(difftextList, difftext)
				}
				if equal, difftext := diff.Compare(oldContainer.ReadinessProbe, newContainer.ReadinessProbe); !equal {
					oldReadinessProbeMap := probeToMap(oldContainer.ReadinessProbe)
					newReadinessProbeMap := probeToMap(newContainer.ReadinessProbe)
					for _, key := range diffMaps(oldReadinessProbeMap, newReadinessProbeMap) {
						changeValueList = append(changeValueList, "- ReadinessProbe."+key)
					}
					difftextList = append(difftextList, difftext)
				}

				if len(changeValueList) != 0 {
					res = append(res, FieldDiff{
						ChangeID:        changes.PodTemplateProbe,
						ContainerName:   oldContainer.Name,
						ChangeValueList: changeValueList,
						DiffText:        strings.Join(difftextList, "\n"),
					})
				}
			}
		}
	}
	return res
}

func compareTolerations(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	if equal, difftext := diff.Compare(oldPod.Spec.Tolerations, newPod.Spec.Tolerations); !equal {
		oldTolerationMap := tolerationsToMap(oldPod.Spec.Tolerations)
		newTolerationMap := tolerationsToMap(newPod.Spec.Tolerations)
		res = append(res, FieldDiff{
			ChangeID:        changes.PodTemplateTolerations,
			ChangeValueList: diffMaps(oldTolerationMap, newTolerationMap),
			DiffText:        difftext,
		})
	}
	return
}

func compareServiceAccount(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	if oldPod.Spec.ServiceAccountName != newPod.Spec.ServiceAccountName {
		res = append(res, FieldDiff{
			ChangeID: changes.PodTemplateServiceAccount,
			OldValue: oldPod.Spec.ServiceAccountName,
			NewValue: newPod.Spec.ServiceAccountName,
			DiffText: formatAsDiffLines("serviceAccountName", oldPod.Spec.ServiceAccountName, newPod.Spec.ServiceAccountName),
		})
	}
	return
}

func compareNodeSelector(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	if equal, difftext := diff.Compare(oldPod.Spec.NodeSelector, newPod.Spec.NodeSelector); !equal {
		res = append(res, FieldDiff{
			ChangeID:        changes.PodTemplateNodeSelector,
			ChangeValueList: diffMaps(oldPod.Spec.NodeSelector, newPod.Spec.NodeSelector),
			DiffText:        difftext,
		})
	}
	return
}

func compareVolumes(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	if equal, difftext := diff.Compare(oldPod.Spec.Volumes, newPod.Spec.Volumes); !equal {
		res = append(res, FieldDiff{
			ChangeID: changes.PodTemplateVolumes,
			DiffText: difftext,
		})
	}
	return
}

func compareAffinity(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	if oldPod.Spec.Affinity == nil && newPod.Spec.Affinity == nil {
		return
	}
	if equal, difftext := diff.Compare(oldPod.Spec.Affinity, newPod.Spec.Affinity); !equal {
		res = append(res, FieldDiff{
			ChangeID: changes.PodTemplateAffinity,
			DiffText: difftext,
		})
	}
	return
}

func compareNetworkPolicy(oldPod, newPod *apicorev1.PodTemplateSpec) (res []FieldDiff) {
	oldNetworkPolicyInfo := &networkPolicyInfo{
		HostNetwork:           oldPod.Spec.HostNetwork,
		DNSPolicy:             oldPod.Spec.DNSPolicy,
		ShareProcessNamespace: oldPod.Spec.ShareProcessNamespace,
		DNSConfig:             oldPod.Spec.DNSConfig,
		HostAliases:           oldPod.Spec.HostAliases,
	}
	newNetworkPolicyInfo := &networkPolicyInfo{
		HostNetwork:           newPod.Spec.HostNetwork,
		DNSPolicy:             newPod.Spec.DNSPolicy,
		ShareProcessNamespace: newPod.Spec.ShareProcessNamespace,
		DNSConfig:             newPod.Spec.DNSConfig,
		HostAliases:           newPod.Spec.HostAliases,
	}
	if equal, difftext := diff.Compare(oldNetworkPolicyInfo, newNetworkPolicyInfo); !equal {
		oldNetworkPolicyMap := networkPolicyInfoToMap(oldNetworkPolicyInfo)
		newNetworkPolicyMap := networkPolicyInfoToMap(newNetworkPolicyInfo)
		res = append(res, FieldDiff{
			ChangeID:        changes.PodTemplateNetworkPolilcy,
			ChangeValueList: diffMaps(oldNetworkPolicyMap, newNetworkPolicyMap),
			DiffText:        difftext,
		})
	}
	return
}

func compareLabels(changeID changes.ChangeID, oldObj, newObj *metav1.ObjectMeta) (res []FieldDiff) {
	if equal, difftext := diff.Compare(oldObj.Labels, newObj.Labels); !equal {
		res = append(res, FieldDiff{
			ChangeID:        changeID,
			ChangeValueList: diffMaps(oldObj.Labels, newObj.Labels),
			DiffText:        difftext,
		})
	}
	return
}

func compareAnnotations(changeID changes.ChangeID, oldObj, newObj *metav1.ObjectMeta) (res []FieldDiff) {
	oldAnnotations := deepCopyAnnotations(oldObj.Annotations)
	newAnnotations := deepCopyAnnotations(newObj.Annotations)
	filterMeaninglessAnnotations(oldAnnotations)
	filterMeaninglessAnnotations(newAnnotations)

	if equal, difftext := diff.Compare(oldAnnotations, newAnnotations); !equal {
		res = append(res, FieldDiff{
			ChangeID:        changeID,
			ChangeValueList: diffMaps(oldAnnotations, newAnnotations),
			DiffText:        difftext,
		})
	}
	return
}

func diffMaps(oldMap, newMap map[string]string) []string {
	var changeValueList []string

	for name, newVal := range newMap {
		if oldVal, exists := oldMap[name]; exists {
			if oldVal != newVal { // update
				changeValueList = append(changeValueList, formatChangeValue(name, oldVal, newVal))
			}
		} else { // add
			changeValueList = append(changeValueList, formatChangeValue(name, "", newVal))
		}
	}
	for name := range oldMap {
		if _, exists := newMap[name]; !exists { // delete
			changeValueList = append(changeValueList, formatChangeValue(name, "", ""))
		}
	}

	return changeValueList
}

func envSliceToMap(envs []apicorev1.EnvVar) map[string]string {
	res := make(map[string]string)
	for _, env := range envs {
		if env.ValueFrom != nil {
			res[env.Name] = fmt.Sprintf("Ref:%s", env.ValueFrom.String())
		} else {
			res[env.Name] = env.Value
		}
	}
	return res
}

func volumeMountsToMap(mounts []apicorev1.VolumeMount) map[string]string {
	res := make(map[string]string)
	for _, mount := range mounts {
		values := []string{
			fmt.Sprintf("ReadOnly=%v", mount.ReadOnly),
		}
		if mount.MountPath != "" {
			values = append(values, "MountPath="+mount.MountPath)
		}
		if mount.SubPath != "" {
			values = append(values, "SubPath="+mount.SubPath)
		}
		if mount.SubPathExpr != "" {
			values = append(values, "SubPathExpr="+mount.SubPathExpr)
		}
		if mount.MountPropagation != nil {
			values = append(values, "MountPropagation="+string(*mount.MountPropagation))
		}
		res[mount.Name] = strings.Join(values, ",")
	}
	return res
}

type networkPolicyInfo struct {
	HostNetwork           bool                    `json:"hostNetwork,omitempty"`
	DNSPolicy             apicorev1.DNSPolicy     `json:"dnsPolicy,omitempty"`
	ShareProcessNamespace *bool                   `json:"shareProcessNamespace,omitempty"`
	DNSConfig             *apicorev1.PodDNSConfig `json:"dnsConfig,omitempty"`
	HostAliases           []apicorev1.HostAlias   `json:"hostAliases,omitempty"`
}

func networkPolicyInfoToMap(spec *networkPolicyInfo) map[string]string {
	res := make(map[string]string)
	if spec == nil {
		return res
	}

	res["HostNetwork"] = strconv.FormatBool(spec.HostNetwork)
	res["DNSPolicy"] = string(spec.DNSPolicy)
	if spec.ShareProcessNamespace != nil {
		res["ShareProcessNamespace"] = strconv.FormatBool(*spec.ShareProcessNamespace)
	}
	if spec.DNSConfig != nil {
		if len(spec.DNSConfig.Nameservers) != 0 {
			res["DNSConfig.Nameservers"] = stringSliceToString(spec.DNSConfig.Nameservers)
		}
		if len(spec.DNSConfig.Searches) != 0 {
			res["DNSConfig.Searches"] = stringSliceToString(spec.DNSConfig.Searches)
		}
		if len(spec.DNSConfig.Options) != 0 {
			b, _ := json.Marshal(spec.DNSConfig.Options)
			if len(b) != 0 {
				res["DNSConfig.Options"] = string(b)
			}
		}
	}
	if len(spec.HostAliases) != 0 {
		b, _ := json.Marshal(spec.HostAliases)
		if len(b) != 0 {
			res["HostAliases"] = string(b)
		}
	}
	return res
}

func securityContextToMap(sc *apicorev1.SecurityContext) map[string]string {
	res := make(map[string]string)
	if sc == nil {
		return res
	}

	if sc.Privileged != nil {
		res["Privileged"] = strconv.FormatBool(*sc.Privileged)
	}
	if sc.RunAsUser != nil {
		res["RunAsUser"] = strconv.FormatInt(*sc.RunAsUser, 10)
	}
	if sc.RunAsGroup != nil {
		res["RunAsGroup"] = strconv.FormatInt(*sc.RunAsGroup, 10)
	}
	if sc.RunAsNonRoot != nil {
		res["RunAsNonRoot"] = strconv.FormatBool(*sc.RunAsNonRoot)
	}
	if sc.AllowPrivilegeEscalation != nil {
		res["AllowPrivilegeEscalation"] = strconv.FormatBool(*sc.AllowPrivilegeEscalation)
	}
	if sc.ReadOnlyRootFilesystem != nil {
		res["ReadOnlyRootFilesystem"] = strconv.FormatBool(*sc.ReadOnlyRootFilesystem)
	}
	if sc.ProcMount != nil {
		res["ProcMount"] = string(*sc.ProcMount)
	}

	if sc.Capabilities != nil {
		if len(sc.Capabilities.Add) > 0 {
			var add []string
			for _, value := range sc.Capabilities.Add {
				add = append(add, string(value))
			}
			res["Capabilities.Add"] = stringSliceToString(add)
		}
		if len(sc.Capabilities.Drop) > 0 {
			var drop []string
			for _, value := range sc.Capabilities.Drop {
				drop = append(drop, string(value))
			}
			res["Capabilities.Drop"] = stringSliceToString(drop)
		}
	}
	if sc.SELinuxOptions != nil {
		if sc.SELinuxOptions.User != "" {
			res["SELinuxOptions.User"] = sc.SELinuxOptions.User
		}
		if sc.SELinuxOptions.Role != "" {
			res["SELinuxOptions.Role"] = sc.SELinuxOptions.Role
		}
		if sc.SELinuxOptions.Type != "" {
			res["SELinuxOptions.Type"] = sc.SELinuxOptions.Type
		}
		if sc.SELinuxOptions.Level != "" {
			res["SELinuxOptions.Level"] = sc.SELinuxOptions.Level
		}
	}
	if sc.SeccompProfile != nil {
		res["SeccompProfile.Type"] = string(sc.SeccompProfile.Type)
		if sc.SeccompProfile.LocalhostProfile != nil {
			res["SeccompProfile.LocalhostProfile"] = *sc.SeccompProfile.LocalhostProfile
		}
	}
	if sc.WindowsOptions != nil {
		if sc.WindowsOptions.GMSACredentialSpecName != nil {
			res["WindowsOptions.GMSACredentialSpecName"] = *sc.WindowsOptions.GMSACredentialSpecName
		}
		if sc.WindowsOptions.GMSACredentialSpec != nil {
			res["WindowsOptions.GMSACredentialSpec"] = *sc.WindowsOptions.GMSACredentialSpec
		}
		if sc.WindowsOptions.RunAsUserName != nil {
			res["WindowsOptions.RunAsUserName"] = *sc.WindowsOptions.RunAsUserName
		}
		if sc.WindowsOptions.HostProcess != nil {
			res["WindowsOptions.HostProcess"] = strconv.FormatBool(*sc.WindowsOptions.HostProcess)
		}
	}

	return res
}

func tolerationsToMap(tolerations []apicorev1.Toleration) map[string]string {
	res := make(map[string]string)
	for _, t := range tolerations {
		if t.Key == "" {
			continue
		}
		values := []string{
			fmt.Sprintf("Operator=%s", t.Operator),
			fmt.Sprintf("Value=%s", t.Value),
			fmt.Sprintf("Effect=%s", t.Effect),
		}
		if t.TolerationSeconds != nil {
			values = append(values, fmt.Sprintf("TolerationSeconds=%d", *t.TolerationSeconds))
		}
		res[t.Key] = strings.Join(values, ",")
	}
	return res
}

func probeToMap(probe *apicorev1.Probe) map[string]string {
	res := make(map[string]string)
	if probe == nil {
		return res
	}

	res["InitialDelaySeconds"] = strconv.Itoa(int(probe.InitialDelaySeconds))
	res["TimeoutSeconds"] = strconv.Itoa(int(probe.TimeoutSeconds))
	res["PeriodSeconds"] = strconv.Itoa(int(probe.PeriodSeconds))
	res["SuccessThreshold"] = strconv.Itoa(int(probe.SuccessThreshold))
	res["FailureThreshold"] = strconv.Itoa(int(probe.FailureThreshold))
	if probe.TerminationGracePeriodSeconds != nil {
		res["TerminationGracePeriodSeconds"] = strconv.Itoa(int(*probe.TerminationGracePeriodSeconds))
	}

	if probe.Exec != nil {
		res["Exec.Command"] = stringSliceToString(probe.Exec.Command)
	}
	if probe.HTTPGet != nil {
		res["Path"] = probe.HTTPGet.Path
		res["Port"] = probe.HTTPGet.Port.String()
		res["Host"] = probe.HTTPGet.Host
		res["Scheme"] = string(probe.HTTPGet.Scheme)

		b, _ := json.Marshal(probe.HTTPGet.HTTPHeaders)
		if len(b) != 0 {
			res["HTTPHeaders"] = string(b)
		}
	}
	if probe.TCPSocket != nil {
		res["TCPSocket.Port"] = probe.TCPSocket.Port.String()
		res["TCPSocket.Host"] = probe.TCPSocket.Host
	}
	if probe.GRPC != nil {
		res["GRPC.Port"] = strconv.Itoa(int(probe.GRPC.Port))
		if probe.GRPC.Service != nil {
			res["GRPC.Service"] = *probe.GRPC.Service
		}
	}

	return res
}

func resourceToStringSlice(resources apicorev1.ResourceList) []string {
	res := make([]string, 0, len(resources))

	if value, exists := resources[apicorev1.ResourceCPU]; exists {
		res = append(res, fmt.Sprintf("%s=%s", apicorev1.ResourceCPU, value.String()))
	}
	if value, exists := resources[apicorev1.ResourceMemory]; exists {
		res = append(res, fmt.Sprintf("%s=%s", apicorev1.ResourceMemory, value.String()))
	}
	if value, exists := resources[apicorev1.ResourceStorage]; exists {
		res = append(res, fmt.Sprintf("%s=%s", apicorev1.ResourceStorage, value.String()))
	}
	if value, exists := resources[apicorev1.ResourceEphemeralStorage]; exists {
		res = append(res, fmt.Sprintf("%s=%s", apicorev1.ResourceEphemeralStorage, value.String()))
	}

	return res
}

func formatAsDiffLines(key, oldVal, newVal string) string {
	return fmt.Sprintf("- %s: %s\n+ %s: %s", key, oldVal, key, newVal)
}

func formatChangeValue(key, oldVal, newVal string) string {
	if oldVal == "" && newVal == "" {
		return fmt.Sprintf("- Delete: %s", key)
	}
	if oldVal == "" && newVal != "" {
		return fmt.Sprintf("- Add: %s = %s", key, newVal)
	}
	return fmt.Sprintf("- %s: %s -> %s", key, oldVal, newVal)
}

func deepCopyAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}
	m := make(map[string]string, len(annotations))
	for k, v := range annotations {
		m[k] = v
	}
	return m
}

func filterMeaninglessAnnotations(annotations map[string]string) {
	for key := range annotations {
		for _, pattern := range filterAnnotationPatterns {
			if strings.Contains(key, pattern) {
				delete(annotations, key)
				break
			}
		}
	}
}

func stringSliceToString(slice []string) string {
	quoted := make([]string, len(slice))
	for i, s := range slice {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func fillOwnerInfoForDiffs(diffs []FieldDiff, namespace, kind, name string) {
	for idx := range diffs {
		diffs[idx].Namespace = namespace
		diffs[idx].OwnerKind = kind
		diffs[idx].OwnerName = name
	}
}
