package k8s

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type K8SHelper struct {
	ClientSet *kubernetes.Clientset
	PodStore  cache.Store
}

func InitK8sHelper() *K8SHelper {
	k8sHelper := &K8SHelper{}
	var err error
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Getwd fail, err: %v", err)
		return nil
	}
	kubeConfPath := fmt.Sprintf("%s/k8s/kube.conf", currentDir)
	k8sHelper.ClientSet, err = NewK8sClientSet(kubeConfPath)
	if err != nil {
		logger.Errorf("NewK8sClientSet fail, err: %v", err)
		return nil
	}
	stopChan := make(chan struct{}, 1)

	// 监听 pod
	go WatchPods(k8sHelper)

	// 监听 node
	nodeConf := NewNodeConfig(k8sHelper.ClientSet)
	nodeConf.RegisterEventHandler(NewNodeHandlerMock())
	nodeConf.Run(stopChan)

	return k8sHelper
}

func NewK8sClientSet(kubeConfPath string) (*kubernetes.Clientset, error) {

	var config *rest.Config
	var err error

	if kubeConfPath == "" {
		logger.Info("Using in cluster config")
		config, err = rest.InClusterConfig()
	} else {
		logger.Info("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfPath)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (kh *K8SHelper) GetClientSet() *kubernetes.Clientset {
	return kh.ClientSet
}

// 获取所有的 node 列表
func (kh *K8SHelper) GetNodes() (*v1.NodeList, error) {
	return kh.ClientSet.CoreV1().Nodes().List(meta_v1.ListOptions{})
}

// 获得所有 pod 列表
func (kh *K8SHelper) GetPods(namespace string) (*v1.PodList, error) {
	return kh.ClientSet.CoreV1().Pods(namespace).List(meta_v1.ListOptions{})
}

// 通过标签筛选 pod 列表
func (kh *K8SHelper) GetPodsBySelector(namespace string, selector map[string]string) (*v1.PodList, error) {
	return kh.ClientSet.CoreV1().Pods(namespace).List(
		meta_v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(selector).String(),
		})
}

// 获得指定的 pod
func (kh *K8SHelper) GetPod(namespace string, pod_name string) (*v1.Pod, error) {
	return kh.ClientSet.CoreV1().Pods(namespace).Get(pod_name, meta_v1.GetOptions{})
}
