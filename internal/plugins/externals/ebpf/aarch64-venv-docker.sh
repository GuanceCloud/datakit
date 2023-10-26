sudo docker run --privileged --rm tonistiigi/binfmt --install all

sudo docker run --platform arm64 -ti -v $(go env GOPATH)/src/github.com/GuanceCloud/datakit-ebpf:/root/go/src/github.com/GuanceCloud/datakit-ebpf \
    -w /root/go/src/github.com/GuanceCloud/datakit-ebpf pubrepo.jiagouyun.com/ebpf-dev/datakit-developer:1.7 /bin/bash

