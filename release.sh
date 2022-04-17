#/bin/bash
# date: Wed Jul 21 11:29:22 CST 2021
# author: tan
# NOTE: 必须先发布 Mac 版本，不然 Mac 版本发布会缺少历史安装包入口，参见 #53

branch_name="$(git symbolic-ref HEAD 2>/dev/null)" ||
	branch_name="(unnamed branch)"     # detached HEAD

branch_name=${branch_name##refs/heads/} # remove suffix: refs/heads/

case $branch_name in
	"master"|"community"|"unstable") echo "release prod release..."
		if [[ "$OSTYPE" == "darwin"* ]]; then # Release darwin version first
			make production_mac VERSION=$1 &&
				echo "[I] darwin prod release ok"
		fi
		;;

	"github-mirror") echo "release to github & jihulab"
		git push github github-mirror
		git push github github-mirror --tags
		git push jihulab github-mirror
		git push jihulab github-mirror --tags
		;;

	*) echo "[E] unsupported branch '$branch_name' for release, exited"
		;;
esac
