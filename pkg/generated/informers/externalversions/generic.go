package externalversions

import (
	"fmt"
	v1 "github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}
type genericInformer struct {
	informer	cache.SharedIndexInformer
	resource	schema.GroupResource
}

func (f *genericInformer) Informer() cache.SharedIndexInformer {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return f.informer
}
func (f *genericInformer) Lister() cache.GenericLister {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch resource {
	case v1.SchemeGroupVersion.WithResource("configs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Samples().V1().Configs().Informer()}, nil
	}
	return nil, fmt.Errorf("no informer found for %v", resource)
}
