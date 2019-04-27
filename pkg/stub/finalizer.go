package stub

import "github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"

func (h *Handler) AddFinalizer(cfg *v1.Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	hasFinalizer := false
	for _, f := range cfg.Finalizers {
		if f == v1.ConfigFinalizer {
			hasFinalizer = true
			break
		}
	}
	if !hasFinalizer {
		cfg.Finalizers = append(cfg.Finalizers, v1.ConfigFinalizer)
	}
}
func (h *Handler) RemoveFinalizer(cfg *v1.Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	newFinalizers := []string{}
	for _, f := range cfg.Finalizers {
		if f == v1.ConfigFinalizer {
			continue
		}
		newFinalizers = append(newFinalizers, f)
	}
	cfg.Finalizers = newFinalizers
}
func (h *Handler) NeedsFinalizing(cfg *v1.Config) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if cfg.ConditionFalse(v1.SamplesExist) {
		return false
	}
	for _, f := range cfg.Finalizers {
		if f == v1.ConfigFinalizer {
			return true
		}
	}
	return false
}
