// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"path"
	"strings"
	T "testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestOperatorDocsAreOwnedByOperator(t *T.T) {
	for _, lang := range []string{"zh", "en"} {
		docs, err := AllDocs.ReadDir(path.Join("doc", lang))
		assert.NoError(t, err)

		for _, doc := range docs {
			name := doc.Name()
			isOperatorDoc := name == "datakit-operator.md" ||
				(strings.HasPrefix(name, "operator-") && strings.HasSuffix(name, ".md"))
			assert.False(t, isOperatorDoc, "%s/%s is maintained by datakit-operator", lang, name)
		}
	}
}

func TestOperatorDocsNavigationRemainsInDatakit(t *T.T) {
	operatorDocs := []string{
		"datakit-operator.md",
		"operator-changelog.md",
		"operator-asyncprofile.md",
		"operator-ddtrace.md",
		"operator-flameshot.md",
		"operator-logfwd.md",
		"operator-logging.md",
		"operator-otel.md",
		"operator-pyspy.md",
	}

	operatorSections := map[string]string{
		"zh": "DataKit Operator",
		"en": "Operator Configuration",
	}
	for lang, operatorSection := range operatorSections {
		data, err := AllDocs.ReadFile(path.Join("doc", lang, "datakit.pages"))
		if !assert.NoError(t, err) {
			continue
		}

		var pages struct {
			Nav []any `yaml:"nav"`
		}
		if !assert.NoError(t, yaml.Unmarshal(data, &pages)) {
			continue
		}

		operatorNav, ok := findNavigationSection(pages.Nav, operatorSection)
		if !assert.True(t, ok, "%s navigation must contain %q", lang, operatorSection) {
			continue
		}
		assert.ElementsMatch(t, operatorDocs, navigationTargets(operatorNav),
			"%s navigation must contain exactly the Operator-owned pages", lang)
	}
}

func findNavigationSection(nodes []any, title string) ([]any, bool) {
	for _, node := range nodes {
		item, ok := node.(map[any]any)
		if !ok {
			continue
		}

		for label, value := range item {
			children, ok := value.([]any)
			if !ok {
				continue
			}
			if label == title {
				return children, true
			}
			if section, found := findNavigationSection(children, title); found {
				return section, true
			}
		}
	}

	return nil, false
}

func navigationTargets(nodes []any) []string {
	var targets []string
	for _, node := range nodes {
		switch item := node.(type) {
		case string:
			targets = append(targets, item)
		case map[any]any:
			for _, value := range item {
				if target, ok := value.(string); ok {
					targets = append(targets, target)
				}
			}
		}
	}

	return targets
}

func TestList(t *T.T) {
	t.Run("list-all-docs", func(t *T.T) {
		dirs, err := AllDocs.ReadDir("doc/zh")
		assert.NoError(t, err)

		t.Logf("get %d dirs", len(dirs))
		for _, x := range dirs {
			t.Logf("%s", x.Name())
		}
	})

	t.Run("list-all-dashboard", func(t *T.T) {
		t.Skip("all templates removed")

		dirs, err := AllTemplates.ReadDir("dashboard")
		assert.NoError(t, err)

		t.Logf("get %d dirs", len(dirs))
		for _, x := range dirs {
			t.Logf("%s, is dir: %v", x.Name(), x.IsDir())
		}
	})
}
