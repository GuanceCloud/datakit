package kubernetes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const discoveryDir = "/usr/local/datakit/data/exporter_urls"

type PromPod struct {
	Name      string            `json:"pod,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Status    string            `json:"status,omitempty"`
	PodIp     string            `json:"podIp,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	NodeName  string            `json:"nodeName,omitempty"`
	Targets   []string          `json:"targets,omitempty"`
}

func (i *Input) collectPodsExporter() error {
	// mkdir DiscoveryDir
	l.Info("start discovery server")

	list, err := i.client.getPods()
	if err != nil {
		return err
	}

	if err := i.getPod(list); err != nil {
		return err
	}

	return nil
}

func (i *Input) getPod(p *corev1.PodList) error {
	var targetPods = make(map[string][]PromPod)
	for _, pod := range p.Items {
		pd := PromPod{
			Name:      pod.Name,
			Status:    fmt.Sprintf("%v", pod.Status.Phase),
			PodIp:     pod.Status.PodIP,
			NodeName:  pod.Spec.NodeName,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
		}

		for ankey, anvalue := range pod.Annotations {
			if strings.HasPrefix(ankey, "exporter_url.") {
				exporters := strings.Split(ankey, ".")
				if len(exporters) > 1 {
					promKey := exporters[1]
					promFile := fmt.Sprintf("%s.json", promKey)
					promFile = filepath.Join(discoveryDir, promFile)
					if anvalue == "off" {
						targetPods[promFile] = []PromPod{}
						continue
					}

					targets := parseExporter(anvalue, pd.PodIp)

					if len(targets) > 0 {
						pd.Targets = targets
					}

					if pds, ok := targetPods[promFile]; ok {
						pds = append(pds, pd)
						targetPods[promFile] = pds
					} else {
						targetPods[promFile] = []PromPod{pd}
					}
				}
			}
		}
	}

	if len(targetPods) > 0 {
		for promfile, pods := range targetPods {
			if len(pods) > 0 {
				m, err := json.MarshalIndent(pods, "", "  ")
				if err != nil {
					l.Errorf("marshal message err:%s", err.Error())
					return err
				}

				l.Info("create discovery file...")
				if err = ioutil.WriteFile(promfile, m, os.ModePerm); err != nil {
					l.Errorf("write to %s error %v", promfile, err)
					return err
				}
			} else {
				if datakit.FileExist(promfile) {
					if err := os.Remove(promfile); err != nil {
						l.Errorf("delete discovery file %s error %v", promfile, err)
						continue
					}
				}
			}
		}
	}

	return nil
}

func parseExporter(anvalue string, podIp string) []string {
	targets := make([]string, 0)

	targetArr := strings.Split(anvalue, ",")

	for _, target := range targetArr {
		target := strings.Replace(target, "$ip", podIp, -1)
		targets = append(targets, target)
	}
	// }

	return targets
}
