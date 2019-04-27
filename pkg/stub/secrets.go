package stub

import (
	"fmt"
	operatorsv1api "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	"github.com/openshift/cluster-samples-operator/pkg/cache"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	coreosPullSecretNamespace	= "openshift-config"
	coreosPullSecretName		= "pull-secret"
)

func (h *Handler) copyDefaultClusterPullSecret(secret *corev1.Secret) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var err error
	if secret == nil {
		secret, err = h.secretclientwrapper.Get(coreosPullSecretNamespace, coreosPullSecretName)
		if err != nil {
			return err
		}
		if secret == nil {
			return nil
		}
	}
	logrus.Printf("Copying secret %s from the %s namespace into the operator's namespace", coreosPullSecretName, coreosPullSecretNamespace)
	secretToCreate := corev1.Secret{}
	secret.DeepCopyInto(&secretToCreate)
	secretToCreate.Name = v1.SamplesRegistryCredentials
	secretToCreate.Namespace = ""
	secretToCreate.ResourceVersion = ""
	secretToCreate.UID = ""
	secretToCreate.Annotations = make(map[string]string)
	secretToCreate.Annotations[v1.SamplesVersionAnnotation] = h.version
	_, err = h.secretclientwrapper.Create("openshift", &secretToCreate)
	if kerrors.IsAlreadyExists(err) {
		_, err = h.secretclientwrapper.Update("openshift", &secretToCreate)
	}
	return err
}
func secretsWeCareAbout(secret *corev1.Secret) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	kubeSecret := secret.Name == coreosPullSecretName && secret.Namespace == coreosPullSecretNamespace
	openshiftSecret := secret.Name == v1.SamplesRegistryCredentials && secret.Namespace == "openshift"
	return kubeSecret || openshiftSecret
}
func (h *Handler) manageDockerCfgSecret(deleted bool, cfg *v1.Config, secret *corev1.Secret) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !secretsWeCareAbout(secret) {
		return nil
	}
	switch secret.Name {
	case v1.SamplesRegistryCredentials:
		if deleted {
			err := h.copyDefaultClusterPullSecret(nil)
			if err != nil {
				if kerrors.IsNotFound(err) {
					h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.ImportCredentialsExist)
					return nil
				}
				return err
			}
			h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.ImportCredentialsExist)
			return nil
		}
	case coreosPullSecretName:
		if deleted {
			err := h.secretclientwrapper.Delete("openshift", v1.SamplesRegistryCredentials, &metav1.DeleteOptions{})
			if err != nil && !kerrors.IsNotFound(err) {
				return err
			}
			logrus.Printf("registry dockerconfig secret %s was deleted from the %s namespacae so deleted secret %s in the openshift namespace", secret.Name, secret.Namespace, v1.SamplesRegistryCredentials)
			h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.ImportCredentialsExist)
			return nil
		}
		err := h.copyDefaultClusterPullSecret(secret)
		if err == nil {
			h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.ImportCredentialsExist)
		}
		return err
	}
	return nil
}
func (h *Handler) WaitingForCredential(cfg *v1.Config) (bool, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, err := h.secretclientwrapper.Get("openshift", v1.SamplesRegistryCredentials)
	if err != nil {
		cred := cfg.Condition(v1.ImportCredentialsExist)
		if len(cred.Message) > 0 {
			return true, false
		}
		err := fmt.Errorf("Cannot create rhel imagestreams to registry.redhat.io without the credentials being available")
		h.processError(cfg, v1.ImportCredentialsExist, corev1.ConditionFalse, err, "%v")
		return true, true
	}
	if !cfg.ConditionTrue(v1.ImportCredentialsExist) {
		h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.ImportCredentialsExist)
	}
	return false, false
}
func (h *Handler) processSecretEvent(cfg *v1.Config, dockercfgSecret *corev1.Secret, event v1.Event) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if cache.UpsertsAmount() > 0 {
		return fmt.Errorf("retry secret event because in the middle of an sample upsert cycle")
	}
	removedState := false
	switch cfg.Spec.ManagementState {
	case operatorsv1api.Removed:
		logrus.Printf("processing secret watch event while in Removed state; deletion event: %v", event.Deleted)
		removedState = true
	case operatorsv1api.Unmanaged:
		logrus.Debugln("Ignoring secret event because samples resource is in unmanaged state")
		return nil
	case operatorsv1api.Managed:
		logrus.Printf("processing secret watch event while in Managed state; deletion event: %v", event.Deleted)
	default:
		logrus.Printf("processing secret watch event like we are in Managed state, even though it is set to %v; deletion event %v", cfg.Spec.ManagementState, event.Deleted)
	}
	deleted := event.Deleted
	if dockercfgSecret.Namespace == "openshift" {
		if !deleted {
			if dockercfgSecret.Annotations != nil {
				_, ok := dockercfgSecret.Annotations[v1.SamplesVersionAnnotation]
				if ok {
					logrus.Println("creation/update of credential in openshift namespace recognized")
					if !cfg.ConditionTrue(v1.ImportCredentialsExist) {
						h.GoodConditionUpdate(cfg, corev1.ConditionTrue, v1.ImportCredentialsExist)
						logrus.Printf("CRDUPDATE switching import cred to true following openshift namespace event")
						return h.crdwrapper.UpdateStatus(cfg)
					}
					return nil
				}
			}
			err := fmt.Errorf("the samples credential was created/updated in the openshift namespace without the version annotation")
			return h.processError(cfg, v1.ImportCredentialsExist, corev1.ConditionUnknown, err, "%v")
		}
		if cfg.ConditionTrue(v1.ImportCredentialsExist) {
			if h.secretRetryCount < 3 {
				err := fmt.Errorf("retry on credential deletion in the openshift namespace to make sure the operator deleted it")
				h.secretRetryCount++
				return err
			}
		}
		if removedState {
			logrus.Println("deletion of credential in openshift namespace for removed state recognized")
			h.GoodConditionUpdate(cfg, corev1.ConditionFalse, v1.ImportCredentialsExist)
			logrus.Printf("CRDUPDATE secret deletion recognized")
			return h.crdwrapper.UpdateStatus(cfg)
		}
	}
	h.secretRetryCount = 0
	if removedState {
		return nil
	}
	beforeStatus := cfg.Condition(v1.ImportCredentialsExist).Status
	err := h.manageDockerCfgSecret(deleted, cfg, dockercfgSecret)
	if err != nil {
		h.processError(cfg, v1.ImportCredentialsExist, corev1.ConditionUnknown, err, "%v")
		logrus.Printf("CRDUPDATE event secret update error")
	} else {
		afterStatus := cfg.Condition(v1.ImportCredentialsExist).Status
		if beforeStatus == afterStatus {
			return nil
		}
		logrus.Printf("CRDUPDATE event secret update")
	}
	return h.crdwrapper.UpdateStatus(cfg)
}
