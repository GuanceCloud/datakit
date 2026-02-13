package cli

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	apiRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	ResourceTypePod         = "pod"
	ResourceTypeService     = "service"
	ResourceTypeDeployment  = "deployment"
	ResourceTypeStatefulSet = "statefulset"
	ResourceTypeDaemonSet   = "daemonset"
	ResourceTypeReplicaSet  = "replicaset"
	ResourceTypeCronJob     = "cronjob"
	ResourceTypeJob         = "job"
)

type ResourceEvent struct {
	ResourceType string
	Key          string
}

func NewInformers(clientset *kubernetes.Clientset, podInformer bool, stopCh <-chan struct{}) map[string]cache.SharedIndexInformer {
	// 为每种资源创建Informer
	informers := map[string]cache.SharedIndexInformer{
		ResourceTypeService: createServiceInformer(clientset, stopCh),
	}
	if podInformer {
		informers[ResourceTypePod] = createPodInformer(clientset, stopCh)
	}

	// 等待所有缓存同步完成
	log.Infof("waiting for informer caches to sync ...")
	syncedFuncs := make([]cache.InformerSynced, 0, len(informers))
	for k := range informers {
		syncedFuncs = append(syncedFuncs, informers[k].HasSynced)
	}

	if !cache.WaitForCacheSync(stopCh, syncedFuncs...) {
		log.Errorf("failed to wait for caches to sync")
		return nil
	}
	log.Infof("all informer caches are synced, start processing events ...")

	return informers
}

func createPodInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.CoreV1().RESTClient(),
		"pods",
		&corev1.Pod{},
		stopCh,
	)
}

func createServiceInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.CoreV1().RESTClient(),
		"services",
		&corev1.Service{},
		stopCh,
	)
}

func createInformer(
	restClient cache.Getter,
	resourceName string,
	objType apiRuntime.Object,
	stopCh <-chan struct{},
) cache.SharedIndexInformer {
	lw := cache.NewListWatchFromClient(
		restClient,
		resourceName,
		corev1.NamespaceAll,
		fields.Everything(),
	)

	informer := cache.NewSharedIndexInformer(
		lw,
		objType,
		0, // 重同步间隔
		cache.Indexers{},
	)

	// 注册事件处理器
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})

	// 启动Informer
	go informer.Run(stopCh)
	return informer
}
