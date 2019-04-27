package stub

import (
	"os"
	"strings"
	"github.com/openshift/cluster-samples-operator/pkg/apis/samples/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func (h *Handler) processFiles(dir string, files []os.FileInfo, opcfg *v1.Config) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, file := range files {
		if file.IsDir() {
			logrus.Printf("processing subdir %s from dir %s", file.Name(), dir)
			subfiles, err := h.Filefinder.List(dir + "/" + file.Name())
			if err != nil {
				return h.processError(opcfg, v1.SamplesExist, corev1.ConditionUnknown, err, "error reading in content: %v")
			}
			err = h.processFiles(dir+"/"+file.Name(), subfiles, opcfg)
			if err != nil {
				return err
			}
			continue
		}
		logrus.Printf("processing file %s from dir %s", file.Name(), dir)
		if strings.HasSuffix(dir, "imagestreams") {
			path := dir + "/" + file.Name()
			imagestream, err := h.Fileimagegetter.Get(path)
			if err != nil {
				return h.processError(opcfg, v1.SamplesExist, corev1.ConditionUnknown, err, "%v error reading file %s", path)
			}
			h.imagestreamFile[imagestream.Name] = path
			continue
		}
		if strings.HasSuffix(dir, "templates") {
			template, err := h.Filetemplategetter.Get(dir + "/" + file.Name())
			if err != nil {
				return h.processError(opcfg, v1.SamplesExist, corev1.ConditionUnknown, err, "%v error reading file %s", dir+"/"+file.Name())
			}
			h.templateFile[template.Name] = dir + "/" + file.Name()
		}
	}
	return nil
}
func (h *Handler) GetBaseDir(arch string, opcfg *v1.Config) (dir string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch arch {
	case v1.X86Architecture:
		dir = x86OCPContentRootDir
	default:
	}
	return dir
}
