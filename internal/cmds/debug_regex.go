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
	"regexp"
)

func regexMatching(configFile string) error {
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

	var texts []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		texts = append(texts, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(texts) == 0 || len(texts[0]) == 0 {
		return fmt.Errorf("invalid config, only supports 1 regex and multiple texts")
	}

	matcher, err := regexp.Compile(texts[0])
	if err != nil {
		return fmt.Errorf("unable to parse regex %s, err: %w", texts[0], err)
	}

	fmt.Printf("============= regex rule =============\n")
	fmt.Printf("%s\n", texts[0])

	fmt.Printf("\n========== matching results ==========\n")
	for _, text := range texts[1:] {
		if matcher.MatchString(text) {
			fmt.Printf("%4s:  %s\n", "Ok", text)
		} else {
			fmt.Printf("%4s:  %s\n", "Fail", text)
		}
	}

	return nil
}
