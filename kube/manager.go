package kube

import (
	"errors"
	"time"

	"github.com/ideahitme/cluster-autoscaler/logger"
	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	hotPodAnnotation  = "zalando.org/hot-pod"
	numWorkers        = 1
	workerPeriod      = time.Second
	retryLimit        = 5
	workerSleepPeriod = time.Second
)

var total uint32

// Manager responsible for all kubernetes related operations
type Manager struct {
	client   *kubernetes.Clientset
	queue    workqueue.RateLimitingInterface
	podStore cache.Indexer
}

// NewManager returns new Manager object
func NewManager(clientset *kubernetes.Clientset) *Manager {
	return &Manager{
		client: clientset,
	}
}

// NodeList returns list of nodes for the current cluster
func (km *Manager) NodeList() (*v1.NodeList, error) {
	return km.client.Nodes().List(meta_v1.ListOptions{})
}

// RunController runs controller watching the events
func (km *Manager) RunController(stopCh <-chan struct{}) {
	allPods, err := km.client.Pods(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
	if err == nil {
		logger.Log.Info("all pods: ", zap.Int("@num_pods", len(allPods.Items)))
	}
	source := cache.NewListWatchFromClient(
		km.client.Core().RESTClient(),
		"pods",
		api.NamespaceAll,
		fields.Everything(),
	)
	podStore, informer := cache.NewIndexerInformer(
		source,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    km.onPodCreate,
			UpdateFunc: km.onPodUpdate,
			DeleteFunc: km.onPodDelete,
		},
		cache.Indexers{})

	km.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	km.podStore = podStore
	go informer.Run(stopCh)

	for i := 0; i < numWorkers; i++ {
		go wait.Until(km.runWorker, workerPeriod, stopCh)
	}
}

func (km *Manager) runWorker() {
	key, quit := km.queue.Get()
	if quit {
		return
	}
	defer km.queue.Done(key)

	obj, exists, err := km.podStore.GetByKey(key.(string))
	if err != nil {
		if km.queue.NumRequeues(key) < retryLimit {
			km.queue.Add(key)
		} else {
			km.queue.Forget(key)
		}
	}
	if !exists {
		logger.Log.Info("pod does not exist anymore", zap.String("@pod", key.(string)))
		//delete ghost pods
	} else {
		logger.Log.Info("pod update/create", zap.String("@pod", obj.(*v1.Pod).GetName()))
		//run the routine
	}
	time.Sleep(workerSleepPeriod)
}

func (km *Manager) onPodCreate(obj interface{}) {
	newPod, ok := obj.(*v1.Pod)
	if !ok {
		logger.Log.Error("Pod type cast error", zap.Error(errors.New("not Pod type")))
		return
	}

	if km.skipPod(newPod) {
		return
	}

	logger.Log.Info("Pod created", zap.String("@Pod_name",
		newPod.Name))
	key, err := cache.MetaNamespaceKeyFunc(newPod)
	if err == nil {
		km.queue.Add(key)
	}
}

func (km *Manager) onPodUpdate(oldObj, newObj interface{}) {
	_, ok := oldObj.(*v1.Pod)
	if !ok {
		logger.Log.Error("Pod type cast error", zap.Error(errors.New("not Pod type")))
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		logger.Log.Error("Pod type cast error", zap.Error(errors.New("not Pod type")))
		return
	}

	if km.skipPod(newPod) {
		return
	}

	logger.Log.Info("Pod updated", zap.String("@Pod_name",
		newPod.Name))

	key, err := cache.MetaNamespaceKeyFunc(newPod)
	if err == nil {
		km.queue.Add(key)
	}
}

func (km *Manager) onPodDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		logger.Log.Error("Pod type cast error", zap.Error(errors.New("not Pod type")))
		return
	}
	logger.Log.Info("Pod deleted", zap.String("@Pod_name",
		pod.Name))
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pod)
	if err == nil {
		km.queue.Add(key)
	}
}

func (km *Manager) skipPod(pod *v1.Pod) bool {
	isJob, err := km.isCronJobPod(pod)
	if err != nil || isJob {
		logger.Log.Info("Pod type is a cron job. Skipping", zap.String("@pod_name", pod.Name))
		return true
	}
	isDaemonSet, err := isDaemonSetPod(pod)
	if err != nil || isDaemonSet {
		logger.Log.Info("Pod type is daemonset. Skipping", zap.String("@pod_name", pod.Name))
		return true
	}
	return false
}
