// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package build implement datakit build & release functions.
package build

type brand string

const (
	brandGuance    = brand("guance")
	brandTruewatch = brand("truewatch")
)

func (b brand) chartRepo() string {
	switch b {
	case brandGuance:
		return "guance.com/chartrepo/datakit"
	case brandTruewatch:
		return "truewatch.com/chartrepo/truewatch"
	default:
		return ValueNotSet
	}
}

// chartRepoName get chart repo name.
// we can add repo name locally:
//
//	$ helm repo add <chart-repo-name> https://some.repo.com/chartrepo/xxx
//
// If following chart repo URLs updated, we should update repo-add command in CI/CD machine.
func (b brand) chartRepoName(isTesting bool) string {
	switch b {
	case brandGuance:
		if isTesting {
			return "datakit-testing" // -> https://registry.jiagouyun.com/chartrepo/datakit
		} else {
			return "datakit-chart-cn" // -> https://pubrepo.guance.com/chartrepo/datakit
		}
	case brandTruewatch:
		if isTesting {
			return "datakit-intl-testing" // -> https://registry.jiagouyun.com/chartrepo/truewatch
		} else {
			return "datakit-chart-intl" // -> https://pubrepo.truewatch.com/chartrepo/truewatch
		}
	default:
		return ValueNotSet
	}
}

func (b brand) staticURL() string {
	switch b {
	case brandGuance:
		return "static.guance.com"
	case brandTruewatch:
		return "static.truewatch.com"
	default:
		return ValueNotSet
	}
}

func (b brand) domain() string {
	switch b {
	case brandGuance:
		return "guance.com"
	case brandTruewatch:
		return "truewatch.com"
	default:
		return ValueNotSet
	}
}

func (b brand) dockerImageRepo() string {
	switch b {
	case brandGuance:
		if DockerImageRepo != ValueNotSet {
			return DockerImageRepo
		} else {
			return "pubrepo.guance.com/datakit"
		}

	case brandTruewatch:
		if DockerImageRepo != ValueNotSet {
			return DockerImageRepo
		} else {
			return "pubrepo.truewatch.com/truewatch"
		}

	default:
		return ValueNotSet
	}
}

func (b brand) dcaDockerImageRepo() string {
	switch b {
	case brandGuance:
		if DockerImageRepo != ValueNotSet {
			return DockerImageRepo
		} else {
			return "pubrepo.guance.com/tools"
		}
	case brandTruewatch:
		if DockerImageRepo != ValueNotSet {
			return DockerImageRepo
		} else {
			return "pubrepo.truewatch.com/truewatch"
		}
	default:
		return ValueNotSet
	}
}
