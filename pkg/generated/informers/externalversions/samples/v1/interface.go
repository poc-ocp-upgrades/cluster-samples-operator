package v1

import (
	internalinterfaces "github.com/openshift/cluster-samples-operator/pkg/generated/informers/externalversions/internalinterfaces"
)

type Interface interface{ Configs() ConfigInformer }
type version struct {
	factory			internalinterfaces.SharedInformerFactory
	namespace		string
	tweakListOptions	internalinterfaces.TweakListOptionsFunc
}

func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}
func (v *version) Configs() ConfigInformer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &configInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}
