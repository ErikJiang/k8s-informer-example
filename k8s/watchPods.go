package k8s

import (
	logger "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"time"
)

func WatchPods(k8sHelper *K8SHelper) {
	watchList := cache.NewListWatchFromClient(k8sHelper.ClientSet.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	var podController cache.Controller
	k8sHelper.PodStore, podController = cache.NewInformer(
		watchList, &v1.Pod{}, time.Second*60,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handlePodAdd,
			UpdateFunc: handlePodUpdate,
		},
	)
	stop := make(chan struct{})
	go podController.Run(stop)
}

func handlePodAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	logger.Infof("Pod [%s] is add ...", pod.Name)
}

func handlePodUpdate(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	logger.Info("pod change ...")
	logger.Infof("old pod: %v", oldPod)
	logger.Infof("new pod: %v", newPod)
}
