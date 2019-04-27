package stub

import (
	"fmt"
	"strings"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	kapis "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func importTag(stream *imagev1.ImageStream, tag string) (*imagev1.ImageStreamImport, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	finalTag, existing, multiple, err := followTagReferenceV1(stream, tag)
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.From == nil {
		return nil, nil
	}
	if existing.From != nil && existing.From.Kind != "DockerImage" {
		return nil, fmt.Errorf("tag %q points to existing %s %q, it cannot be re-imported", tag, existing.From.Kind, existing.From.Name)
	}
	if multiple {
		tag = finalTag
	}
	return newImageStreamImportTags(stream, map[string]string{tag: existing.From.Name}), nil
}
func followTagReferenceV1(stream *imagev1.ImageStream, tag string) (finalTag string, ref *imagev1.TagReference, multiple bool, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	seen := sets.NewString()
	for {
		if seen.Has(tag) {
			return tag, nil, multiple, fmt.Errorf("circular reference stream %s tag %s", stream.Name, tag)
		}
		seen.Insert(tag)
		var tagRef imagev1.TagReference
		for _, t := range stream.Spec.Tags {
			if t.Name == tag {
				tagRef = t
				break
			}
		}
		if tagRef.From == nil || tagRef.From.Kind != "ImageStreamTag" {
			return tag, &tagRef, multiple, nil
		}
		if strings.Contains(tagRef.From.Name, ":") {
			tagref := splitImageStreamTag(tagRef.From.Name)
			tag = tagref
		} else {
			tag = tagRef.From.Name
		}
		multiple = true
	}
}
func newImageStreamImportTags(stream *imagev1.ImageStream, tags map[string]string) *imagev1.ImageStreamImport {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	isi := newImageStreamImport(stream)
	for tag, from := range tags {
		var insecure, scheduled bool
		oldTagFound := false
		var oldTag imagev1.TagReference
		for _, t := range stream.Spec.Tags {
			if t.Name == tag {
				oldTag = t
				oldTagFound = true
				break
			}
		}
		if oldTagFound {
			insecure = oldTag.ImportPolicy.Insecure
			scheduled = oldTag.ImportPolicy.Scheduled
		}
		isi.Spec.Images = append(isi.Spec.Images, imagev1.ImageImportSpec{From: corev1.ObjectReference{Kind: "DockerImage", Name: from}, To: &corev1.LocalObjectReference{Name: tag}, ImportPolicy: imagev1.TagImportPolicy{Insecure: insecure, Scheduled: scheduled}, ReferencePolicy: getReferencePolicy()})
	}
	return isi
}
func newImageStreamImport(stream *imagev1.ImageStream) *imagev1.ImageStreamImport {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	isi := &imagev1.ImageStreamImport{ObjectMeta: kapis.ObjectMeta{Name: stream.Name, Namespace: stream.Namespace, ResourceVersion: stream.ResourceVersion}, Spec: imagev1.ImageStreamImportSpec{Import: true}}
	return isi
}
func splitImageStreamTag(nameAndTag string) (tag string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	parts := strings.SplitN(nameAndTag, ":", 2)
	if len(parts) > 1 {
		tag = parts[1]
	}
	if len(tag) == 0 {
		tag = "latest"
	}
	return tag
}
func getReferencePolicy() imagev1.TagReferencePolicy {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ref := imagev1.TagReferencePolicy{}
	ref.Type = imagev1.LocalTagReferencePolicy
	return ref
}
