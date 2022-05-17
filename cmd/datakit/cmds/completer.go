// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

// generate auto completer command/options for DataKit

var (
	completerShell = []byte(`
# DataKit completion                             -*- shell-script -*-

# We should copy this script to /etc/bash_completion.d/datakit for unbuntu
# For Centos relalted Linux, we need more test.

_datakit()
{
	local cur_word prev_word
	cur_word="${COMP_WORDS[COMP_CWORD]}"
	prev_word="${COMP_WORDS[COMP_CWORD-1]}"

	cmd=${COMP_WORDS[1]}

	if [[ -n $cmd ]]; then # we have select specified command
		case "${cmd}" in
			dql)
				COMPREPLY=( $(compgen -W '--auto-json --csv -F,--force --host -J,--json --log -R,--run -T,--token -V,--verbose' -- "${cur_word}") )
				;;

			pipeline)
				COMPREPLY=( $(compgen -W '--date -F,--file --log --tab -T,--txt' -- "${cur_word}") )
				;;

			monitor)
				COMPREPLY=( $(compgen -W '-I,--input --log -W,--max-table-width -R,--refresh --to -V,--verbose' -- "${cur_word}") )
				;;

			service)
				COMPREPLY=( $(compgen -W '--log -I,--reinstall -R,--restart -S,--start -T,--stop -U,--uninstall' -- "${cur_word}") )
				;;

			install)
				COMPREPLY=( $(compgen -W '--log --ipdb --log --scheck --telegraf' -- "${cur_word}") )
				;;

			version)
				COMPREPLY=( $(compgen -W '--log -T,--testing --upgrade-info-off' -- "${cur_word}") )
				;;

			tool)
				COMPREPLY=( $(compgen -W '--check-config --check-sample --default-main-conf --dump-samples
				--ipinfo --log --show-cloud-info --upload-log --workspace-info --setup-completer-script
				--completer-script' -- "${cur_word}") )
				;;

			help)
				COMPREPLY=( $(compgen -W 'dql run pipeline service monitor install version tool'  -- "${cur_word}") )
				;;

			*)
				COMPREPLY=( $(compgen -W 'dql run pipeline service monitor install version tool help' -- "${cur_word}") )
				;;
			esac
	else # command not selected
		COMPREPLY=( $(compgen -W 'dql run pipeline service monitor install version tool help' -- "${cur_word}") )
	fi
} &&
complete -F _datakit datakit ddk

# ex: filetype=sh`)

	bashCompletionDirs = []string{
		"/usr/share/bash-completion/completions",
		"/etc/bash_completion.d",
	}
)

func setupCompleterScripts() {
	if runtime.GOOS != datakit.OSLinux { // only Linux support completion now
		return
	}

	cmd := exec.Command("/bin/bash", "-c", "complete")
	if err := cmd.Run(); err != nil {
		warnf("run completer failed: %s, skip\n", err)
		return
	}

	for _, dir := range bashCompletionDirs {
		if fi, err := os.Stat(dir); err != nil {
			warnf("%s not found: %s, skip\n", dir, err)
			continue
		} else {
			if !fi.IsDir() {
				warnf("invalid %s(not directory), skip\n", dir)
				continue
			}

			if err := ioutil.WriteFile(filepath.Join(dir, "datakit"), completerShell, os.ModePerm); err != nil {
				errorf("ioutil.WriteFile: %s\n", err)
				return
			}
		}
	}
}

func showCompletionScripts() {
	fmt.Println(string(completerShell))
}
