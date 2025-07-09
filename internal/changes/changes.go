// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package changes provides templating and processing for Kubernetes and host changes
package changes

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pelletier/go-toml/v2"
)

var globalManifests = Manifests{}

func LoadAllManifests() error {
	var err error
	if globalManifests.K8sManifest, err = loadManifest("manifests/k8s.toml"); err != nil {
		return err
	}
	return nil
}

func loadManifest(path string) (*Manifest, error) {
	data, err := AllManifests.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func RenderK8sTemplate(language Language, changeID ChangeID, data interface{}) (title string, message string, err error) {
	if globalManifests.K8sManifest == nil {
		return "", "", fmt.Errorf("k8s manifest not initialized")
	}

	for _, change := range globalManifests.K8sManifest.Changes {
		if change.ID != string(changeID) {
			continue
		}

		switch language {
		case LangEn:
			title, err = renderTemplate(change.Title.En, data)
			if err != nil {
				return
			}
			message, err = renderTemplate(change.Message.En, data)
		case LangZh:
			title, err = renderTemplate(change.Title.Zh, data)
			if err != nil {
				return
			}
			message, err = renderTemplate(change.Message.Zh, data)
		default:
			err = fmt.Errorf("unrechable error")
			// nil
		}
		return
	}
	return "", "", fmt.Errorf("unrechable, not found changeID %s", changeID)
}

func renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
