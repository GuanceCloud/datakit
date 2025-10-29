// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/hashicorp/consul/api"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type ConsulSD struct {
	Server          string        `toml:"server"`
	PathPrefix      string        `toml:"path_prefix"`
	Token           string        `toml:"token"`
	Datacenter      string        `toml:"datacenter"`
	Namespace       string        `toml:"namespace"`
	Partition       string        `toml:"partition"`
	Scheme          string        `toml:"scheme"`
	Services        []string      `toml:"services"`
	Filter          string        `toml:"filter"`
	AllowStale      bool          `toml:"allow_stale"`
	RefreshInterval time.Duration `toml:"refresh_interval"`
	Auth            *Auth         `toml:"auth"`

	targets []consulSDTarget
	tasks   []scraper

	clientConfig *api.Config
	queryOptions *api.QueryOptions
	logger       *logger.Logger
}

func (sd *ConsulSD) SetLogger(logger *logger.Logger) { sd.logger = logger }

func (sd *ConsulSD) StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) {
	sd.logger.Infof("consul_sd: starting service discovery for %s", sd.Server)
	sd.setup()

	ticker := time.NewTicker(sd.RefreshInterval)
	defer ticker.Stop()

	for {
		if err := sd.produceScrapers(ctx, cfg, opts, out); err != nil {
			sd.logger.Warnf("consul_sd: failed to produce scrapers: %s", err)
		}

		select {
		case <-ctx.Done():
			sd.terminatedTasks()
			sd.logger.Info("consul_sd: terminating all tasks and exiting")
			return

		case <-ticker.C:
			// next
		}
	}
}

func (sd *ConsulSD) produceScrapers(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) error {
	newTargets, err := sd.discoveryTargets()
	if err != nil {
		return err
	}

	if !sd.targetsChanged(newTargets) {
		sd.logger.Debugf("consul_sd: targets unchanged")
		return nil
	}

	scrapers, err := sd.convertTargetsToScraper(cfg, opts, newTargets)
	if err != nil {
		return err
	}

	for _, scraper := range scrapers {
		if ctx.Err() != nil {
			return err
		}

		select {
		case out <- scraper:
			// next
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	sd.terminatedTasks()
	sd.targets = newTargets
	sd.tasks = scrapers
	sd.logger.Infof("consul_sd: updated targets, found %d new scrapers", len(scrapers))
	return nil
}

func (sd *ConsulSD) convertTargetsToScraper(cfg *ScrapeConfig, opts []promscrape.Option, newTargets []consulSDTarget) ([]scraper, error) {
	var scrapers []scraper

	for _, target := range newTargets {
		u := &url.URL{
			Scheme: cfg.Scheme,
			Host:   target.Address,
			Path:   cfg.MetricsPath,
		}

		paramValues, err := url.ParseQuery(cfg.Params)
		if err != nil {
			sd.logger.Warnf("consul_sd: unexpected scrape params: %s", cfg.Params)
		} else {
			u.RawQuery = paramValues.Encode()
		}

		scraper, err := newPromScraper(u.String(), opts)
		if err != nil {
			return nil, err
		}
		scrapers = append(scrapers, scraper)
	}
	return scrapers, nil
}

func (sd *ConsulSD) discoveryTargets() ([]consulSDTarget, error) {
	client, err := api.NewClient(sd.clientConfig)
	if err != nil {
		return nil, fmt.Errorf("consul client error: %w", err)
	}

	services, err := sd.getServices(client)
	if err != nil {
		return nil, fmt.Errorf("fetch services error: %w", err)
	}

	var targets []consulSDTarget
	for _, service := range services {
		res, err := sd.getServiceTargets(client, service)
		if err != nil {
			sd.logger.Warnf("consul_sd: get service[%s] error: %s", service, err)
			continue
		}
		targets = append(targets, res...)
	}
	return targets, nil
}

func (sd *ConsulSD) getServices(client *api.Client) ([]string, error) {
	if len(sd.Services) > 0 {
		return sd.Services, nil
	}

	services, _, err := client.Catalog().Services(sd.queryOptions)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(services))
	for name := range services {
		res = append(res, name)
	}
	return res, nil
}

func (sd *ConsulSD) getServiceTargets(client *api.Client, service string) ([]consulSDTarget, error) {
	entries, _, err := client.Catalog().Service(service, "", sd.queryOptions)
	if err != nil {
		return nil, err
	}

	var targets []consulSDTarget
	for _, entry := range entries {
		address := fmt.Sprintf("%s:%d", entry.ServiceAddress, entry.ServicePort)
		if entry.ServiceAddress == "" {
			address = fmt.Sprintf("%s:%d", entry.Address, entry.ServicePort)
		}
		targets = append(targets, consulSDTarget{Address: address})
	}
	return targets, nil
}

func (sd *ConsulSD) setup() {
	sd.clientConfig = api.DefaultConfig()
	sd.clientConfig.Address = sd.Server
	sd.clientConfig.PathPrefix = sd.PathPrefix
	sd.clientConfig.Token = sd.Token
	sd.clientConfig.Datacenter = sd.Datacenter
	sd.clientConfig.Namespace = sd.Namespace
	sd.clientConfig.Partition = sd.Partition

	if sd.Scheme == "https" {
		sd.clientConfig.Scheme = sd.Scheme
	}

	if sd.Auth != nil && sd.Auth.TLSClientConfig != nil && len(sd.Auth.CaCerts) > 0 {
		sd.clientConfig.TLSConfig = api.TLSConfig{
			CAFile:             sd.Auth.CaCerts[0],
			CertFile:           sd.Auth.Cert,
			KeyFile:            sd.Auth.CertKey,
			InsecureSkipVerify: sd.Auth.InsecureSkipVerify,
		}
	}

	sd.queryOptions = &api.QueryOptions{
		AllowStale: sd.AllowStale,
		Filter:     sd.Filter,
	}
}

func (sd *ConsulSD) targetsChanged(newTargets []consulSDTarget) bool {
	return !reflect.DeepEqual(sd.targets, newTargets)
}

func (sd *ConsulSD) terminatedTasks() {
	for _, task := range sd.tasks {
		task.markAsTerminated()
	}
}

type consulSDTarget struct {
	Address string
}
