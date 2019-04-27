package v1

import (
	"fmt"
	"os"
	"strings"
	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ConfigList struct {
	metav1.TypeMeta	`json:",inline"`
	metav1.ListMeta	`json:"metadata"`
	Items		[]Config	`json:"items"`
}
type Config struct {
	metav1.TypeMeta		`json:",inline"`
	metav1.ObjectMeta	`json:"metadata"`
	Spec			ConfigSpec	`json:"spec"`
	Status			ConfigStatus	`json:"status,omitempty"`
}

const (
	SamplesRegistryCredentials		= "samples-registry-credentials"
	ConfigName				= "cluster"
	X86Architecture				= "x86_64"
	ConfigFinalizer				= GroupName + "/finalizer"
	SamplesManagedLabel			= GroupName + "/managed"
	SamplesVersionAnnotation		= GroupName + "/version"
	SamplesRecreateCredentialAnnotation	= GroupName + "/recreate"
)

type ConfigSpec struct {
	ManagementState		operatorv1.ManagementState	`json:"managementState,omitempty" protobuf:"bytes,1,opt,name=managementState"`
	SamplesRegistry		string				`json:"samplesRegistry,omitempty" protobuf:"bytes,2,opt,name=samplesRegistry"`
	Architectures		[]string			`json:"architectures,omitempty" protobuf:"bytes,4,opt,name=architectures"`
	SkippedImagestreams	[]string			`json:"skippedImagestreams,omitempty" protobuf:"bytes,5,opt,name=skippedImagestreams"`
	SkippedTemplates	[]string			`json:"skippedTemplates,omitempty" protobuf:"bytes,6,opt,name=skippedTemplates"`
}
type ConfigStatus struct {
	ManagementState		operatorv1.ManagementState	`json:"managementState,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=managementState"`
	Conditions		[]ConfigCondition		`json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
	SamplesRegistry		string				`json:"samplesRegistry,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,3,rep,name=samplesRegistry"`
	Architectures		[]string			`json:"architectures,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,5,rep,name=architectures"`
	SkippedImagestreams	[]string			`json:"skippedImagestreams,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,6,rep,name=skippedImagestreams"`
	SkippedTemplates	[]string			`json:"skippedTemplates,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,7,rep,name=skippedTemplates"`
	Version			string				`json:"version,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,8,rep,name=version"`
}
type ConfigConditionType string

const (
	ImportCredentialsExist	ConfigConditionType	= "ImportCredentialsExist"
	SamplesExist		ConfigConditionType	= "SamplesExist"
	ConfigurationValid	ConfigConditionType	= "ConfigurationValid"
	ImageChangesInProgress	ConfigConditionType	= "ImageChangesInProgress"
	RemovePending		ConfigConditionType	= "RemovePending"
	MigrationInProgress	ConfigConditionType	= "MigrationInProgress"
	ImportImageErrorsExist	ConfigConditionType	= "ImportImageErrorsExist"
	numconfigConditionType				= 7
)

type ConfigCondition struct {
	Type			ConfigConditionType	`json:"type" protobuf:"bytes,1,opt,name=type,casttype=ConfigConditionType"`
	Status			corev1.ConditionStatus	`json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/kubernetes/pkg/api/v1.ConditionStatus"`
	LastUpdateTime		metav1.Time		`json:"lastUpdateTime,omitempty" protobuf:"bytes,3,opt,name=lastUpdateTime"`
	LastTransitionTime	metav1.Time		`json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	Reason			string			`json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	Message			string			`json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

func (s *Config) ConditionTrue(c ConfigConditionType) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if s.Status.Conditions == nil {
		return false
	}
	for _, rc := range s.Status.Conditions {
		if rc.Type == c && rc.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
func (s *Config) ConditionFalse(c ConfigConditionType) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if s.Status.Conditions == nil {
		return false
	}
	for _, rc := range s.Status.Conditions {
		if rc.Type == c && rc.Status == corev1.ConditionFalse {
			return true
		}
	}
	return false
}
func (s *Config) ConditionUnknown(c ConfigConditionType) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if s.Status.Conditions == nil {
		return false
	}
	for _, rc := range s.Status.Conditions {
		if rc.Type == c && rc.Status == corev1.ConditionUnknown {
			return true
		}
	}
	return false
}
func (s *Config) AnyConditionUnknown() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, rc := range s.Status.Conditions {
		if rc.Status == corev1.ConditionUnknown {
			return true
		}
	}
	return false
}
func (s *Config) ConditionsMessages() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	consolidatedMessage := ""
	for _, c := range s.Status.Conditions {
		if len(c.Message) > 0 {
			consolidatedMessage = consolidatedMessage + c.Message + ";"
		}
	}
	return consolidatedMessage
}
func (s *Config) ConditionUpdate(c *ConfigCondition) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if s.Status.Conditions == nil {
		return
	}
	for i, ec := range s.Status.Conditions {
		if ec.Type == c.Type {
			s.Status.Conditions[i] = *c
			return
		}
	}
}
func (s *Config) Condition(c ConfigConditionType) *ConfigCondition {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if s.Status.Conditions != nil {
		for _, rc := range s.Status.Conditions {
			if rc.Type == c {
				return &rc
			}
		}
	}
	now := metav1.Now()
	newCondition := ConfigCondition{Type: c, Status: corev1.ConditionFalse, LastTransitionTime: now, LastUpdateTime: now}
	s.Status.Conditions = append(s.Status.Conditions, newCondition)
	return &newCondition
}
func (s *Config) NameInReason(reason, name string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch {
	case strings.Index(reason, name+" ") == 0:
		return true
	case strings.Contains(reason, " "+name+" "):
		return true
	default:
		return false
	}
}
func (s *Config) ClearNameInReason(reason, name string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch {
	case strings.Index(reason, name+" ") == 0:
		return strings.Replace(reason, name+" ", "", 1)
	case strings.Contains(reason, " "+name+" "):
		return strings.Replace(reason, " "+name+" ", " ", 1)
	default:
		return reason
	}
}

const (
	noInstallDetailed	= "Samples installation in error at %s: %s"
	installed		= "Samples installation successful at %s"
	moving			= "Samples processing to %s"
)

func (s *Config) ClusterOperatorStatusAvailableCondition() (configv1.ConditionStatus, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	falseRC := configv1.ConditionFalse
	if !s.ConditionTrue(SamplesExist) {
		return falseRC, ""
	}
	versionToNote := s.Status.Version
	if len(versionToNote) == 0 {
		versionToNote = os.Getenv("RELEASE_VERSION")
	}
	return configv1.ConditionTrue, fmt.Sprintf(installed, versionToNote)
}
func (s *Config) ClusterOperatorStatusFailingCondition() (configv1.ConditionStatus, string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(s.Status.Conditions) < numconfigConditionType {
		return configv1.ConditionFalse, "", ""
	}
	trueRC := configv1.ConditionTrue
	if s.ConditionFalse(ConfigurationValid) {
		return trueRC, "invalid configuration", fmt.Sprintf(noInstallDetailed, os.Getenv("RELEASE_VERSION"), s.Condition(ConfigurationValid).Message)
	}
	if s.ClusterNeedsCreds() {
		return trueRC, "image pull credentials needed", fmt.Sprintf(noInstallDetailed, os.Getenv("RELEASE_VERSION"), s.Condition(ImportCredentialsExist).Message)
	}
	if s.AnyConditionUnknown() {
		return trueRC, "bad API object operation", s.ConditionsMessages()
	}
	return configv1.ConditionFalse, "", ""
}
func (s *Config) ClusterOperatorStatusProgressingCondition(failingState string, available configv1.ConditionStatus) (configv1.ConditionStatus, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(failingState) > 0 {
		return configv1.ConditionTrue, fmt.Sprintf(noInstallDetailed, os.Getenv("RELEASE_VERSION"), failingState)
	}
	if s.ConditionTrue(ImageChangesInProgress) {
		return configv1.ConditionTrue, fmt.Sprintf(moving, os.Getenv("RELEASE_VERSION"))
	}
	if available == configv1.ConditionTrue {
		return configv1.ConditionFalse, fmt.Sprintf(installed, s.Status.Version)
	}
	return configv1.ConditionFalse, ""
}
func (s *Config) ClusterNeedsCreds() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if strings.TrimSpace(s.Spec.SamplesRegistry) != "" && strings.TrimSpace(s.Spec.SamplesRegistry) != "registry.redhat.io" {
		return false
	}
	if s.Spec.ManagementState == operatorv1.Removed || s.Spec.ManagementState == operatorv1.Unmanaged {
		return false
	}
	if s.Status.Conditions == nil {
		return true
	}
	foundImportCred := false
	for _, rc := range s.Status.Conditions {
		if rc.Type == ImportCredentialsExist {
			foundImportCred = true
			break
		}
	}
	if !foundImportCred {
		return true
	}
	return s.ConditionFalse(ImportCredentialsExist)
}

type Event struct {
	Object	runtime.Object
	Deleted	bool
}
