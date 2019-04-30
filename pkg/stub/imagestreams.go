package stub

import (
	"fmt"
	"strings"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	"github.com/openshift/cluster-samples-operator/pkg/cache"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kapis "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) processImageStreamWatchEvent(is *imagev1.ImageStream, deleted bool) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cfg, filePath, doUpsert, err := h.prepSamplesWatchEvent("imagestream", is.Name, is.Annotations, deleted)
	if cfg != nil && cfg.ConditionTrue(v1.ImageChangesInProgress) {
		logrus.Printf("Imagestream %s watch event do upsert %v; no errors in prep %v,  possibly update operator conditions %v", is.Name, doUpsert, err == nil, cfg != nil)
	} else {
		logrus.Debugf("Imagestream %s watch event do upsert %v; no errors in prep %v,  possibly update operator conditions %v", is.Name, doUpsert, err == nil, cfg != nil)
	}
	if !doUpsert {
		if err != nil {
			return err
		}
		if cfg == nil {
			return nil
		}
		anyChange := false
		if cache.UpsertsAmount() > 0 {
			cache.AddReceivedEventFromUpsert(is)
			if !cfg.ConditionTrue(v1.ImageChangesInProgress) || !cfg.ConditionTrue(v1.SamplesExist) {
				logrus.Printf("caching imagestream event %s because we have not yet completed all the samples upserts", is.Name)
				return nil
			}
			if !cache.AllUpsertEventsArrived() {
				logrus.Printf("caching imagestream event %s because we have not received all %d imagestream events", is.Name, cache.UpsertsAmount())
				return nil
			}
			streams := cache.GetUpsertImageStreams()
			keysToClear := []string{}
			for key, is := range streams {
				if is == nil {
					var e error
					is, e = h.imageclientwrapper.Get("openshift", key, kapis.GetOptions{})
					if e != nil {
						keysToClear = append(keysToClear, key)
						anyChange = true
						continue
					}
				}
				nonMatchDetail := ""
				cfg, nonMatchDetail, anyChange = h.processImportStatus(is, cfg)
				if !anyChange {
					logrus.Printf("imagestream %s still not finished with its image imports, including %s", is.Name, nonMatchDetail)
					return nil
				}
			}
			for _, key := range keysToClear {
				cache.RemoveUpsert(key)
				cfg.ClearNameInReason(cfg.Condition(v1.ImageChangesInProgress).Reason, key)
				cfg.ClearNameInReason(cfg.Condition(v1.ImportImageErrorsExist).Reason, key)
			}
			if anyChange {
				logrus.Printf("CRDUPDATE updating progress/error condition (within caching block) after results for %s", is.Name)
				err = h.crdwrapper.UpdateStatus(cfg)
				if err == nil {
					cache.ClearUpsertsCache()
				}
			}
			return err
		}
		cfg, _, anyChange = h.processImportStatus(is, cfg)
		if anyChange {
			logrus.Printf("CRDUPDATE updating progress/error condition after results for %s", is.Name)
			return h.crdwrapper.UpdateStatus(cfg)
		}
		return nil
	}
	if cfg == nil {
		return fmt.Errorf("cannot upsert imagestream %s because could not obtain Config", is.Name)
	}
	if cfg.ClusterNeedsCreds() {
		return fmt.Errorf("cannot upsert imagestream %s because rhel credentials do not exist", is.Name)
	}
	imagestream, err := h.Fileimagegetter.Get(filePath)
	if err != nil {
		h.processError(cfg, v1.SamplesExist, corev1.ConditionUnknown, err, "%v error reading file %s", filePath)
		logrus.Printf("CRDUPDATE event img update err bad fs read %s", filePath)
		h.crdwrapper.UpdateStatus(cfg)
		return nil
	}
	if deleted {
		is = nil
	}
	cache.AddUpsert(imagestream.Name)
	err = h.upsertImageStream(imagestream, is, cfg)
	if err != nil {
		cache.RemoveUpsert(imagestream.Name)
		if kerrors.IsAlreadyExists(err) {
			return nil
		}
		h.processError(cfg, v1.SamplesExist, corev1.ConditionUnknown, err, "%v error replacing imagestream %s", imagestream.Name)
		logrus.Printf("CRDUPDATE event img update err bad api obj update %s", imagestream.Name)
		h.crdwrapper.UpdateStatus(cfg)
		return err
	}
	progressing := cfg.Condition(v1.ImageChangesInProgress)
	now := kapis.Now()
	progressing.LastUpdateTime = now
	progressing.LastTransitionTime = now
	logrus.Debugf("Handle changing processing from false to true for imagestream %s", imagestream.Name)
	progressing.Status = corev1.ConditionTrue
	if !cfg.NameInReason(progressing.Reason, imagestream.Name) {
		progressing.Reason = progressing.Reason + imagestream.Name + " "
	}
	cfg.ConditionUpdate(progressing)
	logrus.Printf("CRDUPDATE progressing true update for imagestream %s", imagestream.Name)
	return h.crdwrapper.UpdateStatus(cfg)
}
func (h *Handler) upsertImageStream(imagestreamInOperatorImage, imagestreamInCluster *imagev1.ImageStream, opcfg *v1.Config) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	imagestreamInOperatorImage = jenkinsOverrides(imagestreamInOperatorImage)
	h.clearStreamFromImportError(imagestreamInOperatorImage.Name, opcfg.Condition(v1.ImportImageErrorsExist), opcfg)
	if _, isok := h.skippedImagestreams[imagestreamInOperatorImage.Name]; isok {
		if imagestreamInCluster != nil {
			if imagestreamInCluster.Labels == nil {
				imagestreamInCluster.Labels = make(map[string]string)
			}
			imagestreamInCluster.Labels[v1.SamplesManagedLabel] = "false"
			h.imageclientwrapper.Update("openshift", imagestreamInCluster)
		}
		return nil
	}
	h.updateDockerPullSpec([]string{"docker.io", "registry.redhat.io", "registry.access.redhat.com", "quay.io"}, imagestreamInOperatorImage, opcfg)
	if imagestreamInOperatorImage.Labels == nil {
		imagestreamInOperatorImage.Labels = make(map[string]string)
	}
	if imagestreamInOperatorImage.Annotations == nil {
		imagestreamInOperatorImage.Annotations = make(map[string]string)
	}
	imagestreamInOperatorImage.Labels[v1.SamplesManagedLabel] = "true"
	imagestreamInOperatorImage.Annotations[v1.SamplesVersionAnnotation] = h.version
	if imagestreamInCluster == nil {
		_, err := h.imageclientwrapper.Create("openshift", imagestreamInOperatorImage)
		if err != nil {
			if kerrors.IsAlreadyExists(err) {
				logrus.Printf("imagestream %s recreated since delete event", imagestreamInOperatorImage.Name)
				return err
			}
			return h.processError(opcfg, v1.SamplesExist, corev1.ConditionUnknown, err, "imagestream create error: %v")
		}
		logrus.Printf("created imagestream %s", imagestreamInOperatorImage.Name)
		return nil
	}
	imagestreamInOperatorImage.ResourceVersion = imagestreamInCluster.ResourceVersion
	_, err := h.imageclientwrapper.Update("openshift", imagestreamInOperatorImage)
	if err != nil {
		return h.processError(opcfg, v1.SamplesExist, corev1.ConditionUnknown, err, "imagestream update error: %v")
	}
	logrus.Printf("updated imagestream %s", imagestreamInCluster.Name)
	return nil
}
func (h *Handler) updateDockerPullSpec(oldies []string, imagestream *imagev1.ImageStream, opcfg *v1.Config) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch imagestream.Name {
	case "jenkins":
		return
	case "jenkins-agent-nodejs":
		return
	case "jenkins-agent-maven":
		return
	}
	if len(opcfg.Spec.SamplesRegistry) == 0 {
		return
	}
	logrus.Debugf("updateDockerPullSpec stream %s has repo %s", imagestream.Name, imagestream.Spec.DockerImageRepository)
	if len(imagestream.Spec.DockerImageRepository) > 0 && !strings.HasPrefix(imagestream.Spec.DockerImageRepository, opcfg.Spec.SamplesRegistry) {
		imagestream.Spec.DockerImageRepository = h.coreUpdateDockerPullSpec(imagestream.Spec.DockerImageRepository, opcfg.Spec.SamplesRegistry, oldies)
	}
	for _, tagref := range imagestream.Spec.Tags {
		logrus.Debugf("updateDockerPullSpec stream %s and tag %s has from %#v", imagestream.Name, tagref.Name, tagref.From)
		if tagref.From != nil {
			switch tagref.From.Kind {
			case "DockerImage":
				if !strings.HasPrefix(tagref.From.Name, opcfg.Spec.SamplesRegistry) {
					tagref.From.Name = h.coreUpdateDockerPullSpec(tagref.From.Name, opcfg.Spec.SamplesRegistry, oldies)
				}
			}
		}
	}
}
func (h *Handler) coreUpdateDockerPullSpec(oldreg, newreg string, oldies []string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	hasRegistry := false
	if strings.Count(oldreg, "/") == 2 {
		hasRegistry = true
	}
	logrus.Debugf("coreUpdatePull hasRegistry %v", hasRegistry)
	if hasRegistry {
		for _, old := range oldies {
			if strings.HasPrefix(oldreg, old) {
				oldreg = strings.Replace(oldreg, old, newreg, 1)
				logrus.Debugf("coreUpdatePull hasReg1 reg now %s", oldreg)
			} else {
				parts := strings.Split(oldreg, "/")
				oldreg = newreg + "/" + parts[1] + "/" + parts[2]
				logrus.Debugf("coreUpdatePull hasReg2 reg now %s", oldreg)
			}
		}
	} else {
		oldreg = newreg + "/" + oldreg
		logrus.Debugf("coreUpdatePull no hasReg reg now %s", oldreg)
	}
	return oldreg
}
func (h *Handler) clearStreamFromImportError(name string, importError *v1.ConfigCondition, cfg *v1.Config) *v1.ConfigCondition {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if cfg.NameInReason(importError.Reason, name) {
		logrus.Printf("clearing imagestream %s from the import image error condition", name)
	}
	start := strings.Index(importError.Message, "<imagestream/"+name+">")
	end := strings.LastIndex(importError.Message, "<imagestream/"+name+">")
	if start >= 0 && end > 0 {
		now := kapis.Now()
		importError.Reason = cfg.ClearNameInReason(importError.Reason, name)
		entireMsg := importError.Message[start : end+len("<imagestream/"+name+">")]
		importError.Message = strings.Replace(importError.Message, entireMsg, "", -1)
		if len(strings.TrimSpace(importError.Reason)) == 0 {
			importError.Status = corev1.ConditionFalse
			importError.Reason = ""
			importError.Message = ""
		} else {
			importError.Status = corev1.ConditionTrue
		}
		importError.LastTransitionTime = now
		importError.LastUpdateTime = now
		cfg.ConditionUpdate(importError)
	}
	return importError
}
func (h *Handler) processImportStatus(is *imagev1.ImageStream, cfg *v1.Config) (*v1.Config, string, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pending := false
	anyErrors := false
	importError := cfg.Condition(v1.ImportImageErrorsExist)
	nonMatchDetail := ""
	anyConditionUpdate := false
	_, skipped := h.skippedImagestreams[is.Name]
	if !skipped {
		logrus.Debugf("checking tag spec/status for %s spec len %d status len %d", is.Name, len(is.Spec.Tags), len(is.Status.Tags))
		for _, statusTag := range is.Status.Tags {
			if statusTag.Conditions != nil && len(statusTag.Conditions) > 0 {
				var latestGeneration int64
				var mostRecentErrorGeneration int64
				message := ""
				for _, condition := range statusTag.Conditions {
					if condition.Generation > latestGeneration {
						latestGeneration = condition.Generation
					}
					if condition.Status == corev1.ConditionFalse && len(condition.Message) > 0 {
						if condition.Generation > mostRecentErrorGeneration {
							mostRecentErrorGeneration = condition.Generation
							message = condition.Message
						}
					}
				}
				if mostRecentErrorGeneration > 0 && mostRecentErrorGeneration >= latestGeneration {
					logrus.Warningf("Image import for imagestream %s tag %s generation %v failed with detailed message %s", is.Name, statusTag.Tag, mostRecentErrorGeneration, message)
					anyErrors = true
					if !cfg.NameInReason(importError.Reason, is.Name) {
						now := kapis.Now()
						importError.Reason = importError.Reason + is.Name + " "
						importError.Message = importError.Message + "<imagestream/" + is.Name + ">" + message + "<imagestream/" + is.Name + ">"
						importError.Status = corev1.ConditionTrue
						importError.LastTransitionTime = now
						importError.LastUpdateTime = now
						cfg.ConditionUpdate(importError)
						anyConditionUpdate = true
						imgImport, err := importTag(is, statusTag.Tag)
						if err != nil {
							logrus.Warningf("attempted to define and imagestreamimport for imagestream/tag %s/%s but got err %v; simply moving on", is.Name, statusTag.Tag, err)
							break
						}
						if imgImport == nil {
							break
						}
						imgImport, err = h.imageclient.ImageStreamImports("openshift").Create(imgImport)
						if err != nil {
							logrus.Warningf("attempted to initiate an imagestreamimport retry for imagestream/tag %s/%s but got err %v; simply moving on", is.Name, statusTag.Tag, err)
							break
						}
						logrus.Printf("initiated an imagestreamimport retry for imagestream/tag %s/%s", is.Name, statusTag.Tag)
					}
					break
				}
			}
		}
		if !anyErrors {
			if len(is.Spec.Tags) == len(is.Status.Tags) {
				for _, specTag := range is.Spec.Tags {
					matched := false
					for _, statusTag := range is.Status.Tags {
						logrus.Debugf("checking spec tag %s against status tag %s with num items %d", specTag.Name, statusTag.Tag, len(statusTag.Items))
						if specTag.Name == statusTag.Tag {
							if statusTag.Items != nil {
								for _, event := range statusTag.Items {
									if specTag.Generation != nil {
										logrus.Debugf("checking status tag %d against spec tag %d", event.Generation, *specTag.Generation)
									}
									if specTag.Generation != nil && *specTag.Generation <= event.Generation {
										logrus.Debugf("got match")
										matched = true
										break
									}
									nonMatchDetail = fmt.Sprintf("spec tag %s is at generation %d, but status tag %s is at generation %d", specTag.Name, *specTag.Generation, statusTag.Tag, event.Generation)
								}
							}
						}
					}
					if !matched {
						pending = true
						break
					}
				}
			} else {
				pending = true
				nonMatchDetail = "the number of status tags does not equal the number of spec tags"
			}
		}
	} else {
		logrus.Debugf("no error/progress checks cause stream %s is skipped", is.Name)
		h.clearStreamFromImportError(is.Name, importError, cfg)
	}
	processing := cfg.Condition(v1.ImageChangesInProgress)
	logrus.Debugf("pending is %v any errors %v for %s", pending, anyErrors, is.Name)
	if !pending {
		if !anyErrors {
			h.clearStreamFromImportError(is.Name, importError, cfg)
		}
		if processing.Status == corev1.ConditionTrue {
			now := kapis.Now()
			logrus.Debugf("current reason %s ", processing.Reason)
			replaceOccurs := cfg.NameInReason(processing.Reason, is.Name)
			if replaceOccurs {
				processing.Reason = cfg.ClearNameInReason(processing.Reason, is.Name)
				logrus.Debugf("processing reason now %s", processing.Reason)
				if len(strings.TrimSpace(processing.Reason)) == 0 && processing.Status != corev1.ConditionFalse {
					logrus.Println("The last in progress imagestream has completed (import status check)")
					processing.Status = corev1.ConditionFalse
					processing.Reason = ""
				}
				processing.LastTransitionTime = now
				processing.LastUpdateTime = now
				cfg.ConditionUpdate(processing)
				logrus.Printf("There are no more image imports in flight for imagestream %s", is.Name)
				anyConditionUpdate = true
			}
		}
	}
	if cfg.NameInReason(importError.Reason, is.Name) && !anyErrors {
		logrus.Printf("There are no image import errors for %s", is.Name)
		h.clearStreamFromImportError(is.Name, importError, cfg)
	}
	return cfg, nonMatchDetail, anyConditionUpdate
}
