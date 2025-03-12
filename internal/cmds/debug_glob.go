// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
)

func globPath(configFile string) error {
	configPath := configFile
	if !path.IsAbs(configFile) {
		currentDir, _ := os.Getwd()
		configPath = filepath.Join(currentDir, configFile)
		_, err := os.Stat(configPath)
		if err != nil {
			return fmt.Errorf("not found config %s, err: %w", configFile, err)
		}
	}

	f, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return fmt.Errorf("unable to open file %s, err: %w", configPath, err)
	}
	defer func() {
		_ = f.Close()
	}()

	var globPaths []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		globPaths = append(globPaths, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	sc, err := fileprovider.NewScanner(globPaths)
	if err != nil {
		return fmt.Errorf("invalid patterns, err: %w", err)
	}
	paths, err := sc.ScanFiles()
	if err != nil {
		return fmt.Errorf("unable to parse glob rules, err: %w", err)
	}

	cp.Printf("============= glob paths ============\n")
	for _, path := range globPaths {
		cp.Printf("%s\n", path)
	}
	cp.Printf("\n========== found the files ==========\n")
	for _, path := range paths {
		cp.Printf("%s\n", path)
	}

	return nil
}
