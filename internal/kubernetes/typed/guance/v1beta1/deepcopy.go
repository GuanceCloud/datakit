// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1beta1 wraps Datakit resource by kubernetes client-gen.

//nolint
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *Datakit) DeepCopyInto(out *Datakit) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy copying the receiver, creating a new Datakit.
func (in *Datakit) DeepCopy() *Datakit {
	if in == nil {
		return nil
	}
	out := new(Datakit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject copying the receiver, creating a new runtime.Object.
func (in *Datakit) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *DatakitList) DeepCopyInto(out *DatakitList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Datakit, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy copying the receiver, creating a new DatakitList.
func (in *DatakitList) DeepCopy() *DatakitList {
	if in == nil {
		return nil
	}
	out := new(DatakitList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject copying the receiver, creating a new runtime.Object.
func (in *DatakitList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *DatakitSpec) DeepCopyInto(out *DatakitSpec) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Instances != nil {
		in, out := &in.Instances, &out.Instances
		*out = make([]DatakitInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy copying the receiver, creating a new DatakitSpec.
func (in *DatakitSpec) DeepCopy() *DatakitSpec {
	if in == nil {
		return nil
	}
	out := new(DatakitSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copying the receiver, writing into out. in must be non-nil.
func (in *DatakitInstance) DeepCopyInto(out *DatakitInstance) {
	*out = *in
	return
}

// DeepCopy copying the receiver, creating a new DatakitSpec.
func (in *DatakitInstance) DeepCopy() *DatakitInstance {
	if in == nil {
		return nil
	}
	out := new(DatakitInstance)
	in.DeepCopyInto(out)
	return out
}
