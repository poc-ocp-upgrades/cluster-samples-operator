package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Version		= "v1"
	GroupName	= "samples.operator.openshift.io"
)

var (
	scheme			= runtime.NewScheme()
	SchemeBuilder		= runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme		= SchemeBuilder.AddToScheme
	SchemeGroupVersion	= schema.GroupVersion{Group: GroupName, Version: Version}
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	AddToScheme(scheme)
}
func addKnownTypes(scheme *runtime.Scheme) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	scheme.AddKnownTypes(SchemeGroupVersion, &Config{}, &ConfigList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
func Kind(kind string) schema.GroupKind {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}
func Resource(resource string) schema.GroupResource {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
