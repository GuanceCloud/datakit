// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const cdnCacheTTL = time.Hour * 24 * 7 // 7d

var cdnCache = newLruCDNCache(8192)

type cdnResolved struct {
	domain  string
	cname   string
	cdnName string
	created time.Time
}

func newCDNResolved(domain, cname, cdnName string, created time.Time) *cdnResolved {
	return &cdnResolved{
		domain:  domain,
		cname:   cname,
		cdnName: cdnName,
		created: created,
	}
}

func (ipt *Input) handleProvider(p *point.Point) (*point.Point, error) {
	var (
		providerType = "unknown"
		providerName = "unknown"
	)

	resourceDomain, ok := p.Get("resource_url_host").(string)
	if !ok {
		log.Warnf("invalid key resource_url_host(expect string) in point %s", p.Pretty())
		return nil, fmt.Errorf("invalid key resource_url_host")
	}

	if resourceDomain != "" && isDomainName(resourceDomain) {
		node := cdnCache.get(resourceDomain)
		var (
			cname   string
			cdnName string
			err     error
		)
		if node != nil {
			if node.Data.created.Add(cdnCacheTTL).Before(time.Now()) {
				// cache expired
				cname, cdnName, err = lookupCDNName(resourceDomain)
				if err != nil {
					log.Warnf("unable to lookup cdn name for domain [%s]: %s", resourceDomain, err)
				}
				node.Data.cname = cname
				node.Data.cdnName = cdnName
				node.Data.created = time.Now()
				cdnCache.moveToFront(node)
			} else {
				// cache is valid
				cname = node.Data.cname
				cdnName = node.Data.cdnName
				cdnCache.moveToFront(node)
			}
		} else {
			// cache not exists
			cname, cdnName, err = lookupCDNName(resourceDomain)
			if err != nil {
				log.Warnf("unable to lookup cdn name for domain [%s]: %s", resourceDomain, err)
			}
			cr := newCDNResolved(resourceDomain, cname, cdnName, time.Now())
			cdnCache.push(cr)
		}

		if cname != "" {
			if cname == resourceDomain {
				providerType = "first-party"
			} else {
				providerType = "CDN"
			}
		}
		if cdnName != "" {
			providerName = cdnName
		}
	}
	p.MustAdd("provider_type", providerType)
	p.MustAdd("provider_name", providerName)
	return p, nil
}

func lookupCDNNameForeach(cname string) (string, error) {
	for domain, cdn := range CDNList.literal {
		if strings.Contains(domain, strings.ToLower(cname)) {
			return cdn.Name, nil
		}
	}
	for pattern, cdn := range CDNList.glob {
		if (*pattern).Match(cname) {
			return cdn.Name, nil
		}
	}
	return "", fmt.Errorf("unable to resolve cdn name for domain: %s", cname)
}

func lookupCDNName(domain string) (string, string, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return "", "", fmt.Errorf("net.LookupCNAME(%s): %w", domain, err)
	}
	cname = strings.TrimRight(cname, ".")

	segments := strings.Split(cname, ".")

	// O(1)
	if len(segments) >= 2 {
		secondLevel := segments[len(segments)-2] + "." + segments[len(segments)-1]
		if cdn, ok := CDNList.literal[secondLevel]; ok {
			return cname, cdn.Name, nil
		}
	}

	// O(n)
	cdnName, err := lookupCDNNameForeach(cname)
	return cname, cdnName, err
}
