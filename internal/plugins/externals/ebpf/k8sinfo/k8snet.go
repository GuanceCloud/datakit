package k8sinfo

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type K8sNetInfo struct {
	sync.RWMutex
	cli                *K8sClient
	svcNetInfo         map[string]*K8sServicesNet
	svcNetInfoNodePort map[Port]*K8sServicesNet
	poNetInfoIP        map[string]*K8sPodNet
	// Pods using the host network need to determine the port
	poNetInfoPort map[string]map[Port]*K8sPodNet

	autoUpdate bool
}

func (kinfo *K8sNetInfo) Update() error {
	// svc ip -> svc
	k8sSvcMap := map[string]*K8sServicesNet{}
	k8sSvcPortMap := map[Port]*K8sServicesNet{}

	// pod ip (not include host ip) -> pod
	k8sPodMap := map[string]*K8sPodNet{}
	// pod ip + port -> pod
	k8sPodNetPortMap := map[string]map[Port]*K8sPodNet{} // pod (include host_network)

	k8sPodTmpNetMap := map[string][]*K8sPodNet{}

	ns, err := kinfo.cli.GetNamespaces()
	if err != nil {
		return err
	}

	for _, ns := range ns {
		svcNet, err := kinfo.cli.GetServiceNet(ns)
		if err != nil {
			return err
		}

		endpointNet, err := kinfo.cli.GetEndpointNet(ns)
		if err != nil {
			return err
		}

		podNet, err := kinfo.cli.GetPodNet(ns)
		if err != nil {
			return err
		}

		deploymnt, err := kinfo.cli.GetDeployment(ns)
		if err != nil {
			return err
		}

		for ip, list := range podNet {
			for _, v := range list {
				for _, d := range deploymnt {
					if MatchLabel(d.MatchLabels, v.Labels) {
						v.DeploymentName = d.Name
						break
					}
				}
				if !v.HostNetwork {
					k8sPodMap[ip] = v
				}
				k8sPodTmpNetMap[ip] = append(k8sPodTmpNetMap[ip], v)
			}
		}

		// for range services
		for name, svc := range svcNet {
			for _, v := range svc.ClusterIPs {
				k8sSvcMap[v] = svc
			}

			for _, v := range svc.NodePort {
				k8sSvcPortMap[v] = svc
			}

			ep, ok := endpointNet[name]
			if !ok {
				continue
			}
			// svc' endpoint' ip port
			// Take the ip and port of the endpoint and match the pod through the label selector
			for ip, ports := range ep.IPPort {
				pods, ok := k8sPodTmpNetMap[ip]
				if ok {
					// for range pods
					for _, pod := range pods {
						if !MatchLabel(svc.Selector, pod.Labels) {
							continue
						}
						svc.DeploymentName = pod.DeploymentName
						if _, ok := k8sPodNetPortMap[ip]; !ok {
							k8sPodNetPortMap[ip] = map[Port]*K8sPodNet{}
						}
						for _, port := range ports {
							pod.ServiceName = svc.Name
							k8sPodNetPortMap[ip][port] = pod
						}
					}
				}
				pod, ok := k8sPodMap[ip]
				if ok {
					pod.ServiceName = svc.Name
				}
			}
		}
	}

	kinfo.Lock()
	defer kinfo.Unlock()

	kinfo.svcNetInfo = k8sSvcMap
	kinfo.svcNetInfoNodePort = k8sSvcPortMap

	kinfo.poNetInfoIP = k8sPodMap
	kinfo.poNetInfoPort = k8sPodNetPortMap

	return nil
}

func (kinfo *K8sNetInfo) AutoUpdate(ctx context.Context) {
	if kinfo.autoUpdate {
		return
	} else {
		kinfo.autoUpdate = true
	}

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-ticker.C:
				// TODO: log error
				_ = kinfo.Update()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// QueryPodInfo returns pod name, svc name, namespace, deployment,port, err).
func (kinfo *K8sNetInfo) QueryPodInfo(ip string, port uint32, protocol string) (
	string, string, string, string, error,
) {
	kinfo.RLock()
	defer kinfo.RUnlock()

	pP := Port{
		Port: port,
	}
	switch protocol {
	case "tcp":
		pP.Protocol = "TCP"
	case "udp":
		pP.Protocol = "UDP"
	}
	if p, ok := kinfo.poNetInfoPort[ip]; ok {
		// It may be a HostNetwork ip pod, which needs port assistance to determine
		if v, ok := p[pP]; ok {
			return v.Name, v.ServiceName, v.Namespace, v.DeploymentName, nil
		}
	}

	// The pod that sends the request as the client, without (host network ip)
	if v, ok := kinfo.poNetInfoIP[ip]; ok {
		return v.Name, v.ServiceName, v.Namespace, v.DeploymentName, nil
	}

	return "", "", "", "", fmt.Errorf("no match pod")
}

// QuerySvcInfo returns (svc name, namespace, error).
func (kinfo *K8sNetInfo) QuerySvcInfo(ip string, port uint32, protocol string) (string, string, string, error) {
	kinfo.RLock()
	defer kinfo.RUnlock()
	if v, ok := kinfo.svcNetInfo[ip]; ok {
		return v.Name, v.Namespace, v.DeploymentName, nil
	}

	pP := Port{
		Port: port,
	}
	switch protocol {
	case "tcp":
		pP.Protocol = "TCP"
	case "udp":
		pP.Protocol = "UDP"
	}

	if svc, ok := kinfo.svcNetInfoNodePort[pP]; ok {
		return svc.Name, svc.Namespace, svc.DeploymentName, nil
	}
	return "", "", "", fmt.Errorf("no match svc")
}

func NewK8sNetInfo(cli *K8sClient) (*K8sNetInfo, error) {
	kinfo := &K8sNetInfo{
		cli:        cli,
		autoUpdate: false,
	}

	if err := kinfo.Update(); err != nil {
		return nil, err
	}

	return kinfo, nil
}

func (kinfo *K8sNetInfo) IsServer(srcIP string, srcPort uint32, protocol string) bool {
	kinfo.RLock()
	defer kinfo.RUnlock()

	pP := Port{
		Port: srcPort,
	}
	switch protocol {
	case "tcp":
		pP.Protocol = "TCP"
	case "udp":
		pP.Protocol = "UDP"
	}
	// ip + port
	if p, ok := kinfo.poNetInfoPort[srcIP]; ok {
		if _, ok := p[pP]; ok {
			return true
		}
	}

	// kube-proxy(NodePort)
	if _, ok := kinfo.svcNetInfoNodePort[pP]; ok {
		return true
	}

	return false
}
