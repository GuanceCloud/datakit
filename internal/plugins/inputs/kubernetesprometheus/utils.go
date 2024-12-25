// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

func isTrue(s string) bool {
	return s == "true"
}

func isKeywords(s string) bool {
	return strings.HasPrefix(s, "__kubernetes")
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func buildURLWithParams(scheme, address, port, path, params string) (*url.URL, error) {
	u := &url.URL{
		Scheme: scheme,
		Host:   address + ":" + port,
		Path:   path,
	}
	if params != "" {
		query, err := url.ParseQuery(params)
		if err != nil {
			return nil, err
		} else {
			u.RawQuery = query.Encode()
		}
	}

	if _, err := url.Parse(u.String()); err != nil {
		return nil, err
	}
	return u, nil
}

func splitHost(remote string) string {
	host := remote

	// try get 'host' tag from remote URL.
	if u, err := url.Parse(remote); err == nil && u.Host != "" { // like scheme://host:[port]/...
		host = u.Host
		if ip, _, err := net.SplitHostPort(u.Host); err == nil {
			host = ip
		}
	} else { // not URL, only IP:Port
		if ip, _, err := net.SplitHostPort(remote); err == nil {
			host = ip
		}
	}

	if host == "localhost" || net.ParseIP(host).IsLoopback() {
		return ""
	}

	return host
}

func getLocalNodeName() (string, error) {
	var e string
	if os.Getenv("NODE_NAME") != "" {
		e = os.Getenv("NODE_NAME")
	}
	if os.Getenv("ENV_K8S_NODE_NAME") != "" {
		e = os.Getenv("ENV_K8S_NODE_NAME")
	}
	if e != "" {
		return e, nil
	}
	return "", fmt.Errorf("invalid ENV_K8S_NODE_NAME environment, cannot be empty")
}
