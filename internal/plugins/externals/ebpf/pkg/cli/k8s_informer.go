package cli

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
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

func NewInformaers(clientset *kubernetes.Clientset, stopCh <-chan struct{}) map[string]cache.SharedIndexInformer {
	// 为每种资源创建Informer
	informers := map[string]cache.SharedIndexInformer{
		ResourceTypePod:         createPodInformer(clientset, stopCh),
		ResourceTypeService:     createServiceInformer(clientset, stopCh),
		ResourceTypeDeployment:  createDeploymentInformer(clientset, stopCh),
		ResourceTypeStatefulSet: createStatefulSetInformer(clientset, stopCh),
		ResourceTypeDaemonSet:   createDaemonSetInformer(clientset, stopCh),
		ResourceTypeReplicaSet:  createReplicaSetInformer(clientset, stopCh),
		ResourceTypeCronJob:     createCronJobInformer(clientset, stopCh),
		ResourceTypeJob:         createJobInformer(clientset, stopCh),
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
		clientset,
		stopCh,
		ResourceTypePod,
	)
}

func createServiceInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.CoreV1().RESTClient(),
		"services",
		&corev1.Service{},
		clientset,
		stopCh,
		ResourceTypeService,
	)
}

func createDeploymentInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.AppsV1().RESTClient(),
		"deployments",
		&appsv1.Deployment{},
		clientset,
		stopCh,
		ResourceTypeDeployment,
	)
}

func createStatefulSetInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.AppsV1().RESTClient(),
		"statefulsets",
		&appsv1.StatefulSet{},
		clientset,
		stopCh,
		ResourceTypeStatefulSet,
	)
}

func createDaemonSetInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.AppsV1().RESTClient(),
		"daemonsets",
		&appsv1.DaemonSet{},
		clientset,
		stopCh,
		ResourceTypeDaemonSet,
	)
}

func createReplicaSetInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.AppsV1().RESTClient(),
		"replicasets",
		&appsv1.ReplicaSet{},
		clientset,
		stopCh,
		ResourceTypeReplicaSet,
	)
}

func createJobInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.BatchV1().RESTClient(),
		"jobs",
		&batchv1.Job{},
		clientset,
		stopCh,
		ResourceTypeJob,
	)
}

func createCronJobInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) cache.SharedIndexInformer {
	return createInformer(
		clientset.BatchV1().RESTClient(),
		"cronjobs",
		&batchv1.CronJob{},
		clientset,
		stopCh,
		ResourceTypeCronJob,
	)
}

func createInformer(
	restClient cache.Getter,
	resourceName string,
	objType apiRuntime.Object,
	clientset *kubernetes.Clientset,
	stopCh <-chan struct{},
	resourceType string,
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
		5*time.Minute, // 重同步间隔
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
