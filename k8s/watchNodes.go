package k8s

import (
	logger "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

// receive notifications about nodes object changes
type NodeHandler interface {
	OnNodeAdd(node *api_v1.Node)
	OnNodeUpdate(oldNode, node *api_v1.Node)
	OnNodeDelete(node *api_v1.Node)
	OnNodeSynced()
}

// invokes registered handlers on change.
type NodeConfig struct {
	controller    cache.Controller
	store         cache.Store
	eventHandlers []NodeHandler
}

func NewNodeConfig(clientSet *kubernetes.Clientset) *NodeConfig {
	logger.SetLevel(logger.InfoLevel)
	nodeListWatcher := cache.NewListWatchFromClient(
		clientSet.Core().RESTClient(),
		"nodes",
		api_v1.NamespaceAll,
		fields.Everything(),
	)

	result := &NodeConfig{}
	nodeEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    result.handleAddNode,
		UpdateFunc: result.handleUpdateNode,
		DeleteFunc: result.handleDeleteNode,
	}

	result.store, result.controller = cache.NewInformer(
		nodeListWatcher,
		&api_v1.Node{},
		time.Second*10,
		nodeEventHandler,
	)

	return result
}

func (nc *NodeConfig) RegisterEventHandler(handler NodeHandler) {
	nc.eventHandlers = append(nc.eventHandlers, handler)
}

func (nc *NodeConfig) Run(stopCh chan struct{}) {
	go nc.controller.Run(stopCh)
}

func (nc *NodeConfig) handleAddNode(obj interface{}) {
	node, ok := obj.(*api_v1.Node)
	if !ok {
		logger.Errorf("unexpected object type: %v", obj)
		return
	}

	for i := range nc.eventHandlers {
		nc.eventHandlers[i].OnNodeAdd(node)
	}
}

func (nc *NodeConfig) handleUpdateNode(oldObj, newObj interface{}) {
	oldNode, ok := oldObj.(*api_v1.Node)
	if !ok {
		logger.Errorf("unexpected object type: %v", oldObj)
		return
	}

	node, ok := newObj.(*api_v1.Node)
	if !ok {
		logger.Errorf("unexpected object type: %v", newObj)
		return
	}

	for i := range nc.eventHandlers {
		nc.eventHandlers[i].OnNodeUpdate(oldNode, node)
	}
}

func (nc *NodeConfig) handleDeleteNode(obj interface{}) {
	node, ok := obj.(*api_v1.Node)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			logger.Errorf("unexpected object type: %v", obj)
			return
		}
		if node, ok = tombstone.Obj.(*api_v1.Node); !ok {
			logger.Errorf("unexpected object type: %v", obj)
			return
		}
	}

	for i := range nc.eventHandlers {
		nc.eventHandlers[i].OnNodeDelete(node)
	}
}

type NodeHandlerMock struct {
	listenLabelKey string
}

func NewNodeHandlerMock() *NodeHandlerMock {
	return &NodeHandlerMock{
		listenLabelKey: "marwin",
	}
}

func (nhm *NodeHandlerMock) OnNodeAdd(node *api_v1.Node) {
	logger.Debugf("on node add > labels: %v", node.Labels)

	if _, ok := node.Labels[nhm.listenLabelKey]; ok {
		logger.Infof("send message: create node success.")
	}

}

func (nhm *NodeHandlerMock) OnNodeUpdate(oldNode, newNode *api_v1.Node) {
	logger.Debugf("on node update > old node labels: %v", oldNode.Labels)
	logger.Debugf("on node update > new node labels: %v", newNode.Labels)

	oldLabel, newLabel, oldOK, newOK := "", "", false, false
	if _, oldOK = oldNode.Labels[nhm.listenLabelKey]; oldOK {
		oldLabel = oldNode.Labels[nhm.listenLabelKey]
	}
	if _, newOK = newNode.Labels[nhm.listenLabelKey]; newOK {
		newLabel = newNode.Labels[nhm.listenLabelKey]
	}

	logger.Debugf("oldLabel: %s, newLabel: %s", oldLabel, newLabel)
	if (oldLabel != newLabel) || (oldOK != newOK) {
		logger.Infof("send message: update node label success. [%s] to [%s]", oldLabel, newLabel)
	}
}

func (nhm *NodeHandlerMock) OnNodeDelete(node *api_v1.Node) {
	return
}

func (nhm *NodeHandlerMock) OnNodeSynced() {
	return
}
