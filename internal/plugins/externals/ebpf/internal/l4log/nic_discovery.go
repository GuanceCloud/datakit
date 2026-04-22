//go:build linux
// +build linux

package l4log

import (
	"os"

	"github.com/vishvananda/netns"
	cruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
)

func ListContainersAndHostNetNS(ctrLi []cruntime.ContainerRuntime, allowLo bool,
) map[string]*netnsSnapshot {
	netnsInfo := map[string]*netnsSnapshot{}
	var curNetnsStr string
	curNetns, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		log.Errorf("get netns from pid: %w", err)
	} else {
		curNetnsStr = NSInode(curNetns)

		if _, ok := netnsInfo[curNetnsStr]; !ok {
			netnsInfo[curNetnsStr] = &netnsSnapshot{
				hostNS:      true,
				nsUID:       curNetnsStr,
				nns:         newNetNsHandle(true, allowLo, curNetns),
				contianerID: "",
			}
		} else {
			if err := curNetns.Close(); err != nil {
				log.Error(err)
			}
		}
	}

	for _, containerdCtr := range ctrLi {
		if containerdCtr == nil {
			continue
		}
		ctrs, err := containerdCtr.ListContainers()
		if err != nil {
			log.Errorf("get containers: %s", err.Error())
		}
		for _, c := range ctrs {
			nsH, err := netns.GetFromPid(c.Pid)
			if err != nil {
				log.Errorf("get netns from pid: %w", err)
				continue
			}
			nsHStr := NSInode(nsH)
			if nsHStr == curNetnsStr {
				if err := nsH.Close(); err != nil {
					log.Error(err)
				}
				continue
			}

			k8sTags := getK8sTags(c.Labels)
			if v, ok := netnsInfo[nsHStr]; !ok {
				nns := newNetNsHandle(false, allowLo, nsH)
				netnsInfo[nsHStr] = &netnsSnapshot{
					nsUID:       nsHStr,
					nns:         nns,
					contianerID: c.ID,
					pid:         map[int]struct{}{c.Pid: {}},
					tags:        k8sTags,
				}
			} else {
				v.pid[c.Pid] = struct{}{}
				if err := nsH.Close(); err != nil {
					log.Error(err)
				}
			}
		}
	}
	return netnsInfo
}

func getK8sTags(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	v := map[string]string{}
	v["k8s_namespace"] = labels["io.kubernetes.pod.namespace"]
	v["k8s_pod_name"] = labels["io.kubernetes.pod.name"]
	v["k8s_container_name"] = labels["io.kubernetes.container.name"]
	return v
}
