//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 CMU-SV.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosClusterIssuer) DeepCopyInto(out *ChaosClusterIssuer) {
*out = *in
out.TypeMeta = in.TypeMeta
in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
out.Spec = in.Spec
in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosClusterIssuer.
func (in *ChaosClusterIssuer) DeepCopy() *ChaosClusterIssuer {
	if in == nil { return nil }
	out := new(ChaosClusterIssuer)
	in.DeepCopyInto(out)
	return out
}


// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ChaosClusterIssuer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosClusterIssuerList) DeepCopyInto(out *ChaosClusterIssuerList) {
*out = *in
out.TypeMeta = in.TypeMeta
in.ListMeta.DeepCopyInto(&out.ListMeta)
if in.Items != nil {
in, out := &in.Items, &out.Items
*out = make([]ChaosClusterIssuer, len(*in))
for i := range *in {
(*in)[i].DeepCopyInto(&(*out)[i])
}
}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosClusterIssuerList.
func (in *ChaosClusterIssuerList) DeepCopy() *ChaosClusterIssuerList {
	if in == nil { return nil }
	out := new(ChaosClusterIssuerList)
	in.DeepCopyInto(out)
	return out
}


// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ChaosClusterIssuerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosIssuer) DeepCopyInto(out *ChaosIssuer) {
*out = *in
out.TypeMeta = in.TypeMeta
in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
out.Spec = in.Spec
in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosIssuer.
func (in *ChaosIssuer) DeepCopy() *ChaosIssuer {
	if in == nil { return nil }
	out := new(ChaosIssuer)
	in.DeepCopyInto(out)
	return out
}


// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ChaosIssuer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosIssuerList) DeepCopyInto(out *ChaosIssuerList) {
*out = *in
out.TypeMeta = in.TypeMeta
in.ListMeta.DeepCopyInto(&out.ListMeta)
if in.Items != nil {
in, out := &in.Items, &out.Items
*out = make([]ChaosIssuer, len(*in))
for i := range *in {
(*in)[i].DeepCopyInto(&(*out)[i])
}
}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosIssuerList.
func (in *ChaosIssuerList) DeepCopy() *ChaosIssuerList {
	if in == nil { return nil }
	out := new(ChaosIssuerList)
	in.DeepCopyInto(out)
	return out
}


// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ChaosIssuerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosIssuerSpec) DeepCopyInto(out *ChaosIssuerSpec) {
*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosIssuerSpec.
func (in *ChaosIssuerSpec) DeepCopy() *ChaosIssuerSpec {
	if in == nil { return nil }
	out := new(ChaosIssuerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosIssuerStatus) DeepCopyInto(out *ChaosIssuerStatus) {
*out = *in
if in.Conditions != nil {
in, out := &in.Conditions, &out.Conditions
*out = make([]invalid type, len(*in))
for i := range *in {
}
}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosIssuerStatus.
func (in *ChaosIssuerStatus) DeepCopy() *ChaosIssuerStatus {
	if in == nil { return nil }
	out := new(ChaosIssuerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IssuerCondition) DeepCopyInto(out *IssuerCondition) {
*out = *in
if in.LastTransitionTime != nil {
in, out := &in.LastTransitionTime, &out.LastTransitionTime
*out = (*in).DeepCopy()
}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IssuerCondition.
func (in *IssuerCondition) DeepCopy() *IssuerCondition {
	if in == nil { return nil }
	out := new(IssuerCondition)
	in.DeepCopyInto(out)
	return out
}
