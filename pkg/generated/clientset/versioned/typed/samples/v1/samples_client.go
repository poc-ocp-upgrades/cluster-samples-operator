package v1

import (
	v1 "github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	"github.com/openshift/cluster-samples-operator/pkg/generated/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type SamplesV1Interface interface {
	RESTClient() rest.Interface
	ConfigsGetter
}
type SamplesV1Client struct{ restClient rest.Interface }

func (c *SamplesV1Client) Configs() ConfigInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return newConfigs(c)
}
func NewForConfig(c *rest.Config) (*SamplesV1Client, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &SamplesV1Client{client}, nil
}
func NewForConfigOrDie(c *rest.Config) *SamplesV1Client {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}
func New(c rest.Interface) *SamplesV1Client {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &SamplesV1Client{c}
}
func setConfigDefaults(config *rest.Config) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	return nil
}
func (c *SamplesV1Client) RESTClient() rest.Interface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if c == nil {
		return nil
	}
	return c.restClient
}
