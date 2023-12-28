sudo docker run --privileged --rm tonistiigi/binfmt --install all

sudo docker run --platform amd64 -ti -v /var/www/html/:/var/www/html/ -v $(go env GOPATH)/src/:/root/go/src/ \
    -w /root/go/src/ pubrepo.jiagouyun.com/ebpf-dev/datakit-developer:1.7 /bin/bash
