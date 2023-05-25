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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
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

	paths, err := tailer.NewProvider().SearchFiles(globPaths).Result()
	if err != nil {
		return fmt.Errorf("unable to parse glob rules, err: %w", err)
	}

	fmt.Printf("============= glob paths ============\n")
	for _, path := range globPaths {
		fmt.Printf("%s\n", path)
	}
	fmt.Printf("\n========== found the files ==========\n")
	for _, path := range paths {
		fmt.Printf("%s\n", path)
	}

	return nil
}
