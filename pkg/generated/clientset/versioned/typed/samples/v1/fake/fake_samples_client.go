package fake

import (
	v1 "github.com/openshift/cluster-samples-operator/pkg/generated/clientset/versioned/typed/samples/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeSamplesV1 struct{ *testing.Fake }

func (c *FakeSamplesV1) Configs() v1.ConfigInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &FakeConfigs{c}
}
func (c *FakeSamplesV1) RESTClient() rest.Interface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var ret *rest.RESTClient
	return ret
}
