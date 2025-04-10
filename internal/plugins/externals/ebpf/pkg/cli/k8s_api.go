// Package cli used to create k8s client and get some k8s info
package cli

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type K8sClient struct {
	*kubernetes.Clientset

	workloadLabels      []string
	workloadLablePrefix string
	timeout             time.Duration
}

func newTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

func (cli *K8sClient) ListNamespaces() (*corev1.NamespaceList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListEndpoints(ns string) (*corev1.EndpointsList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.CoreV1().Endpoints(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListPods(ns string) (*corev1.PodList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()

	return cli.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListServices(ns string) (*corev1.ServiceList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListReplicaSets(ns string) (*appsv1.ReplicaSetList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.AppsV1().ReplicaSets(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListDeployments(ns string) (*appsv1.DeploymentList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListStatefulSets(ns string) (*appsv1.StatefulSetList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListDaemonSet(ns string) (*appsv1.DaemonSetList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.AppsV1().DaemonSets(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListCronJobs(ns string) (*batchv1.CronJobList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.BatchV1().CronJobs(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListJobs(ns string) (*batchv1.JobList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListNodes() (*corev1.NodeList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}
