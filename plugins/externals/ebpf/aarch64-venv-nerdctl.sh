nerdctl run --privileged --rm tonistiigi/binfmt --install all

nerdctl run --platform arm64 -ti -v ${1}go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit:/root/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit \
    vircoys/datakit-developer:1.5 -- /bin/bash

