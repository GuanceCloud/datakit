package container

import (
	"fmt"
	"regexp"
)

type DepercatedLog struct {
	FilterMessage []string `toml:"filter_message"`
	Source        string   `toml:"source"`
	Service       string   `toml:"service"`
	Pipeline      string   `toml:"pipeline"`
}

const (
	logMatchByContainerName  = "container-name"
	logMatchByDeploymentName = "deployment-name"
)

type Logs []Log

type Log struct {
	MatchBy           string   `toml:"match_by"`
	Match             []string `toml:"match"`
	Source            string   `toml:"source"`
	Service           string   `toml:"service"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore_status"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`

	pattern []*regexp.Regexp
}

func (gs Logs) Init() error {
	for idx, g := range gs {
		if g.Source == "" {
			return fmt.Errorf("log[%d] source cannot be empty", idx)
		}
		switch g.MatchBy {
		case logMatchByContainerName, logMatchByDeploymentName:
			// nil
		default:
			return fmt.Errorf("invalind by %s, only accept %s and %s",
				g.MatchBy, logMatchByContainerName, logMatchByDeploymentName)
		}

		// regexp
		for _, match := range g.Match {
			pattern, err := regexp.Compile(match)
			if err != nil {
				return fmt.Errorf("config match index[%d], error: %s", idx, err)
			}
			gs[idx].pattern = append(gs[idx].pattern, pattern)
		}

		if g.Service == "" {
			g.Service = g.Source
		}
	}

	return nil
}

// Match 如果匹配成功则返回该项下标，否则返回 -1
func (gs Logs) Match(by, str string) (index int) {
	if str == "" {
		return -1
	}
	for idx, g := range gs {
		if by != g.MatchBy {
			continue
		}
		for _, pattern := range g.pattern {
			if pattern.MatchString(str) {
				return idx
			}
		}
	}
	return -1
}

func (gs Logs) MatchName(deploymentName, containerName string) (index int) {
	n := gs.Match(logMatchByDeploymentName, deploymentName)
	if n != -1 {
		return n
	}
	return gs.Match(logMatchByContainerName, containerName)
}
