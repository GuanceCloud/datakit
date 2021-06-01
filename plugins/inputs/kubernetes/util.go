package kubernetes

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
)

func convertQuantity(s string, m float64) int64 {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		l.Debugf("D! [inputs.kube_inventory] failed to parse quantity: %s", err.Error())
		return 0
	}
	f, err := strconv.ParseFloat(fmt.Sprint(q.AsDec()), 64)
	if err != nil {
		l.Debugf("D! [inputs.kube_inventory] failed to parse float: %s", err.Error())
		return 0
	}
	if m < 1 {
		m = 1
	}
	return int64(f * m)
}

func atoi(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
