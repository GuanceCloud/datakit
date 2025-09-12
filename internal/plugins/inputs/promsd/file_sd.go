// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/cespare/xxhash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type FileSD struct {
	Patterns        []string      `toml:"files"`
	RefreshInterval time.Duration `toml:"refresh_interval"`

	files  []string
	hashes []string

	tasks []scraper
	log   *logger.Logger
}

func (sd *FileSD) SetLogger(log *logger.Logger) { sd.log = log }

func (sd *FileSD) StartScraperProducer(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) {
	sd.log.Infof("file_sd: start %v", sd.Patterns)

	ticker := time.NewTicker(sd.RefreshInterval)
	defer ticker.Stop()

	for {
		if err := sd.produceScrapers(ctx, cfg, opts, out); err != nil {
			sd.log.Warnf("file_sd: failed of produce scrapers, err: %s", err)
		}

		select {
		case <-ctx.Done():
			sd.terminatedTasks()
			sd.log.Info("file_sd: terminating all tasks and exitting")
			return

		case <-ticker.C:
			// next
		}
	}
}

func (sd *FileSD) produceScrapers(ctx context.Context, cfg *ScrapeConfig, opts []promscrape.Option, out chan<- scraper) error {
	files, hashes, err := sd.scanFilesAndReadHashes()
	if err != nil {
		return err
	}

	if reflect.DeepEqual(sd.files, files) && reflect.DeepEqual(sd.hashes, hashes) {
		sd.log.Debugf("file_sd: files unchanged")
		return nil
	}

	targetGroups, err := readTargetGroups(files)
	if err != nil {
		return err
	}

	scrapers, err := convertTargetGroupsToScraper(cfg, opts, targetGroups)
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
	sd.files = files
	sd.hashes = hashes
	sd.tasks = scrapers
	sd.log.Infof("file_sd: found new targetGroups and replaced, len(%d) files", len(sd.files))
	return nil
}

func (sd *FileSD) terminatedTasks() {
	for _, task := range sd.tasks {
		task.markAsTerminated()
	}
}

func (sd *FileSD) scanFilesAndReadHashes() ([]string, []string, error) {
	scanner, err := fileprovider.NewScanner(sd.Patterns)
	if err != nil {
		return nil, nil, err
	}

	files, err := scanner.ScanFiles()
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(files)

	hashes, err := calculateFileHashes(files)
	if err != nil {
		return nil, nil, err
	}

	return files, hashes, nil
}

func readTargetGroups(files []string) ([]TargetGroup, error) {
	var res TargetGroups

	for _, path := range files {
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, err
		}

		var groups TargetGroups
		if err := json.Unmarshal(content, &groups); err != nil {
			return nil, err
		}

		res = append(res, groups...)
	}

	return res, nil
}

func calculateFileHashes(files []string) ([]string, error) {
	var res []string

	for _, path := range files {
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		defer file.Close() //nolint:errcheck,gosec

		hasher := xxhash.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return nil, err
		}
		res = append(res, hex.EncodeToString(hasher.Sum(nil)))
	}

	return res, nil
}
