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
				COMPREPLY=( $(compgen -W '--auto-json --csv -F --force --host -J --json --log -R --run -T --token -V --verbose' -- "${cur_word}") )
				;;

			pipeline)
				COMPREPLY=( $(compgen -W '--date --file -F --log --tab -T --txt' -- "${cur_word}") )
				;;


			monitor)
				COMPREPLY=( $(compgen -W '-I --input --log -W --max-table-width -R --refresh --to -V --verbose' -- "${cur_word}") )
				;;

			service)
				COMPREPLY=( $(compgen -W '--log -I --reinstall -R --restart -S --start -T --stop -U --uninstall' -- "${cur_word}") )
				;;

			install)
				COMPREPLY=( $(compgen -W '--log --ipdb --log --scheck --telegraf' -- "${cur_word}") )
				;;

			tool | debug)
				COMPREPLY=( $(compgen -W '--check-config --check-sample --default-main-conf --dump-samples
				--ipinfo --log --show-cloud-info --upload-log --workspace-info' -- "${cur_word}") )
				;;

			help)
				COMPREPLY=( $(compgen -W 'dql run pipeline service monitor install tool'  -- "${cur_word}") )
				;;

			*)
				COMPREPLY=( $(compgen -W 'dql run pipeline service monitor install tool help' -- "${cur_word}") )
				;;
			esac
	else # command not selected
		COMPREPLY=( $(compgen -W 'dql run pipeline service monitor install tool help' -- "${cur_word}") )
	fi
} &&
complete -F _datakit datakit ddk

# ex: filetype=sh
