// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1beta1 wraps DataKit resource by kubernetes client-gen.

//nolint
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *DataKit) DeepCopyInto(out *DataKit) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy copying the receiver, creating a new DataKit.
func (in *DataKit) DeepCopy() *DataKit {
	if in == nil {
		return nil
	}
	out := new(DataKit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject copying the receiver, creating a new runtime.Object.
func (in *DataKit) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *DataKitList) DeepCopyInto(out *DataKitList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DataKit, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy copying the receiver, creating a new DataKitList.
func (in *DataKitList) DeepCopy() *DataKitList {
	if in == nil {
		return nil
	}
	out := new(DataKitList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject copying the receiver, creating a new runtime.Object.
func (in *DataKitList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *DataKitSpec) DeepCopyInto(out *DataKitSpec) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	out.InputConf = in.InputConf
	out.K8sDeployment = in.K8sDeployment
	out.K8sNamespace = in.K8sNamespace
	out.Tags = in.Tags
	return
}

// DeepCopy copying the receiver, creating a new DataKitSpec.
func (in *DataKitSpec) DeepCopy() *DataKitSpec {
	if in == nil {
		return nil
	}
	out := new(DataKitSpec)
	in.DeepCopyInto(out)
	return out
}
