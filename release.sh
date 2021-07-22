#/bin/bash
<<<<<<< HEAD
# date: Wed Jul 21 11:29:22 CST 2021
# author: tan
# NOTE: 必须先发布 Mac 版本，不然 Mac 版本发布会缺少历史安装包入口，参见 #53

branch_name="$(git symbolic-ref HEAD 2>/dev/null)" ||
branch_name="(unnamed branch)"     # detached HEAD

branch_name=${branch_name##refs/heads/}

new_tag=$1
latest_tag=$(git describe --abbrev=0 --tags)

case $branch_name in
	"testing") echo "release test release..."
		make &&
		make pub_testing_mac &&
		git push origin testing
		;;

	"dev") echo "release prod release..."
		if [ -z $new_tag ]; then
			echo "[E] new tag required to release production datakit, latest tag is ${latest_tag}"
		else
			# Release darwin version first
			git tag -f $new_tag  &&
			make release_mac     &&
			make pub_release_mac &&

			echo "[I] darwin prod release ok"

			# Trigger CI to release other platforms
			git push -f --tags   &&
			git push
		fi
		;;

	"github") echo "release to github"
		git push github dev
		;;

	*) echo "[E] unsupported branch '$branch_name' for release, exited"
		;;
esac
=======
# date: Thu Apr 22 14:35:59 CST 2021
# author: tan
# add new release tag, then release the new version
# NOTE: 必须先发布 Mac 版本，不然 Mac 版本发布会缺少历史安装包入口，参见 #53
git tag -f $1        &&
make release_mac     &&
make pub_release_mac &&
git push -f --tags   &&
git push
>>>>>>> 959cf0636413128084d725ef8c54485504b7f177
