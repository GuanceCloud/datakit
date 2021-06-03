package kubernetes

// import (
// 	"fmt"

// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/watch"
// 	// "k8s.io/client-go/pkg/api/v1"
// )

// func (i *Input) collectEvents() error {
// 	watchers, err := i.client.CoreV1().Services(i.client.namespace).Watch(metav1.ListOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	for event := range watchers.ResultChan() {
// 		svc := event.Object.(*corev1.Service)

// 		switch event.Type {
// 		case watch.Added:
// 			fmt.Printf("Service %s/%s added\n", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name)
// 		case watch.Modified:
// 			fmt.Printf("Service %s/%s modified\n", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name)
// 		case watch.Deleted:
// 			fmt.Printf("Service %s/%s deleted\n", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name)
// 		default:
// 			fmt.Println("service add =======>")
// 		}
// 	}

// 	return nil
// }
