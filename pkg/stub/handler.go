package stub

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kapis "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	imagev1client "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	templatev1client "github.com/openshift/client-go/template/clientset/versioned/typed/template/v1"
	operatorsv1api "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	"github.com/openshift/cluster-samples-operator/pkg/cache"
	operatorstatus "github.com/openshift/cluster-samples-operator/pkg/operatorstatus"
	sampleclientv1 "github.com/openshift/cluster-samples-operator/pkg/generated/clientset/versioned/typed/samples/v1"
)

const (
	x86OCPContentRootDir	= "/opt/openshift/operator/ocp-x86_64"
	installtypekey		= "keyForInstallTypeField"
	regkey			= "keyForSamplesRegistryField"
	skippedstreamskey	= "keyForSkippedImageStreamsField"
	skippedtempskey		= "keyForSkippedTemplatesField"
)

func NewSamplesOperatorHandler(kubeconfig *restclient.Config) (*Handler, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	h := &Handler{}
	h.initter = &defaultInClusterInitter{}
	h.initter.init(h, kubeconfig)
	crdWrapper := &generatedCRDWrapper{}
	client, err := sampleclientv1.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	crdWrapper.client = client.Configs()
	h.crdwrapper = crdWrapper
	h.Fileimagegetter = &DefaultImageStreamFromFileGetter{}
	h.Filetemplategetter = &DefaultTemplateFromFileGetter{}
	h.Filefinder = &DefaultResourceFileLister{}
	h.imageclientwrapper = &defaultImageStreamClientWrapper{h: h}
	h.templateclientwrapper = &defaultTemplateClientWrapper{h: h}
	h.secretclientwrapper = &defaultSecretClientWrapper{coreclient: h.coreclient}
	h.cvowrapper = operatorstatus.NewClusterOperatorHandler(h.configclient)
	h.skippedImagestreams = make(map[string]bool)
	h.skippedTemplates = make(map[string]bool)
	h.CreateDefaultResourceIfNeeded(nil)
	h.imagestreamFile = make(map[string]string)
	h.templateFile = make(map[string]string)
	h.mapsMutex = sync.Mutex{}
	h.version = os.Getenv("RELEASE_VERSION")
	return h, nil
}

type Handler struct {
	initter			InClusterInitter
	crdwrapper		CRDWrapper
	cvowrapper		*operatorstatus.ClusterOperatorHandler
	restconfig		*restclient.Config
	tempclient		*templatev1client.TemplateV1Client
	imageclient		*imagev1client.ImageV1Client
	coreclient		*corev1client.CoreV1Client
	configclient		*configv1client.ConfigV1Client
	imageclientwrapper	ImageStreamClientWrapper
	templateclientwrapper	TemplateClientWrapper
	secretclientwrapper	SecretClientWrapper
	Fileimagegetter		ImageStreamFromFileGetter
	Filetemplategetter	TemplateFromFileGetter
	Filefinder		ResourceFileLister
	skippedTemplates	map[string]bool
	skippedImagestreams	map[string]bool
	imagestreamFile		map[string]string
	templateFile		map[string]string
	mapsMutex		sync.Mutex
	secretRetryCount	int8
	version			string
}

func (h *Handler) prepSamplesWatchEvent(kind, name string, annotations map[string]string, deleted bool) (*v1.Config, string, bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	cfg, err := h.crdwrapper.Get(v1.ConfigName)
	if cfg == nil || err != nil {
		logrus.Printf("Received watch event %s but not upserting since not have the Config yet: %#v %#v", kind+"/"+name, err, cfg)
		return nil, "", false, err
	}
	if cfg.ConditionFalse(v1.ImageChangesInProgress) {
		if cfg.DeletionTimestamp != nil {
			logrus.Printf("Received watch event %s but not upserting since deletion of the Config is in progress and image changes are not in progress", kind+"/"+name)
			return nil, "", false, nil
		}
		switch cfg.Spec.ManagementState {
		case operatorsv1api.Removed:
			logrus.Debugf("Not upserting %s/%s event because operator is in removed state and image changes are not in progress", kind, name)
			return nil, "", false, nil
		case operatorsv1api.Unmanaged:
			logrus.Debugf("Not upserting %s/%s event because operator is in unmanaged state and image changes are not in progress", kind, name)
			return nil, "", false, nil
		}
	}
	filePath := ""
	h.buildFileMaps(cfg, false)
	h.buildSkipFilters(cfg)
	inInventory := false
	skipped := false
	switch kind {
	case "imagestream":
		filePath, inInventory = h.imagestreamFile[name]
		_, skipped = h.skippedImagestreams[name]
	case "template":
		filePath, inInventory = h.templateFile[name]
		_, skipped = h.skippedTemplates[name]
	}
	if !inInventory {
		logrus.Printf("watch event %s not part of operators inventory", name)
		return nil, "", false, nil
	}
	if skipped {
		logrus.Printf("watch event %s in skipped list for %s", name, kind)
		return cfg, "", false, nil
	}
	if deleted && (kind == "template" || cache.UpsertsAmount() == 0) {
		logrus.Printf("going to recreate deleted managed sample %s/%s", kind, name)
		return cfg, filePath, true, nil
	}
	if cfg.ConditionFalse(v1.ImageChangesInProgress) && (cfg.ConditionTrue(v1.MigrationInProgress) || h.version != cfg.Status.Version) {
		logrus.Printf("watch event for %s/%s while migration in progress, image in progress is false; will not update sample because of this event", kind, name)
		return cfg, "", false, nil
	}
	if annotations != nil {
		isv, ok := annotations[v1.SamplesVersionAnnotation]
		logrus.Debugf("Comparing %s/%s version %s ok %v with git version %s", kind, name, isv, ok, h.version)
		if ok && isv == h.version {
			logrus.Debugf("Not upserting %s/%s cause operator version matches", kind, name)
			return cfg, "", false, nil
		}
	}
	return cfg, filePath, true, nil
}
func (h *Handler) GoodConditionUpdate(cfg *v1.Config, newStatus corev1.ConditionStatus, conditionType v1.ConfigConditionType) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	logrus.Debugf("updating condition %s to %s", conditionType, newStatus)
	condition := cfg.Condition(conditionType)
	if condition.Status != newStatus {
		now := kapis.Now()
		condition.LastUpdateTime = now
		condition.Status = newStatus
		condition.LastTransitionTime = now
		condition.Message = ""
		condition.Reason = ""
		cfg.ConditionUpdate(condition)
	}
}
func IsRetryableAPIError(err error) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err == nil {
		return false
	}
	if kerrors.IsInternalError(err) || kerrors.IsTimeout(err) || kerrors.IsServerTimeout(err) || kerrors.IsTooManyRequests(err) || utilnet.IsProbableEOF(err) || utilnet.IsConnectionReset(err) {
		return true
	}
	if _, shouldRetry := kerrors.SuggestsClientDelay(err); shouldRetry {
		return true
	}
	return false
}
func (h *Handler) CreateDefaultResourceIfNeeded(cfg *v1.Config) (*v1.Config, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	deleteInProgress := cfg != nil && cfg.DeletionTimestamp != nil
	var err error
	if deleteInProgress {
		cfg = &v1.Config{}
		cfg.Name = v1.ConfigName
		cfg.Kind = "Config"
		cfg.APIVersion = v1.GroupName + "/" + v1.Version
		err = wait.PollImmediate(3*time.Second, 30*time.Second, func() (bool, error) {
			s, e := h.crdwrapper.Get(v1.ConfigName)
			if kerrors.IsNotFound(e) {
				return true, nil
			}
			if err != nil {
				logrus.Printf("create default config access error %v", err)
				return false, nil
			}
			if s == nil {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return nil, h.processError(cfg, v1.SamplesExist, corev1.ConditionUnknown, err, "issues waiting for delete to complete: %v")
		}
		cfg = nil
		logrus.Println("delete of Config recognized")
	}
	if cfg == nil || kerrors.IsNotFound(err) {
		cfg = &v1.Config{}
		cfg.Spec.SkippedTemplates = []string{}
		cfg.Spec.SkippedImagestreams = []string{}
		cfg.Status.SkippedImagestreams = []string{}
		cfg.Status.SkippedTemplates = []string{}
		cfg.Name = v1.ConfigName
		cfg.Kind = "Config"
		cfg.APIVersion = v1.GroupName + "/" + v1.Version
		cfg.Spec.Architectures = append(cfg.Spec.Architectures, v1.X86Architecture)
		cfg.Spec.ManagementState = operatorsv1api.Managed
		h.AddFinalizer(cfg)
		h.copyDefaultClusterPullSecret(nil)
		logrus.Println("creating default Config")
		err = h.crdwrapper.Create(cfg)
		if err != nil {
			if !kerrors.IsAlreadyExists(err) {
				return nil, err
			}
			logrus.Println("got already exists error on create default")
		}
	} else {
		logrus.Printf("Config %#v found during operator startup", cfg)
	}
	return cfg, nil
}
func (h *Handler) initConditions(cfg *v1.Config) *v1.Config {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	now := kapis.Now()
	cfg.Condition(v1.SamplesExist)
	cfg.Condition(v1.ImportCredentialsExist)
	valid := cfg.Condition(v1.ConfigurationValid)
	if valid.Status != corev1.ConditionTrue {
		valid.Status = corev1.ConditionTrue
		valid.LastUpdateTime = now
		valid.LastTransitionTime = now
		cfg.ConditionUpdate(valid)
	}
	cfg.Condition(v1.ImageChangesInProgress)
	cfg.Condition(v1.RemovePending)
	cfg.Condition(v1.MigrationInProgress)
	cfg.Condition(v1.ImportImageErrorsExist)
	return cfg
}
func (h *Handler) CleanUpOpenshiftNamespaceOnDelete(cfg *v1.Config) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	h.buildSkipFilters(cfg)
	iopts := metav1.ListOptions{LabelSelector: v1.SamplesManagedLabel + "=true"}
	streamList, err := h.imageclientwrapper.List("openshift", iopts)
	if err != nil && !kerrors.IsNotFound(err) {
		logrus.Warnf("Problem listing openshift imagestreams on Config delete: %#v", err)
		return err
	} else {
		if streamList.Items != nil {
			for _, stream := range streamList.Items {
				manage, ok := stream.Labels[v1.SamplesManagedLabel]
				if !ok || strings.TrimSpace(manage) != "true" {
					continue
				}
				err = h.imageclientwrapper.Delete("openshift", stream.Name, &metav1.DeleteOptions{})
				if err != nil && !kerrors.IsNotFound(err) {
					logrus.Warnf("Problem deleting openshift imagestream %s on Config delete: %#v", stream.Name, err)
					return err
				}
				cache.ImageStreamMassDeletesAdd(stream.Name)
			}
		}
	}
	cache.ClearUpsertsCache()
	tempList, err := h.templateclientwrapper.List("openshift", iopts)
	if err != nil && !kerrors.IsNotFound(err) {
		logrus.Warnf("Problem listing openshift templates on Config delete: %#v", err)
		return err
	} else {
		if tempList.Items != nil {
			for _, temp := range tempList.Items {
				manage, ok := temp.Labels[v1.SamplesManagedLabel]
				if !ok || strings.TrimSpace(manage) != "true" {
					continue
				}
				err = h.templateclientwrapper.Delete("openshift", temp.Name, &metav1.DeleteOptions{})
				if err != nil && !kerrors.IsNotFound(err) {
					logrus.Warnf("Problem deleting openshift template %s on Config delete: %#v", temp.Name, err)
					return err
				}
				cache.TemplateMassDeletesAdd(temp.Name)
			}
		}
	}
	return nil
}
func (h *Handler) Handle(event v1.Event) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch event.Object.(type) {
	case *imagev1.ImageStream:
		is, _ := event.Object.(*imagev1.ImageStream)
		if is.Namespace != "openshift" {
			return nil
		}
		err := h.processImageStreamWatchEvent(is, event.Deleted)
		return err
	case *templatev1.Template:
		t, _ := event.Object.(*templatev1.Template)
		if t.Namespace != "openshift" {
			return nil
		}
		err := h.processTemplateWatchEvent(t, event.Deleted)
		return err
	case *corev1.Secret:
		dockercfgSecret, _ := event.Object.(*corev1.Secret)
		if !secretsWeCareAbout(dockercfgSecret) {
			return nil
		}
		cfg, _ := h.crdwrapper.Get(v1.ConfigName)
		if cfg != nil {
			return h.processSecretEvent(cfg, dockercfgSecret, event)
		} else {
			return fmt.Errorf("Received secret %s but do not have the Config yet, requeuing", dockercfgSecret.Name)
		}
	case *v1.Config:
		cfg, _ := event.Object.(*v1.Config)
		if cfg.Name != v1.ConfigName || cfg.Namespace != "" {
			return nil
		}
		if event.Deleted {
			logrus.Info("A previous delete attempt has been successfully completed")
			return nil
		}
		if cfg.DeletionTimestamp != nil {
			if cfg.ConditionTrue(v1.ImageChangesInProgress) {
				return nil
			}
			if h.NeedsFinalizing(cfg) {
				logrus.Println("Initiating samples delete and marking exists false")
				err := h.CleanUpOpenshiftNamespaceOnDelete(cfg)
				if err != nil {
					return err
				}
				h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.SamplesExist)
				logrus.Printf("CRDUPDATE exist false update")
				err = h.crdwrapper.UpdateStatus(cfg)
				if err != nil {
					logrus.Printf("error on Config update after setting exists condition to false (returning error to retry): %v", err)
					return err
				}
			} else {
				logrus.Println("Initiating finalizer processing for a SampleResource delete attempt")
				h.RemoveFinalizer(cfg)
				logrus.Printf("CRDUPDATE remove finalizer update")
				err := h.crdwrapper.Update(cfg)
				if err != nil {
					logrus.Printf("error removing Config finalizer during delete (hopefully retry on return of error works): %v", err)
					return err
				}
				go func() {
					h.CreateDefaultResourceIfNeeded(cfg)
				}()
			}
			return nil
		}
		if cfg.ConditionUnknown(v1.ImportCredentialsExist) {
			err := h.copyDefaultClusterPullSecret(nil)
			if err == nil {
				h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.ImportCredentialsExist)
				logrus.Println("CRDUPDATE cleared import cred unknown")
				return h.crdwrapper.UpdateStatus(cfg)
			}
		}
		err := h.cvowrapper.UpdateOperatorStatus(cfg)
		if err != nil {
			logrus.Errorf("error updating cluster operator status: %v", err)
			return err
		}
		doit, cfgUpdate, err := h.ProcessManagementField(cfg)
		if !doit || err != nil {
			if err != nil || cfgUpdate {
				logrus.Printf("CRDUPDATE process mgmt update")
				h.crdwrapper.UpdateStatus(cfg)
			}
			return err
		}
		existingValidStatus := cfg.Condition(v1.ConfigurationValid).Status
		err = h.SpecValidation(cfg)
		if err != nil {
			logrus.Printf("CRDUPDATE bad spec validation update")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if existingValidStatus != cfg.Condition(v1.ConfigurationValid).Status {
			logrus.Printf("CRDUPDATE spec corrected")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		h.buildSkipFilters(cfg)
		configChanged := false
		configChangeRequiresUpsert := false
		configChangeRequiresImportErrorUpdate := false
		registryChanged := false
		unskippedStreams := map[string]bool{}
		unskippedTemplates := map[string]bool{}
		if cfg.Spec.ManagementState == cfg.Status.ManagementState {
			configChanged, configChangeRequiresUpsert, configChangeRequiresImportErrorUpdate, registryChanged, unskippedStreams, unskippedTemplates = h.VariableConfigChanged(cfg)
			logrus.Debugf("config changed %v upsert needed %v import error upd needed %v exists/true %v progressing/false %v op version %s status version %s", configChanged, configChangeRequiresUpsert, configChangeRequiresImportErrorUpdate, cfg.ConditionTrue(v1.SamplesExist), cfg.ConditionFalse(v1.ImageChangesInProgress), h.version, cfg.Status.Version)
			if !configChanged && cfg.ConditionTrue(v1.SamplesExist) && cfg.ConditionFalse(v1.ImageChangesInProgress) && h.version == cfg.Status.Version {
				logrus.Debugf("At steady state: config the same and exists is true, in progress false, and version correct")
				if cfg.ConditionTrue(v1.MigrationInProgress) {
					h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.MigrationInProgress)
					logrus.Println("CRDUPDATE turn migration off")
					return h.crdwrapper.UpdateStatus(cfg)
				}
				h.buildFileMaps(cfg, false)
				return h.createSamples(cfg, false, registryChanged, unskippedStreams, unskippedTemplates)
			}
			if configChangeRequiresUpsert && cfg.ConditionTrue(v1.ImageChangesInProgress) {
				h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.ImageChangesInProgress)
				logrus.Printf("CRDUPDATE change in progress from true to false for config change")
				return h.crdwrapper.UpdateStatus(cfg)
			}
		}
		cfg.Status.ManagementState = operatorsv1api.Managed
		stillWaitingForSecret, callSDKToUpdate := h.WaitingForCredential(cfg)
		if callSDKToUpdate {
			logrus.Println("CRDUPDATE Config update ignored since need the RHEL credential")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if stillWaitingForSecret {
			return nil
		}
		if cfg.ConditionFalse(v1.MigrationInProgress) && len(cfg.Status.Version) > 0 && h.version != cfg.Status.Version {
			logrus.Printf("Undergoing migration from %s to %s", cfg.Status.Version, h.version)
			h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.MigrationInProgress)
			logrus.Println("CRDUPDATE turn migration on")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if !configChanged && cfg.ConditionTrue(v1.SamplesExist) && cfg.ConditionFalse(v1.ImageChangesInProgress) && cfg.Condition(v1.MigrationInProgress).LastUpdateTime.Before(&cfg.Condition(v1.ImageChangesInProgress).LastUpdateTime) && h.version != cfg.Status.Version {
			if cfg.ConditionTrue(v1.ImportImageErrorsExist) {
				logrus.Printf("An image import error occurred applying the latest configuration on version %s; this operator will periodically retry the import, or an administrator can investigate and remedy manually", h.version)
			}
			cfg.Status.Version = h.version
			logrus.Printf("The samples are now at version %s", cfg.Status.Version)
			logrus.Println("CRDUPDATE upd status version")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if len(cfg.Spec.Architectures) == 0 {
			cfg.Spec.Architectures = append(cfg.Spec.Architectures, v1.X86Architecture)
		}
		h.StoreCurrentValidConfig(cfg)
		for _, name := range cfg.Status.SkippedTemplates {
			h.setSampleManagedLabelToFalse("template", name)
		}
		for _, name := range cfg.Status.SkippedImagestreams {
			h.setSampleManagedLabelToFalse("imagestream", name)
		}
		if configChangeRequiresImportErrorUpdate && !configChangeRequiresUpsert {
			logrus.Printf("CRDUPDATE config change did not require upsert but did change import errors")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if configChanged && !configChangeRequiresUpsert && cfg.ConditionTrue(v1.SamplesExist) {
			logrus.Printf("CRDUPDATE bypassing upserts for non invasive config change after initial create")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if !cfg.ConditionTrue(v1.ImageChangesInProgress) {
			err = h.buildFileMaps(cfg, true)
			if err != nil {
				return err
			}
			err = h.createSamples(cfg, true, registryChanged, unskippedStreams, unskippedTemplates)
			if err != nil {
				h.processError(cfg, v1.ImageChangesInProgress, corev1.ConditionUnknown, err, "error creating samples: %v")
				logrus.Printf("CRDUPDATE setting in progress to unknown")
				e := h.crdwrapper.UpdateStatus(cfg)
				if e != nil {
					return e
				}
				return err
			}
			now := kapis.Now()
			progressing := cfg.Condition(v1.ImageChangesInProgress)
			progressing.LastUpdateTime = now
			progressing.LastTransitionTime = now
			logrus.Debugf("Handle changing processing from false to true")
			progressing.Status = corev1.ConditionTrue
			for isName := range h.imagestreamFile {
				_, skipped := h.skippedImagestreams[isName]
				unskipping := len(unskippedStreams) > 0
				_, unskipped := unskippedStreams[isName]
				if unskipping && !unskipped {
					continue
				}
				if !cfg.NameInReason(progressing.Reason, isName) && !skipped {
					progressing.Reason = progressing.Reason + isName + " "
				}
			}
			logrus.Debugf("Handle Reason field set to %s", progressing.Reason)
			cfg.ConditionUpdate(progressing)
			cfg = h.initConditions(cfg)
			logrus.Printf("CRDUPDATE progressing true update")
			err = h.crdwrapper.UpdateStatus(cfg)
			if err != nil {
				return err
			}
			return nil
		}
		if !cfg.ConditionTrue(v1.SamplesExist) {
			h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.SamplesExist)
			logrus.Printf("CRDUPDATE exist true update")
			return h.crdwrapper.UpdateStatus(cfg)
		}
		if cache.AllUpsertEventsArrived() && cfg.ConditionTrue(v1.ImageChangesInProgress) {
			keysToClear := []string{}
			for key, is := range cache.GetUpsertImageStreams() {
				if is == nil {
					var e error
					is, e = h.imageclientwrapper.Get("openshift", key, metav1.GetOptions{})
					if e != nil {
						keysToClear = append(keysToClear, key)
						continue
					}
				}
				cfg, _ = h.processImportStatus(is, cfg)
			}
			for _, key := range keysToClear {
				cache.RemoveUpsert(key)
				cfg.ClearNameInReason(cfg.Condition(v1.ImageChangesInProgress).Reason, key)
				cfg.ClearNameInReason(cfg.Condition(v1.ImportImageErrorsExist).Reason, key)
			}
			if len(strings.TrimSpace(cfg.Condition(v1.ImageChangesInProgress).Reason)) == 0 {
				h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.ImageChangesInProgress)
				logrus.Println("The last in progress imagestream has completed (config event loop)")
			}
			if cfg.ConditionFalse(v1.ImageChangesInProgress) {
				logrus.Printf("CRDUPDATE setting in progress to false after examining cached imagestream events")
				err = h.crdwrapper.UpdateStatus(cfg)
				if err == nil {
					cache.ClearUpsertsCache()
				}
				return err
			}
		}
	}
	return nil
}
func (h *Handler) setSampleManagedLabelToFalse(kind, name string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var err error
	switch kind {
	case "imagestream":
		var stream *imagev1.ImageStream
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			stream, err = h.imageclientwrapper.Get("openshift", name, metav1.GetOptions{})
			if err == nil && stream != nil && stream.Labels != nil {
				label, _ := stream.Labels[v1.SamplesManagedLabel]
				if label == "true" {
					stream.Labels[v1.SamplesManagedLabel] = "false"
					_, err = h.imageclientwrapper.Update("openshift", stream)
				}
			}
			return err
		})
	case "template":
		var tpl *templatev1.Template
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			tpl, err = h.templateclientwrapper.Get("openshift", name, metav1.GetOptions{})
			if err == nil && tpl != nil && tpl.Labels != nil {
				label, _ := tpl.Labels[v1.SamplesManagedLabel]
				if label == "true" {
					tpl.Labels[v1.SamplesManagedLabel] = "false"
					_, err = h.templateclientwrapper.Update("openshift", tpl)
				}
			}
			return err
		})
	}
	return nil
}
func (h *Handler) createSamples(cfg *v1.Config, updateIfPresent, registryChanged bool, unskippedStreams, unskippedTemplates map[string]bool) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	imagestreams := []*imagev1.ImageStream{}
	for _, fileName := range h.imagestreamFile {
		imagestream, err := h.Fileimagegetter.Get(fileName)
		if err != nil {
			return err
		}
		if len(unskippedStreams) > 0 {
			if _, ok := unskippedStreams[imagestream.Name]; !ok {
				continue
			}
		}
		if _, isok := h.skippedImagestreams[imagestream.Name]; !isok {
			if updateIfPresent {
				cache.AddUpsert(imagestream.Name)
			}
			imagestreams = append(imagestreams, imagestream)
		}
	}
	for _, imagestream := range imagestreams {
		is, err := h.imageclientwrapper.Get("openshift", imagestream.Name, metav1.GetOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return err
		}
		if err == nil && !updateIfPresent {
			continue
		}
		if kerrors.IsNotFound(err) {
			is = nil
		}
		err = h.upsertImageStream(imagestream, is, cfg)
		if err != nil {
			if updateIfPresent {
				cache.RemoveUpsert(imagestream.Name)
			}
			return err
		}
	}
	if len(unskippedTemplates) == 0 && registryChanged && cfg.ConditionTrue(v1.SamplesExist) && cfg.ConditionFalse(v1.MigrationInProgress) {
		return nil
	}
	for _, fileName := range h.templateFile {
		template, err := h.Filetemplategetter.Get(fileName)
		if err != nil {
			return err
		}
		if len(unskippedTemplates) > 0 {
			if _, ok := unskippedTemplates[template.Name]; !ok {
				continue
			}
		}
		t, err := h.templateclientwrapper.Get("openshift", template.Name, metav1.GetOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return err
		}
		if err == nil && !updateIfPresent {
			continue
		}
		if kerrors.IsNotFound(err) {
			t = nil
		}
		err = h.upsertTemplate(template, t, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}
func getTemplateClient(restconfig *restclient.Config) (*templatev1client.TemplateV1Client, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return templatev1client.NewForConfig(restconfig)
}
func getImageClient(restconfig *restclient.Config) (*imagev1client.ImageV1Client, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return imagev1client.NewForConfig(restconfig)
}
func GetNamespace() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	b, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + corev1.ServiceAccountNamespaceKey)
	return string(b)
}
