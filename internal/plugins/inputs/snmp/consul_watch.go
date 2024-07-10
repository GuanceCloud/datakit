// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// ModuleRegexp module mapping.
// In prom, have a step regexp to mapping snmp_exporter module name.
//
// example: who will be collected by vpn2 module
//
// params:
//
//	module:[vpn2]
//
// relabel_configs:
//   - source_labels:[__meta_consul_service_metadata_brand]
//     regex:'cisco'
//   - source_labels:[__meta_consul_service_metadata_link]
//     regex:'TC'
//   - source_labels:[__meta_consul_service_metadata_deviceType]
//     regex:vpn|router
type ModuleRegexp struct {
	Module      string       `toml:"module"`       // module name in snmp_exporter.yml
	StepRegexps [][2]string  `toml:"step_regexps"` // [0]:field name, [1]:regexp
	StepRules   []ModuleRule `toml:"-"`
}

type ModuleRule struct {
	FieldName string         `toml:"-"`
	StepRule  *regexp.Regexp `toml:"-"`
}

type Watcher struct {
	Address       string                 // consul agent url："127.0.0.1:8500"
	Wp            *watch.Plan            // total plan
	watchers      map[string]*watch.Plan // sub service plan
	checkServices map[string]services    // mark if this service be update or deleted
	RWMutex       *sync.RWMutex
	Ipt           *Input
}

type service struct {
	ip          string // snmp instance ip
	modifyIndex uint64 // mark if this service be update
	idx         uint64 // mark if this service be deleted
}
type services map[string]*service

func (w *Watcher) isMyTask(ip string) bool {
	if len(w.Ipt.ExporterIPs) == 0 {
		return true
	}
	for _, v := range w.Ipt.ExporterIPs {
		if v == ip {
			return true
		}
	}
	return false
}

func (w *Watcher) doConsulJob(svc *consulapi.AgentService) {
	if svc.Port == 0 {
		l.Debugf("delete a device ip: %s", svc.ID)
		w.Ipt.userSpecificDevices.Delete(svc.ID)
		return
	}

	l.Debugf("try add or update a device id: %s", svc.ID)

	deviceIP, ok := svc.Meta[w.Ipt.InstanceIPKey]
	if !ok {
		l.Debugf("no InstanceIPKey in meta, id:%s", svc.ID)
		return
	}

	module, err := w.parseModule(svc)
	if err != nil {
		return
	}

	tags := make(map[string]string)
	for k, v := range svc.Meta {
		tags[k] = v
	}

	w.Ipt.jobs <- Job{
		ID:         USER_DISCOVERY,
		IP:         deviceIP,
		Idx:        -1,
		DeviceType: module,
		Tags:       tags,
	}
}

func (w *Watcher) parseModule(svc *consulapi.AgentService) (string, error) {
	var isMatch bool
	for _, moduleRegexp := range w.Ipt.ModuleRegexps {
		isMatch = true
		for _, stepRule := range moduleRegexp.StepRules {
			value, ok := svc.Meta[stepRule.FieldName]
			if !ok {
				isMatch = false
				break
			}

			if !stepRule.StepRule.MatchString(value) {
				isMatch = false
				break
			}
		}

		if isMatch {
			return moduleRegexp.Module, nil
		}
	}

	l.Debugf("can not parse module, id: %s", svc.ID)
	return "", fmt.Errorf("can not parse module, id: %s", svc.ID)
}

func (ipt *Input) checkConsulDiscovery() error {
	if ipt.ConsulDiscoveryURL == "" {
		return fmt.Errorf("not consul discover")
	}

	if ipt.InstanceIPKey == "" {
		l.Errorf("ipt.InstanceIPKey can not be nil")
		return fmt.Errorf("ipt.InstanceIPKey can not be nil")
	}
	// TODO check consul other config ......

	for i, m := range ipt.ModuleRegexps {
		for _, v := range m.StepRegexps {
			matcher, err := regexp.Compile(v[1])
			if err != nil {
				l.Errorf("unable to parse regex %v, err: %w", v, err)
				return fmt.Errorf("unable to parse regex %v, err: %w", v, err)
			}
			ipt.ModuleRegexps[i].StepRules = append(ipt.ModuleRegexps[i].StepRules, ModuleRule{v[0], matcher})
		}
	}

	return nil
}

// add a new consul sub service watch.
func (w *Watcher) registerSubWatcher(serviceName string, client *consulapi.Client) error {
	// watch, see also：https://www.consul.io/docs/dynamic-app-config/watches#service
	wp, err := watch.Parse(map[string]interface{}{
		"type":    "service",
		"service": serviceName,
		"token":   w.Ipt.ConsulToken,
	})
	if err != nil {
		return err
	}

	// check info
	serviceItems := services{}

	// callback handler
	wp.Handler = func(idx uint64, data interface{}) {
		switch serviceEntries := data.(type) {
		case []*consulapi.ServiceEntry:
			for _, se := range serviceEntries {
				if !w.isMyTask(se.Service.Address) {
					l.Debugf("not my task, id:%s", se.Service.ID)
					continue
				}

				serviceItem, ok := serviceItems[se.Service.ID]
				if ok && serviceItem.modifyIndex == se.Service.ModifyIndex {
					l.Debugf("not update, id:%s", se.Service.ID)
					continue
				}

				ip, ok := se.Service.Meta[w.Ipt.InstanceIPKey]
				if !ok {
					l.Debugf("no InstanceIPKey in meta, id:%s", se.Service.ID)
					continue
				}
				serviceItems[se.Service.ID] = &service{
					ip:          ip,
					modifyIndex: se.Service.ModifyIndex,
					idx:         idx,
				}

				l.Debugf("new or update a job, id:%s", se.Service.ID)
				w.doConsulJob(se.Service)
			}

			// Search delete service
			for id, svc := range serviceItems {
				if idx == svc.idx {
					continue
				}

				l.Debugf("delete a job, id:%s", id)
				fmt.Printf("delete a job, id:%s", id)
				w.doConsulJob(&consulapi.AgentService{
					// here is ip, not id, because in ipt.userSpecificDevices key is ip.
					ID: svc.ip,
				})
			}
		default:
			l.Debugf("unknown type")
		}
	}

	g.Go(func(ctx context.Context) error {
		defer wp.Stop()

		g.Go(func(ctx context.Context) error {
			if err = wp.RunWithClientAndHclog(client, nil); err != nil {
				l.Errorf("sub watch %s Run fail: %w", serviceName, err)
				return fmt.Errorf("sub watch %s Run fail: %w", serviceName, err)
			}
			return nil
		})

		// block here
		l.Infof("consul sub watch %s blocking", serviceName)
		for {
			select {
			case <-datakit.Exit.Wait():
				l.Infof("consul sub watch %s exit", serviceName)
				return nil
			case <-w.Ipt.semStop.Wait():
				l.Infof("consul sub watch %s return", serviceName)
				return nil
			}
		}
	})

	w.RWMutex.Lock()
	w.watchers[serviceName] = wp
	w.checkServices[serviceName] = serviceItems
	w.RWMutex.Unlock()

	return nil
}

func newTotalWatcher(opts map[string]interface{}, consulAddr string, client *consulapi.Client) (*Watcher, error) {
	wp, err := watch.Parse(opts)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Address:       consulAddr,
		Wp:            wp,
		watchers:      make(map[string]*watch.Plan),
		checkServices: make(map[string]services),
		RWMutex:       new(sync.RWMutex),
	}

	wp.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		// total services watch, see also: https://github.com/dmcsorley/avast/blob/master/consul.go
		case map[string][]string:
			for i := range d {
				if _, ok := w.watchers[i]; ok || i == "consul" {
					continue
				}
				if err := w.registerSubWatcher(i, client); err != nil {
					l.Errorf("register service watcher fail, err:%w", err)
				}
			}

			// remove unknown services from watchers
			w.RWMutex.RLock()
			watches := w.watchers
			w.RWMutex.RUnlock()
			for i, svc := range watches {
				if _, ok := d[i]; !ok {
					svc.Stop()
					delete(watches, i)
				}
			}
		default:
			l.Debugf("unknown consul type : %v", &d)
		}
	}

	return w, nil
}

func (ipt *Input) consulDiscovery() {
	if ipt.checkConsulDiscovery() != nil {
		return
	}

	config := &consulapi.Config{
		Address: ipt.ConsulDiscoveryURL,
		Token:   ipt.ConsulToken,
	}

	if ipt.TLSClientConfig != nil {
		caCert := ""
		if len(ipt.TLSClientConfig.CaCerts) > 0 {
			caCert = ipt.TLSClientConfig.CaCerts[0]
		}
		config.TLSConfig = consulapi.TLSConfig{
			CAFile:             caCert,
			CertFile:           ipt.TLSClientConfig.Cert,
			KeyFile:            ipt.TLSClientConfig.CertKey,
			InsecureSkipVerify: ipt.TLSClientConfig.InsecureSkipVerify,
		}
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		l.Errorf("create consul client fail, err: %w", err)
		return
	}

	_, err = client.Agent().Services()
	if err != nil {
		l.Errorf("connect consul server fail, err: %w", err)
		return
	}

	g.Go(func(ctx context.Context) error {
		opts := map[string]interface{}{
			"type": "services",
		}
		if ipt.ConsulToken != "" {
			opts["token"] = ipt.ConsulToken
		}

		w, err := newTotalWatcher(opts, ipt.ConsulDiscoveryURL, client)
		if err != nil {
			l.Errorf("NewWatcher fail: %w", err)
			return fmt.Errorf("NewWatcher fail: %w", err)
		}
		w.Ipt = ipt

		defer w.Wp.Stop()

		g.Go(func(ctx context.Context) error {
			if err = w.Wp.RunWithClientAndHclog(client, nil); err != nil {
				l.Errorf("total watch %s Run fail: %w", ipt.ConsulDiscoveryURL, err)
				return fmt.Errorf("total watch %s Run fail: %w", ipt.ConsulDiscoveryURL, err)
			}
			return nil
		})

		// block here
		l.Infof("consul watch %s blocking", ipt.ConsulDiscoveryURL)
		for {
			select {
			case <-datakit.Exit.Wait():
				l.Infof("consul watch %s exit", ipt.ConsulDiscoveryURL)
				return nil
			case <-ipt.semStop.Wait():
				l.Infof("consul watch %s return", ipt.ConsulDiscoveryURL)
				return nil
			}
		}
	})
}
