## gitlab开发文档

gitlab 有两种获取指标数据的方式，以下分析两种优劣。

### 访问gitlab服务API

最简单的做法是访问 gitlab 服务 API，只需要配置和提供 gitlab 用户 token，就可以拿到该用户权限下的诸如 `Groups`、`Projects`、`Branches` 和 `Commit` 等数据纤细，并且可以通过 API 进行 `create` 和 `delete` 等高级操作。

- 优点：简单易配，数据应有尽有。

- 缺点：API 是面向浏览器页面操作，提供对 git 的超精细控制，但是并不适合指标类的数据采集。旧版 gitlab 采集器就是用此方式，经考量该数据意义不大，故舍弃并进行重构。

### 访问gitlab-promtheus服务

gitlab 提供完整的 promtheus 接入方案，但是需要管理员进行开启，并在 gitlab 配置文件中添加访问端的白名单，具体做法详见 gitlab 使用文档。

- 优点：方案成熟，数据齐全（inputs/gitlab/promtheus-data 数据样本）。

- 缺点：promtheus 格式数据对行协议支持不友好，需要开发者自行决定所有数据的存留，不够精确且缺乏规范。

### 开发方案

经考量，最终决定使用 gitlab-promtheus 服务的方式，获取 gitlab 指标数据。

以固定 interval 访问 gitlab-promtheus 服务接口（例如：127.0.0.1:2080/-/metrics），获取 promtheus 格式数据后，只保留 `count` 和 `histogram` 类型数据，且 `histogram` 类型只存留 `count` 和 `sum` 值。

将 promtheus 格式数据转为行协议，所有选中数据的 measurement 都以 `gitlab_` 开头，tags 内容为 promtheus 数据中的 `label` 转换得来。

经过筛选，最终存得 20 个指标集，每个指标集包含 0-3 个 tag 和 1-2 个 fields，详见 manual.go 文件。
