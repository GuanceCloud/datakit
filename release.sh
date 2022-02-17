#/bin/bash
# date: Wed Jul 21 11:29:22 CST 2021
# author: tan
# NOTE: 必须先发布 Mac 版本，不然 Mac 版本发布会缺少历史安装包入口，参见 #53

branch_name="$(git symbolic-ref HEAD 2>/dev/null)" ||
branch_name="(unnamed branch)"     # detached HEAD

branch_name=${branch_name##refs/heads/}

new_tag=$1
latest_tag=$(git describe --abbrev=0 --tags)

case $branch_name in
	"testing")

		# Darwin's datakit is CGO-enabled, so build it locally
		if [[ "$OSTYPE" == "darwin"* ]]; then
			echo "release testing release for Darwin..."
			make testing_mac && make pub_testing_mac
		else
			echo "release testing DataKit without Darwin"
		fi

		git push origin testing
		;;

	"master") echo "release prod release..."
		if [ -z $new_tag ]; then
			echo "[E] new tag required to release production datakit, latest tag is ${latest_tag}"
		else
			git tag -f $new_tag  &&

			if [[ "$OSTYPE" == "darwin"* ]]; then # Release darwin version first
				make production_mac VERSION=$new_tag &&
				make pub_production_mac VERSION=$new_tag &&
				echo "[I] darwin prod release ok"
			fi

			# Trigger CI to release other platforms
			git push -f --tags   &&
			git push
		fi
		;;

	"github-mirror") echo "release to github"
		git push github github-mirror
		;;

	*) echo "[E] unsupported branch '$branch_name' for release, exited"
		;;
esac
