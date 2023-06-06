# datakit_ebpf

## Notice

* In low-version kernel (source code compilation), such as Linux Kernel 4.4.1:
   * Local variables need to be initialized, because this may lead to BPF validator prompts such as invalid indirect read from stack, etc.;
   * The value of bpfmap cannot be used as the key of another bpfmap, it needs to be copied to a local variable first;

Development environment configuration reference

vscode settings.json reference configuration

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
  && curl -L https://github.com/golangci/golangci-lint/releases/download/v1.42.1/golangci-lint-1.42.1-linux-amd64.deb -o golangci-lint-1.42.1-linux-amd64.deb \
  && dpkg -i golangci-lint-1.42.1-linux-amd64.deb \
  && rm golangci-lint-1.42.1-linux-amd64.deb \
  && cp -r $HOME/go/bin/* /usr/local/bin
```
