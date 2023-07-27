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

func (i *Input) parseURL() error {
	u := i.Scheme + "://" + i.Host + ":" + fmt.Sprint(i.Port)
	if _, err := url.Parse(u); err != nil {
		return fmt.Errorf("parse url %s failed: %w", u, err)
	}

	if i.Port < 1 || i.Port > 65535 {
		return fmt.Errorf("parse port error: %d", i.Port)
	}

	if i.AdditionalPort < 1 || i.AdditionalPort > 65535 {
		return fmt.Errorf("parse additional port error: %d", i.AdditionalPort)
	}

	if !(i.Scheme == "http" || i.Scheme == "https") {
		return fmt.Errorf("parse additional scheme error: %s", i.Scheme)
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
