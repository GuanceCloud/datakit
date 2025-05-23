export const NETWORK_TIMEOUT_CODE = "network.error.timeout"
export const CONSOLE_AUTH_TOKEN_FAILED = "ft.AuthTokenFailed"
export const DCA_AUTH_FAILED = "auth.failed"

export const ERRMSG: Record<string, string> = {
  "innerCorestone.auth.invalidEmailOrMobile": "用户名错误，请重新输入",
  "innerCorestone.auth.passwordIncorrect": "密码错误，请重新输入",
  "innerCorestone.auth.noSuchEmail": "用户名不存在，请重新输入",
  "innerCorestone.auth.MFALoginVerifyCacheExpired": "操作已过期，请重新操作",
  "innerCorestone.auth.MFAPassCodeInvalid": "输入的验证码不正确",
  "innerCorestone.auth.noSuchAccount": "用户不存在，请重新输入",
  "network.error": "网络异常",
  "ft.Forbidden": "您无权限操作",
  "server.request.error": "连接异常，请稍后再试",
  "server.error": "操作失败，请稍后再试",
  "server.request.timeout": "连接超时，请稍后再试",
  "toml.format.error": "当前配置文件格式错误",
  "pipeline.parse.error": "解析失败: {{msg}}",
  "route.not.found": "当前版本 DataKit 不支持该操作，请升级版本(>=1.29.0) ",
  "innerCorestone.auth.smsCodeExpired": "验证码已失效",
  "innerCorestone.auth.smsCodeIncorrect": "验证码错误，请重新获取",
  "whiteList.check.error": "该采集器已开启白名单，当前客户端无权限访问",
  "record.not.exist": "当前文档不存在或已被删除",
  "param.invalid.duplicate": "名称重复",
  "reload.connect.timeout": "重启中，尝试连接失败，请稍后再重新连接",
  "innerCorestone.auth.invalidMobile": "手机号码格式错误",
  "innerCorestone.auth.noSuchMobile": "手机号不存在，请重新输入",
  "request.error.invalidToken": "用户令牌已失效，请重新登录",
  "request.error.requirePermission": "当前用户无权限进行此操作",
  "request.error.notSupport": "当前 DataKit 不支持该操作",
  "datakit.token.invalid": "请检查 DataKit 的 token 是否正确",
  "dql.queryError": "DQL queryError",
  "dql.parseError": "DQL parseError",
  "dca.log.file.invalid": "日志文件未找到",
  "datakit.not.found": "请检查目标 DataKit 是否在线",
  "server.request.refused": "目标 DataKit 无法连接，请稍后再试",
  [NETWORK_TIMEOUT_CODE]: "请求超时，请重试",
  [DCA_AUTH_FAILED]: "当前无权限进行此操作",
}

export const enum DCA_STATUS {
  RUNNING = "running",
  OFFLINE = "offline",
  UPGRADING = "upgrading",
  STOPPED = "stopped",
  RESTARTING = "restarting",
}

export type PageInfo = {
  count: number
  pageIndex: number
  pageSize: number
  totalCount: number
}

export type PageQuery = {
  pageIndex: number
  pageSize: number
}

export type ResonseError = {
  errorCode: string
  message: string
}

export type WorkspaceRole = "owner" | "readOnly" | "wsAdmin" | "general" | "snapshotShared" | "shared" | "anonymous"

export type IWorkspace = {
  accountRole?: WorkspaceRole
  autoAggregation?: string
  createAt?: string
  id?: string
  isOpenWarehouse?: string
  enablePublicDataway?: string
  status?: string
  updateAt?: string
  deleteAt?: string
  dashboardUUID?: string | null
  billingState?: string
  cliToken?: string
  creator?: string
  dbUUID?: string
  desc?: string
  exterId?: string
  name: string
  wsName: string
  roles?: Array<{
    name: string
    uuid: string
  }>
  rpName?: string
  token?: string
  updator?: string
  uuid: string
  versionType?: string
  durationSet?: {
    backup_log: string
    keyevent: string
    logging: string
    rp: string
    rum: string
    security: string
    tracing: string
  }
  extend: {
    isAdmin: boolean
    role: WorkspaceRole
    permissions?: Array<string>
  }
}

export type IDatakit = {
  id: string
  arch: string
  host_name: string
  run_in_container: boolean
  version: string
  updated_at: number
  ip: string
  os: string
  runtime_id: string
  run_mode: string
  usage_cores: number
  uptime: number
  status: string
  update_time: number
  workspace_uuid: string
}

type configItem = {
  catalog: string
  config_dir: string
  sample_config: string
  pipeline_dir: string
  config_paths: Array<{
    loaded: number
    path: string
  }>
}

export type IDatakitStat = {
  config_info: {
    inputs: Record<string, configItem>
    datakit: configItem
  }
  available_inputs: string[]
  enabled_inputs: Array<{
    input: string
    instances: number
    panic: number
  }>
  enabled_input_list: Record<string, {
    input: string
    instances: number
    panic: number
  }>
  goroutine_stats: {
    Items: Record<string, {
      err_count: number
      finished_goroutines: number
      running_goroutines: number
      max_cost_time: string
      min_cost_time: string
      total_cost_time: string
    }>
    avg_cost_time: string
    finished_goroutines: number
    running_goroutines: number
    total_cost_time: string

  }
  inputs_status: Record<string, {
    avg_collect_cost: number
    avg_size: number
    pts_total: number
    max_collect_cost: number
    category: string
    first: string
    last: string
    last_error: string
    last_error_ts: string
    frequency?: string
    feed_total: number
    p90_pts: string
    p90_lat: string
  }>
  io_stats: {
    CO_fail_pts: number
    CO_send_pts: number
    E_fail_pts: number
    E_send_pts: number
    L_fail_pts: number
    L_send_pts: number
    M_fail_pts: number
    M_send_pts: number
    N_chan_pts: number
    N_fail_pts: number
    O_fail_pts: number
    O_send_pts: number
    P_chan_pts: number
    P_fail_pts: number
    R_fail_pts: number
    R_send_pts: number
    S_fail_pts: number
    S_send_pts: number
    T_fail_pts: number
    T_send_pts: number
    drop_pts: number
    chan_usage: Record<string, [number, number]>
  }
  reload_cnt: number
  auto_update: boolean
  docker: boolean
  branch: string
  build_at: string
  elected: string
  io_chan_stats: string
  os_arch: string
  reload: string
  reload_info: string
  uptime: string
  start_time?: number
  version: string
  hostname: string
  cgroup: string
  from: string
  open_files: number
  golang_runtime: {
    cpu_usage: number
    gc_avg_bytes: number
    gc_num: number
    gc_pause_total: number
    goroutines: number
    heap_alloc: number
    total_sys: number
  }
  datakit_runtime_info: {
    cpu_usage: string
  }
  http_metrics: {
    total_count: Number
    limited: Number
    limited_percent: Number
    "2xx": Number
    "3xx": Number
    "4xx": Number
    "5xx": Number
    max_letency: Number
    avg_latency: Number
  }
  filter_stats: {
    pull_count: number
    pull_interval: number
    pull_failed: number
    rule_source: string
    pull_cost: number
    pull_cost_avg: number
    pull_cost_max: number
    last_update: string
    last_err: string
    last_err_time: string
    rule_stats: Record<string, {
      conditions: number
      cost: number
      cost_per_point: number
      filtered: number
      total: number
    }>
  }
  pl_stats: {
    Pt: number
    PtDrop: number
    PtError: number
    RunLastErrs: string[]
    TotalCost: number
    MetaTS: string
    Script: string
    FirstTS: string
    ScriptTS: string
    ScriptUpdateTimes: number
    Category: string
    NS: string
    Name: string
    Enable: boolean
    Deleted: boolean
    CompileError: string
  }[]
  resource_limit: string
  usage_trace: {
    arch: string
    cpu_cores: number
    cpu_limites: number
    datakit_version: string
    hostname: string
    inputs: Array<string>
    ip: string
    os: string
    run_in_container: boolean
    run_mode: string
    runtime_id: string
    token: string
    upgrader_server: string
    uptime: number
    usage_cores: number
  }
}


export type IAccountPermission = {
  permissions: Array<string>
  role: string
  roles: Array<string>
}

export type IDatakitResponse<T> = {
  code: number
  content: T
  errorCode: string
  message: string
  success: boolean
}

export type IFilter = {
  content: string
  filePath: string
}

export type IVersion = {
  version: string
  date_utc: string
  uploader: string
  branch: string
  commit: string
  go: string
}

export type IAccountInfo = {
  name: string
  email: string
}