// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package couchbase

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func (ipt *Input) parseURL() error {
	u := ipt.Scheme + "://" + ipt.Host + ":" + fmt.Sprint(ipt.Port)
	if _, err := url.Parse(u); err != nil {
		return fmt.Errorf("parse url %s failed: %w", u, err)
	}

	if ipt.Port < 1 || ipt.Port > 65535 {
		return fmt.Errorf("parse port error: %d", ipt.Port)
	}

	if ipt.AdditionalPort < 1 || ipt.AdditionalPort > 65535 {
		return fmt.Errorf("parse additional port error: %d", ipt.AdditionalPort)
	}

	if !(ipt.Scheme == "http" || ipt.Scheme == "https") {
		return fmt.Errorf("parse additional scheme error: %s", ipt.Scheme)
	}

	return nil
}

func getPort(u string) (int, error) {
	uu, err := url.Parse(u)
	if err != nil {
		return -1, fmt.Errorf("parse url %s failed: %w", u, err)
	}

	h := strings.Split(uu.Host, ":")
	if len(h) != 2 {
		return -1, fmt.Errorf("get domain & port fail: %s", h)
	}

	port := h[1]
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return -1, fmt.Errorf("get port fail: %s", port)
	}
	if portInt < 1 || portInt > 65535 {
		return -1, fmt.Errorf("parse port error: %d", portInt)
	}

	return portInt, nil
}
