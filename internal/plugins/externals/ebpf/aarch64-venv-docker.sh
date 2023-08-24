sudo docker run --privileged --rm tonistiigi/binfmt --install all

sudo docker run --platform arm64 -ti -v ${1}go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit:/root/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit \
    pubrepo.jiagouyun.com/ebpf-dev/datakit-developer:1.7 /bin/bash

