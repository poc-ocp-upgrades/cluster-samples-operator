package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	sampopclient "github.com/openshift/cluster-samples-operator/pkg/client"
	"github.com/openshift/cluster-samples-operator/test/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeset "k8s.io/client-go/kubernetes"
)

func TestMain(m *testing.M) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	kubeconfig, err := sampopclient.GetConfig()
	if err != nil {
		fmt.Printf("%#v", err)
		os.Exit(1)
	}
	kubeClient, err = kubeset.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("%#v", err)
		os.Exit(1)
	}
	err = waitForOperator()
	if err != nil {
		fmt.Println("failed waiting for operator to start")
		os.Exit(1)
	}
	opClient, err := configv1client.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("problem getting operator client %#v", err)
		os.Exit(1)
	}
	err = framework.DisableCVOForOperator(opClient)
	if err != nil {
		fmt.Printf("problem disabling operator deployment in CVO %#v", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
func waitForOperator() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	depClient := kubeClient.AppsV1().Deployments("openshift-cluster-samples-operator")
	err := wait.PollImmediate(1*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := depClient.Get("cluster-samples-operator", metav1.GetOptions{})
		if err != nil {
			fmt.Printf("error waiting for operator deployment to exist: %v\n", err)
			return false, nil
		}
		fmt.Println("found operator deployment")
		return true, nil
	})
	return err
}
