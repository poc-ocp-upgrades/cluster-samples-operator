package fake

import (
	samplesv1 "github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

type FakeConfigs struct{ Fake *FakeSamplesV1 }

var configsResource = schema.GroupVersionResource{Group: "samples.operator.openshift.io", Version: "v1", Resource: "configs"}
var configsKind = schema.GroupVersionKind{Group: "samples.operator.openshift.io", Version: "v1", Kind: "Config"}

func (c *FakeConfigs) Get(name string, options v1.GetOptions) (result *samplesv1.Config, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, err := c.Fake.Invokes(testing.NewRootGetAction(configsResource, name), &samplesv1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*samplesv1.Config), err
}
func (c *FakeConfigs) List(opts v1.ListOptions) (result *samplesv1.ConfigList, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, err := c.Fake.Invokes(testing.NewRootListAction(configsResource, configsKind, opts), &samplesv1.ConfigList{})
	if obj == nil {
		return nil, err
	}
	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &samplesv1.ConfigList{ListMeta: obj.(*samplesv1.ConfigList).ListMeta}
	for _, item := range obj.(*samplesv1.ConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}
func (c *FakeConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.Fake.InvokesWatch(testing.NewRootWatchAction(configsResource, opts))
}
func (c *FakeConfigs) Create(config *samplesv1.Config) (result *samplesv1.Config, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, err := c.Fake.Invokes(testing.NewRootCreateAction(configsResource, config), &samplesv1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*samplesv1.Config), err
}
func (c *FakeConfigs) Update(config *samplesv1.Config) (result *samplesv1.Config, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, err := c.Fake.Invokes(testing.NewRootUpdateAction(configsResource, config), &samplesv1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*samplesv1.Config), err
}
func (c *FakeConfigs) UpdateStatus(config *samplesv1.Config) (*samplesv1.Config, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, err := c.Fake.Invokes(testing.NewRootUpdateSubresourceAction(configsResource, "status", config), &samplesv1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*samplesv1.Config), err
}
func (c *FakeConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, err := c.Fake.Invokes(testing.NewRootDeleteAction(configsResource, name), &samplesv1.Config{})
	return err
}
func (c *FakeConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	action := testing.NewRootDeleteCollectionAction(configsResource, listOptions)
	_, err := c.Fake.Invokes(action, &samplesv1.ConfigList{})
	return err
}
func (c *FakeConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *samplesv1.Config, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, err := c.Fake.Invokes(testing.NewRootPatchSubresourceAction(configsResource, name, data, subresources...), &samplesv1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*samplesv1.Config), err
}
