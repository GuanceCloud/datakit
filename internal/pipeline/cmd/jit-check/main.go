// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

// jit-check prints a reproducible rollout identity without executing Points.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/manager"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type inspection struct {
	Allow            pljit.AllowRule    `json:"allow"`
	RuntimeSHA256    string             `json:"runtime_sha256,omitempty"`
	Check            *pljit.CheckResult `json:"check,omitempty"`
	ExternalPatterns bool               `json:"external_patterns,omitempty"`
}

func inspect(directory, namespace, category, entry, runtimePath, runtimeSHA string) (inspection, error) {
	var out inspection
	cat := point.CatString(category)
	if cat == point.UnknownCategory {
		return out, fmt.Errorf("unknown category %q", category)
	}
	paths, err := manager.SearchScripts(directory)
	if err != nil {
		return out, err
	}
	sources := make(map[string]string, len(paths))
	for _, path := range paths {
		name, source, err := manager.ReadScript(path)
		if err != nil {
			return out, err
		}
		sources[name] = source
	}
	if _, ok := sources[entry]; !ok {
		return out, fmt.Errorf("script %q is absent from %s", entry, directory)
	}
	m := plval.NewScriptManager(nil, nil)
	if err := m.LoadScriptWithCatChecked(cat, namespace, sources, nil); err != nil {
		return out, err
	}
	defer m.LoadScriptWithCatChecked(cat, namespace, nil, nil) //nolint:errcheck
	lease, ok := m.Acquire()
	if !ok {
		return out, fmt.Errorf("script manager unavailable")
	}
	defer lease.Release()
	script, ok := lease.Manager().QueryScript(cat, entry, struct{}{})
	if !ok {
		return out, fmt.Errorf("script not selected")
	}
	snapshot, err := lease.JITModuleSnapshot(cat, script)
	if err != nil {
		return out, err
	}
	out.Allow = pljit.AllowRule{Category: category, Namespace: namespace, Script: entry, BundleSHA256: fmt.Sprintf("%x", snapshot.BundleIdentity())}
	out.ExternalPatterns = snapshot.RequiresExternalRules()
	if _, err := pljit.NewRolloutPolicy("allowlist", []pljit.AllowRule{out.Allow}, nil); err != nil {
		return out, err
	}
	if runtimePath != "" {
		if runtimeSHA == "" {
			return out, fmt.Errorf("--runtime requires --runtime-sha256 from a trusted build")
		}
		digest, err := pljit.VerifyRuntime(runtimePath, runtimeSHA, false)
		if err != nil {
			return out, err
		}
		runner, err := pljit.NewRunnerWithServices(runtimePath, "pipeline-go-1.4.3-datakit", 256, nil, pljit.ServiceConfig{Version: 1})
		if err != nil {
			return out, err
		}
		defer runner.Close() //nolint:errcheck
		var check pljit.CheckResult
		if snapshot.HasDependencies() {
			check = runner.CheckModules(snapshot)
		} else {
			check = runner.Check(snapshot.EntrySource())
		}
		out.Check = &check
		out.RuntimeSHA256 = fmt.Sprintf("%x", digest)
	}
	return out, nil
}

func main() {
	directory := flag.String("directory", ".", "directory containing this category's .p scripts")
	namespace := flag.String("namespace", "default", "exact DataKit script namespace")
	category := flag.String("category", "logging", "DataKit category")
	entry := flag.String("script", "", "entry script name including .p")
	runtimePath := flag.String("runtime", "", "optional trusted runtime for capability checks")
	runtimeSHA := flag.String("runtime-sha256", "", "SHA-256 pinned by the trusted build")
	patterns := flag.String("patterns-directory", "", "optional DataKit external grok pattern directory")
	flag.Parse()
	if *patterns != "" {
		if err := plval.LoadPatterns(*patterns); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	out, err := inspect(*directory, *namespace, *category, *entry, *runtimePath, *runtimeSHA)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
