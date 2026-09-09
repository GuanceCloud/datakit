# Pipeline JIT 发布库接入 DK CI

`gitlab-ci.yml` 的四个 testing/正式发布 job 默认启用 JIT 打包，全局变量保持关闭，避免普通 lint/UT/local 构建依赖 runtime：

```yaml
PIPELINE_JIT_ENABLED: "1"
PP_JIT_VERSION: "v0.1.1"
```

升级时只修改 `PP_JIT_VERSION`。下载脚本解析两个 manifest，验证源码 SHA 格式及跨架构一致性，再自动设置打包器的 `PLATYPUS_JIT_EXPECTED_REVISION`，无需手工配置 SHA。`PLATYPUS_JIT_RUNTIME_DIR` 默认指向 checkout 下 `.cache/pipeline-jit-runtime`。不要将 Token 写入仓库：在 DK 项目 CI/CD Variables 配置 Masked 的 `PP_JIT_READ_TOKEN`，使用 Rust 项目仅含 `read_package_registry` 权限的 Deploy Token。若变量设为 Protected，发布分支也必须受保护。

现有 `release-testing-guance`、`release-testing-tw`、`release-prod-guance`、`release-prod-tw` 在打包前调用 `scripts/download-pipeline-jit.sh`。这四个发布 job 默认设置 `PIPELINE_JIT_ENABLED=1`，下载正式包 `platypus-jit/<PP_JIT_VERSION>` 的 amd64/arm64 压缩包与外部校验和。现有 MR、testing、正式版 UT job 通过 `scripts/test-pipeline-jit-ci.sh` 下载同版本 runtime，并在当前 Linux runner 上运行 `go test -tags pipeline_jit ./internal/pipeline/...`。测试脚本仅对子进程下载开启打包变量，普通 local 构建不启用打包。下载或 native 测试失败会阻止 job 成功，发布 job 继续依赖对应 UT job。MR runner 同样需要可用的 `PP_JIT_READ_TOKEN`。该测试不代表另一架构的运行验证。

下载需要 Runner 的 Bash、Python 3（标准库）、curl、GNU tar、sha256sum、sort、cmp。脚本使用 `DEPLOY-TOKEN` 请求头，校验压缩包 SHA-256 及四个普通文件的布局，两个架构均通过才复制到打包输入目录。打包器继续校验 manifest、内部哈希、ELF、ABI/profile 和期望源码 SHA；失败停止打包，不使用旧库继续发布。临时下载目录在成功或普通失败后删除。

JIT 仅用于 Linux amd64/arm64 glibc 环境：对应 DK 安装包只携带自身架构的库。其他平台不启用 JIT 编译标签，不携带库，仍使用 Go 执行器。静态/musl 构建须显式设置 `PIPELINE_JIT_ENABLED=0`。CI 打包开关不改变运行时 JIT 默认关闭的配置。

本地验证命令：

```bash
python3 scripts/test_download_pipeline_jit.py
bash -n scripts/download-pipeline-jit.sh
go test ./cmd/make/build -run 'Test(StagePipelineJIT|PipelineJIT|ConfiguredPipelineJIT)' -count=1
```

Python 测试使用模拟下载和临时压缩包，不读取实际 Token。真实跨项目下载必须由 DK 发布 job 验证。
