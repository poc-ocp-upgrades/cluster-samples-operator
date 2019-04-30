package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *Config) DeepCopyInto(out *Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}
func (in *Config) DeepCopy() *Config {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}
func (in *Config) DeepCopyObject() runtime.Object {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (in *ConfigCondition) DeepCopyInto(out *ConfigCondition) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}
func (in *ConfigCondition) DeepCopy() *ConfigCondition {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if in == nil {
		return nil
	}
	out := new(ConfigCondition)
	in.DeepCopyInto(out)
	return out
}
func (in *ConfigList) DeepCopyInto(out *ConfigList) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Config, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}
func (in *ConfigList) DeepCopy() *ConfigList {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if in == nil {
		return nil
	}
	out := new(ConfigList)
	in.DeepCopyInto(out)
	return out
}
func (in *ConfigList) DeepCopyObject() runtime.Object {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (in *ConfigSpec) DeepCopyInto(out *ConfigSpec) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	*out = *in
	if in.Architectures != nil {
		in, out := &in.Architectures, &out.Architectures
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkippedImagestreams != nil {
		in, out := &in.SkippedImagestreams, &out.SkippedImagestreams
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkippedTemplates != nil {
		in, out := &in.SkippedTemplates, &out.SkippedTemplates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}
func (in *ConfigSpec) DeepCopy() *ConfigSpec {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if in == nil {
		return nil
	}
	out := new(ConfigSpec)
	in.DeepCopyInto(out)
	return out
}
func (in *ConfigStatus) DeepCopyInto(out *ConfigStatus) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ConfigCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Architectures != nil {
		in, out := &in.Architectures, &out.Architectures
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkippedImagestreams != nil {
		in, out := &in.SkippedImagestreams, &out.SkippedImagestreams
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SkippedTemplates != nil {
		in, out := &in.SkippedTemplates, &out.SkippedTemplates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}
func (in *ConfigStatus) DeepCopy() *ConfigStatus {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if in == nil {
		return nil
	}
	out := new(ConfigStatus)
	in.DeepCopyInto(out)
	return out
}
func (in *Event) DeepCopyInto(out *Event) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	*out = *in
	if in.Object != nil {
		out.Object = in.Object.DeepCopyObject()
	}
	return
}
func (in *Event) DeepCopy() *Event {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if in == nil {
		return nil
	}
	out := new(Event)
	in.DeepCopyInto(out)
	return out
}
