# datakit_ebpf

## 注意

* 在低版本内核(源码编译), 如 Linux Kernel 4.4.1 中：
  * 需要初始化局部变量, 由于这可能导致 BPF 验证器提示如 invalid indirect read from stack 等;
  * bpfmap 的 value 无法作为另一个 bpfmap 的 key, 需要先拷贝至局部变量;

开发环境配置参考

vscode settings.json 参考配置

```json
{
    ... ,

    "go.buildTags": "-tags ebpf",
    "C_Cpp.default.includePath": [
        "./plugins/externals/ebpf",
        "./plugins/externals/ebpf/c",
        "./plugins/externals/ebpf/c/common",
        "/usr/src/linux-headers-5.11.0-25-generic/arch/x86/include",
        "/usr/src/linux-headers-5.11.0-25-generic/arch/x86/include/uapi",
        "/usr/src/linux-headers-5.11.0-25-generic/arch/x86/include/generated",
        "/usr/src/linux-headers-5.11.0-25-generic/arch/x86/include/generated/uapi",
        "/usr/src/linux-headers-5.11.0-25-generic/include",
        "/usr/src/linux-headers-5.11.0-25-generic/include/uapi",
        "/usr/src/linux-headers-5.11.0-25-generic/include/generated/uapi"
    ],
    "C_Cpp.default.defines": ["__KERNEL__"],
    
    ... ,
}
```

```sh
apt-get update \
  && apt-get -y install tzdata curl tree \
  && apt-get -y install git make \
  && apt-get -y install gcc gcc-multilib \
  && apt-get -y install clang llvm \

# export https_proxy="http://..."

export PATH=$PATH:/usr/local/go/bin

curl -Lo go1.16.12.linux-amd64.tar.gz https://go.dev/dl/go1.16.12.linux-amd64.tar.gz \
  && tar -xzf go1.16.12.linux-amd64.tar.gz -C /usr/local/ \
  && rm go1.16.12.linux-amd64.tar.gz \
  && go install github.com/gobuffalo/packr/v2/packr2@v2.8.3 \
  && go install mvdan.cc/gofumpt@latest \
  && go get -u golang.org/x/tools/cmd/goyacc \
  && go get -u github.com/kevinburke/go-bindata/... \
  && curl -L https://github.com/golangci/golangci-lint/releases/download/v1.42.1/golangci-lint-1.42.1-linux-amd64.deb -o golangci-lint-1.42.1-linux-amd64.deb \
  && dpkg -i golangci-lint-1.42.1-linux-amd64.deb \
  && rm golangci-lint-1.42.1-linux-amd64.deb \
  && cp -r $HOME/go/bin/* /usr/local/bin
```
