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
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const cdnCacheTTL = time.Hour * 24 * 7 // 7d

var cdnCache = expirable.NewLRU[string, *cdnResolved](16384, nil, cdnCacheTTL)

type cdnResolved struct {
	domain  string
	cname   string
	cdnName string
}

func newCDNResolved(domain, cname, cdnName string) *cdnResolved {
	return &cdnResolved{
		domain:  domain,
		cname:   cname,
		cdnName: cdnName,
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
		node, ok := cdnCache.Get(resourceDomain)
		var (
			cname   string
			cdnName string
			err     error
		)
		if ok && node != nil {
			// cache is valid
			cname = node.cname
			cdnName = node.cdnName
		} else {
			// cache doesn't exist
			cname, cdnName, err = lookupCDNName(resourceDomain)
			if err != nil {
				log.Warnf("unable to lookup cdn name for domain [%s]: %s", resourceDomain, err)
			}
			cr := newCDNResolved(resourceDomain, cname, cdnName)
			cdnCache.Add(resourceDomain, cr)
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
