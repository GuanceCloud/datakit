// Package cli used to create k8s client and get some k8s info
package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Tag struct {
	ContainerName string

	Namespace string
	Pod       string
	Service   []string

	ReplicaSet string
	Deployment string

	StatefulSet string

	Job     string
	CronJob string

	Pid int
}

type K8sResource struct {
	Namespace   string
	Container   map[string]*ContainerInfo
	Pod         []*PodInfo
	Service     []*ServiceInfo
	Deployment  []*DeploymentInfo
	ReplicaSet  []*ReplicaSetInfo
	StatefulSet []*StatefulSetInfo
	Job         []*JobInfo
	CronJob     []*CronJobInfo
	DaemonSet   []*DaemonSetInfo
}

func GetRef(p *PodChain, refs []metav1.OwnerReference, res *K8sResource, dp int, lb []string, lbPrefix string) error {
	for _, ref := range refs {
		for {
			dp++
			if dp > 64 {
				return fmt.Errorf("exceeded the maximum number of traversals")
			}

			switch ref.Kind {
			case "ReplicaSet":
				for _, v := range res.ReplicaSet {
					if v.Name == ref.Name {
						if len(v.OwnerReferences) > 0 {
							if err := GetRef(p, v.OwnerReferences, res, dp, lb, lbPrefix); err == nil {
								return nil
							}
						}
						p.Rp = v
						p.Tag.Kind = Rp
						p.Tag.WorkloadName = v.Name
						p.Tag.Labels = getLBs(v.Labels, p.Tag.Labels, lb, lbPrefix)
						return nil
					}
				}
			case "Deployment":
				for _, v := range res.Deployment {
					if v.Name == ref.Name {
						if len(v.OwnerReferences) > 0 {
							if err := GetRef(p, v.OwnerReferences, res, dp, lb, lbPrefix); err == nil {
								return nil
							}
						}
						p.De = v
						p.Tag.Kind = De
						p.Tag.WorkloadName = v.Name
						p.Tag.Labels = getLBs(v.Labels, p.Tag.Labels, lb, lbPrefix)
						return nil
					}
				}
			case "Job":
				for _, v := range res.Job {
					if v.Name == ref.Name {
						if len(v.OwnerReferences) > 0 {
							if err := GetRef(p, v.OwnerReferences, res, dp, lb, lbPrefix); err == nil {
								return nil
							}
						}
						p.Jb = v
						p.Tag.Kind = Rp
						p.Tag.WorkloadName = v.Name
						p.Tag.Labels = getLBs(v.Labels, p.Tag.Labels, lb, lbPrefix)
						return nil
					}
				}
			case "CronJob":
				for _, v := range res.CronJob {
					if v.Name == ref.Name {
						if len(v.OwnerReferences) > 0 {
							if err := GetRef(p, v.OwnerReferences, res, dp, lb, lbPrefix); err == nil {
								return nil
							}
						}
						p.Cr = v
						p.Tag.Kind = Cr
						p.Tag.WorkloadName = v.Name
						p.Tag.Labels = getLBs(v.Labels, p.Tag.Labels, lb, lbPrefix)
						return nil
					}
				}
			case "StatefulSet":
				for _, v := range res.StatefulSet {
					if v.Name == ref.Name {
						if len(v.OwnerReferences) > 0 {
							if err := GetRef(p, v.OwnerReferences, res, dp, lb, lbPrefix); err == nil {
								return nil
							}
						}
						p.St = v
						p.Tag.Kind = St
						p.Tag.WorkloadName = v.Name
						p.Tag.Labels = getLBs(v.Labels, p.Tag.Labels, lb, lbPrefix)
						return nil
					}
				}
			case "DaemonSet":
				for _, v := range res.DaemonSet {
					if v.Name == ref.Name {
						if len(v.OwnerReferences) > 0 {
							if err := GetRef(p, v.OwnerReferences, res, dp, lb, lbPrefix); err == nil {
								return nil
							}
						}
						p.Ds = v
						p.Tag.Kind = Ds
						p.Tag.WorkloadName = v.Name
						p.Tag.Labels = getLBs(v.Labels, p.Tag.Labels, lb, lbPrefix)
						return nil
					}
				}
			default:
			}
		}
	}
	return fmt.Errorf("not found")
}

func GetAllResource(k8sCli *K8sClient, criCli []*CRIClient) (map[string]*K8sResource, error) {
	if k8sCli == nil {
		return nil, nil
	}

	infs, _ := GetContainersInfo(criCli)

	ctrWithNS := PodContainerMapping(infs)

	if err := k8sCli.ListAllPods(); err != nil {
		log.Warnf("list pods failed: %s", err.Error())
	}
	if err := k8sCli.ListAllServices(); err != nil {
		log.Warnf("list services failed: %s", err.Error())
	}
	if err := k8sCli.ListAllDeployments(); err != nil {
		log.Warnf("list deployments failed: %s", err.Error())
	}
	if err := k8sCli.ListAllStatefulSets(); err != nil {
		log.Warnf("list statefulsets failed: %s", err.Error())
	}
	if err := k8sCli.ListAllCronJobs(); err != nil {
		log.Warnf("list cronjobs failed: %s", err.Error())
	}
	if err := k8sCli.ListAllJobs(); err != nil {
		log.Warnf("list jobs failed: %s", err.Error())
	}
	if err := k8sCli.ListAllDaemonSets(); err != nil {
		log.Warnf("list daemonsets failed: %s", err.Error())
	}
	if err := k8sCli.ListAllReplicaSets(); err != nil {
		log.Warnf("list replicasets failed: %s", err.Error())
	}

	nsLi, err := k8sCli.ListNamespaces()
	if err != nil {
		return nil, fmt.Errorf("list namespaces failed: %w", err)
	}
	li := nsLi.Items

	rWithNS := map[string]*K8sResource{}
	for _, v := range li {
		namespace := v.GetName()
		r := &K8sResource{Namespace: namespace}

		r.Container = ctrWithNS[namespace]

		r.Pod, _ = GetPodInfo(k8sCli, namespace)
		r.Service, _ = GetServiceInfo(k8sCli, namespace)
		r.Deployment, _ = GetDeploymentInfo(k8sCli, namespace)
		r.StatefulSet, _ = GetStatefulSetInfo(k8sCli, namespace)
		r.CronJob, _ = GetCronJobInfo(k8sCli, namespace)
		r.Job, _ = GetJobInfo(k8sCli, namespace)
		r.DaemonSet, _ = GetDaemonSetInfo(k8sCli, namespace)
		r.ReplicaSet, _ = GetReplicaSetInfo(k8sCli, namespace)
		rWithNS[namespace] = r
	}

	return rWithNS, nil
}

type Kind int

const (
	UN Kind = iota
	Pd
	Cr
	Ds
	St
	De
	Rp
	Jb
)

func (k Kind) String() string {
	switch k {
	case UN:
		return "unknown"
	case De:
		return "deployment"
	case St:
		return "statefulset"
	case Ds:
		return "daemonset"
	case Cr:
		return "cronjob"
	case Rp:
		return "replicaset"
	case Jb:
		return "job"
	case Pd:
		return "pod"
	}
	return "unknown"
}

type K8sTag struct {
	Kind         Kind
	WorkloadName string
	SvcName      string
	PodName      string
	NS           string
	Labels       map[string]string
}

type PodChain struct {
	Po  *PodInfo
	Co  *ContainerInfo
	Cr  *CronJobInfo
	Jb  *JobInfo
	Ds  *DaemonSetInfo
	Svc []*ServiceInfo
	St  *StatefulSetInfo
	De  *DeploymentInfo
	Rp  *ReplicaSetInfo

	Tag *K8sTag
}

type Labels struct {
	kvs map[string]string
}

func (lb *Labels) Has(key string) bool {
	_, ok := lb.kvs[key]
	return ok
}

func (lb *Labels) Get(key string) string {
	v := lb.kvs[key]
	return v
}

type IPKey struct {
	Protocol string
	IP       string
	Port     int
}

type PodChainSvc struct {
	Chain *PodChain
	Svc   *ServiceInfo
}

type PodChainIPPort struct {
	Chain   *PodChain
	TCPPort map[int]struct{}
	UDPPort map[int]struct{}
	HostNet bool
}

type K8sMapping struct {
	ResChain map[string][]*PodChain

	Pid map[int]*PodChain

	IPSvc map[IPKey]*PodChainSvc
	IPPod map[IPKey]*PodChainIPPort
}

func (mp *K8sMapping) QueryPodName(pid int, ip string) string {
	if c, ok := mp.Pid[pid]; ok && c.Po != nil {
		return c.Po.Name
	}
	if c, ok := mp.IPPod[IPKey{IP: ip}]; ok && c.Chain != nil && c.Chain.Po != nil {
		return c.Chain.Po.Name
	}
	return ""
}

func (mp *K8sMapping) QueryPodInfo(pid int, ip string, port int, protocol string) (*K8sTag, bool) {
	if pid > 0 {
		if c, ok := mp.Pid[pid]; ok && c.Po != nil {
			return c.Tag, true
		}
	}
	if c, ok := mp.IPPod[IPKey{IP: ip}]; ok {
		if c.HostNet {
			switch protocol {
			case "udp", "UDP":
				if _, ok := c.UDPPort[port]; ok {
					return c.Chain.Tag, true
				}
			case "tcp", "TCP":
				if _, ok := c.TCPPort[port]; ok {
					return c.Chain.Tag, true
				}
			}
		} else {
			return c.Chain.Tag, true
		}
	}

	return nil, false
}

func (mp *K8sMapping) QuerySvcInfo(protocol, ip string, port int) (*PodChainSvc, bool) {
	if c, ok := mp.IPSvc[IPKey{
		IP:       ip,
		Port:     port,
		Protocol: protocol,
	}]; ok {
		return c, true
	}
	return nil, false
}

func (mp *K8sMapping) IsServer(pid int, protocol, ip string, port int) bool {
	if c, ok := mp.Pid[pid]; ok && c.Po != nil {
		for i := range c.Po.Ports {
			p := c.Po.Ports[i].ContainerPort
			if p != 0 && p == int32(port) {
				return true
			}
		}
	}

	if c, ok := mp.IPPod[IPKey{IP: ip}]; ok {
		switch protocol {
		case "udp":
			if _, ok := c.UDPPort[port]; ok {
				return true
			}
		case "tcp":
			if _, ok := c.TCPPort[port]; ok {
				return true
			}
		}
	}

	return false
}

func getLBs(src, dst map[string]string, lb []string, lbPrefix string) map[string]string {
	if dst == nil {
		dst = map[string]string{}
	}
	for _, k := range lb {
		if v, ok := src[k]; ok {
			dst[lbPrefix+k] = v
		}
	}
	return dst
}

func GetPidAndIPMapping(resWithNS map[string]*K8sResource, lb []string, lbPrefix string) *K8sMapping {
	mapping := K8sMapping{
		ResChain: map[string][]*PodChain{},
		Pid:      map[int]*PodChain{},
		IPSvc:    map[IPKey]*PodChainSvc{},
		IPPod:    map[IPKey]*PodChainIPPort{},
	}

	for ns, res := range resWithNS {
		li := []*PodChain{}
		for _, pod := range res.Pod {
			lbs := &Labels{kvs: pod.Labels}
			p := &PodChain{
				Po: pod,
				Tag: &K8sTag{
					NS:      ns,
					PodName: pod.Name,
					Labels:  getLBs(pod.Labels, nil, lb, lbPrefix),

					Kind:         Pd,
					WorkloadName: pod.Name,
				},
			}

			tcpp := map[int]struct{}{}
			udpp := map[int]struct{}{}

			ps := p.Po.Ports
			for i := range p.Po.Ports {
				switch ps[i].Protocol {
				case v1.ProtocolTCP:
					tcpp[int(ps[i].ContainerPort)] = struct{}{}
				case v1.ProtocolUDP:
					udpp[int(ps[i].ContainerPort)] = struct{}{}
				case v1.ProtocolSCTP:
					// pass
				}
			}
			for _, ip := range p.Po.PodIPs {
				mapping.IPPod[IPKey{
					IP: ip,
				}] = &PodChainIPPort{
					Chain:   p,
					TCPPort: tcpp,
					UDPPort: udpp,
					HostNet: p.Po.HostNetwork,
				}
			}

			for _, v := range res.Service {
				if v != nil && v.Selector != nil &&
					v.Selector.Matches(lbs) {
					p.Svc = append(p.Svc, v)

					for _, cip := range v.ClusterIPs {
						for i := 0; i < len(v.Port); i++ {
							mapping.IPSvc[IPKey{
								IP:       cip,
								Port:     int(v.Port[i].Port),
								Protocol: string(v.Port[i].Protocol),
							}] = &PodChainSvc{
								Svc:   v,
								Chain: p,
							}
						}
					}
				}
			}

			if len(p.Svc) > 0 {
				p.Tag.SvcName = p.Svc[0].Name
			}

			for _, v := range res.Container {
				if v.PodUID == pod.UID {
					p.Co = v
					mapping.Pid[v.Pid] = p
					break
				}
			}

			// K8s Workload
			if p.Po != nil {
				_ = GetRef(p, p.Po.OwnerReferences, res, 0, lb, lbPrefix)
			}

			li = append(li, p)
		}

		mapping.ResChain[ns] = li
	}

	return &mapping
}

type K8sInfo struct {
	criCli []*CRIClient
	k8sCli *K8sClient

	mapping *K8sMapping

	sync.RWMutex
}

func NewK8sInfo(k8scli *K8sClient, criLi []*CRIClient) *K8sInfo {
	k8sInf := &K8sInfo{
		criCli: criLi,
		k8sCli: k8scli,
	}

	_ = k8sInf.Update()

	return k8sInf
}

func (inf *K8sInfo) QueryPodName(pid int, ip string) string {
	inf.RLock()
	defer inf.RUnlock()
	if inf.mapping == nil {
		return ""
	}
	return inf.mapping.QueryPodName(pid, ip)
}

func (inf *K8sInfo) QueryPodInfo(pid int, ip string, port int, protocol string) (*K8sTag, bool) {
	inf.RLock()
	defer inf.RUnlock()
	if inf.mapping == nil {
		return nil, false
	}
	return inf.mapping.QueryPodInfo(pid, ip, port, protocol)
}

func (inf *K8sInfo) IsServer(pid int, protocol, ip string, port int) (bool, bool) {
	inf.RLock()
	defer inf.RUnlock()

	if inf.mapping == nil {
		return false, false
	}

	return inf.mapping.IsServer(pid, protocol, ip, pid), true
}

func (inf *K8sInfo) QuerySvcInfo(protocol, ip string, port int) (*PodChainSvc, bool) {
	inf.RLock()
	defer inf.RUnlock()
	if inf.mapping == nil {
		return nil, false
	}
	return inf.mapping.QuerySvcInfo(protocol, ip, port)
}

func (inf *K8sInfo) Update() error {
	v, err := GetAllResource(inf.k8sCli, inf.criCli)
	if err != nil {
		return err
	}
	inf.Lock()
	defer inf.Unlock()
	inf.mapping = GetPidAndIPMapping(v,
		inf.k8sCli.workloadLabels,
		inf.k8sCli.workloadLablePrefix)

	return nil
}

func (inf *K8sInfo) AutoUpdate(ctx context.Context, dur time.Duration) {
	ticker := time.NewTicker(dur)
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = inf.Update()
			case <-ctx.Done():
				return
			}
		}
	}()
}
