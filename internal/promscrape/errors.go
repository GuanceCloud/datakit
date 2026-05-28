// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"errors"
	"fmt"
	"net/http"
)

type ScrapeError struct {
	URL        string
	StatusCode int
	Err        error
}

func (e *ScrapeError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("unexpected status code returned when scraping %q: %d", e.URL, e.StatusCode)
	}
	if e.Err != nil {
		return fmt.Sprintf("failed to scrape %q: %s", e.URL, e.Err)
	}
	return fmt.Sprintf("failed to scrape %q", e.URL)
}

func (e *ScrapeError) Unwrap() error {
	return e.Err
}

func (e *ScrapeError) Retryable() bool {
	if e.StatusCode == http.StatusUnauthorized || e.StatusCode >= http.StatusInternalServerError {
		return true
	}
	return e.StatusCode == 0 && e.Err != nil
}

func IsRetryableScrapeError(err error) bool {
	var scrapeErr *ScrapeError
	if errors.As(err, &scrapeErr) {
		return scrapeErr.Retryable()
	}
	return false
}
