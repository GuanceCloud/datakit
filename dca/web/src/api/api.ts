import { NETWORK_TIMEOUT_CODE, ResonseError, type IDatakit } from "../store/type"
import logger from "./logger"
import { getMsg } from 'src/store/baseApi';
import i18n from "../i18n"

export const FETCH_ABORT_ERROR = "fetch aborted"

const apiPath = {
  datakit: {
    stats: "/api/datakit/stats",
    reload: "/api/datakit/reload",
    saveConfig: "/api/datakit/saveConfig",
    deleteConfig: "/api/datakit/deleteConfig",
    getConfig: "/api/datakit/getConfig",
    inputDoc: "/api/datakit/inputDoc",
    pipelines: "/api/datakit/pipelines",
    pipelineDetail: "/api/datakit/pipelines/detail",
    pipelineTest: "/api/datakit/pipelines/test",
    filter: "/api/datakit/filter",
    logTail: "/api/datakit/log/tail",
    logDownload: "/api/datakit/log/download"
  }
}

interface FetchTimeoutOpt extends RequestInit {
  timeout?: number
}

function fetchWithTimeout(url: string, opt: FetchTimeoutOpt = { timeout: 30000 }): Promise<Response> {
  let timeout = 30000
  if (opt.timeout) {
    timeout = opt.timeout
  }
  const controller = new AbortController()
  opt.signal = controller.signal
  let isTimeout = false
  const timer = setTimeout(() => {
    isTimeout = true
    controller.abort()
  }, timeout)
  return fetch(url, opt).then((r) => {
    clearTimeout(timer)
    return r
  }).catch((err) => {
    clearTimeout(timer)
    if (isTimeout) {
      return Promise.reject(new Error(NETWORK_TIMEOUT_CODE))
    }
    return Promise.reject(err)
  })
}

function getQueryPath(path: string, params?: Record<string, unknown>): string {
  let queryStrList: Array<string> = []
  let query = params || {}
  Object.keys(query).forEach((key: string) => {
    queryStrList.push(`${key}=${query[key]}`)
  })
  if (queryStrList.length > 0) {
    path += `?${queryStrList.join("&")}`
  }

  return path
}

export type GetInnerTokenData = {
  authToken?: string,
  idCode: string,
  needVerifyMFA: boolean
  tokenHoldTime?: number
  tokenMaxValidDuration?: number
}

function datakitApi(datakit: IDatakit) {
  let err

  const request = async <T>(path: string, params?: Record<string, unknown>, method: string = "GET"): Promise<[ResonseError | null, T | null]> => {
    if (err) {
      logger.error(err)
      return [{ errorCode: "500", message: err }, null]
    }
    let opt: RequestInit = {
      method,
      headers: {
        "Content-Type": "application/json; charset=utf-8",
      }
    }

    if (!params) {
      params = {}
    }
    if (method === "GET") {
      params["datakit_id"] = datakit.id
      path = getQueryPath(path, params)
    } else {
      path = path + "?datakit_id=" + datakit.id
    }
    if (["POST", "PUT", "PATCH", "DELETE"].includes(method)) {
      opt.body = JSON.stringify(params)
    }
    try {
      const hostURL = ""
      const url = `${hostURL}${path}`
      const resResult = await fetchWithTimeout(url, opt).then((r) => {
        if (r.status === 500) {
          const error: ResonseError = { errorCode: "500", message: i18n.t("api.data.fetch.failed") }
          return [error, null]
        }

        return r.json().then(async (res: DatakitApiResponse<T>) => {
          if (res.code === 200) {
            return [null, res.content]
          } else if (res.code === 401) {
            return [{ errorCode: "datakit.token.invalid", message: "token is not valid" }, null]
          } else if (res.code === 404) {
            return [{ errorCode: "route.not.found", message: "route not found" }, null]
          } else {
            const error: ResonseError = { errorCode: res.errorCode, message: res.message }
            logger.error(JSON.stringify(res))
            return [error, null]
          }
        })
      })
      logger.debug(JSON.stringify({ params: opt, result: resResult }))
      return resResult as [ResonseError | null, any]
    } catch (error: any) {
      logger.error(error)
      const code = error.message === NETWORK_TIMEOUT_CODE ? NETWORK_TIMEOUT_CODE : "network.error"
      return [{ errorCode: code, message: "network error" }, null]
    }

  }

  const get = async <T>(path: string, params?: Record<string, unknown>): Promise<[ResonseError | null, T | null]> => {
    return request<T>(path, params, "GET")
  }

  const post = async <T>(path: string, params?: Record<string, unknown>): Promise<[ResonseError | null, T | null]> => {
    return request<T>(path, params, "POST")
  }

  const deleteRequest = async <T>(path: string, params?: Record<string, unknown>): Promise<[ResonseError | null, T | null]> => {
    return request<T>(path, params, "DELETE")
  }

  const put = async <T>(path: string, params?: Record<string, unknown>): Promise<[ResonseError | null, T | null]> => {
    return request<T>(path, params, "PUT")
  }

  const patch = async <T>(path: string, params?: Record<string, unknown>): Promise<[ResonseError | null, T | null]> => {
    return request<T>(path, params, "PATCH")
  }

  return { get, post, put, patch, delete: deleteRequest }
}

type DatakitResponse<T> = [string | null, T | null | undefined]

type DatakitApiResponse<T> = {
  code: number
  content: T
  errorCode: string
  message: string
  success: boolean
}

export async function saveDatakitConfig(datakit: IDatakit, { path, config, isNew, inputName }): Promise<DatakitResponse<any>> {
  const [err, data] = await datakitApi(datakit).post(apiPath.datakit.saveConfig, {
    path,
    config,
    isNew,
    inputName
  })

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function deleteDatakitConfig(datakit: IDatakit, { path, inputName }): Promise<DatakitResponse<any>> {
  const [err, data] = await datakitApi(datakit).delete(apiPath.datakit.deleteConfig, {
    path,
    inputName,
  })

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function getDatakitConfig(datakit: IDatakit, path: string): Promise<DatakitResponse<any>> {
  const [err, data] = await datakitApi(datakit).get(apiPath.datakit.getConfig, { path })

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function getInputDoc(datakit: IDatakit, inputName: string): Promise<DatakitResponse<any>> {
  const [err, data] = await datakitApi(datakit).get(apiPath.datakit.inputDoc, { inputName })

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export type PipelineInfo = {
  fileName: string
  fileDir?: string
  content?: string
  category?: string
}

export async function getPipelineList(datakit: IDatakit): Promise<DatakitResponse<Array<PipelineInfo>>> {
  const [err, data] = await datakitApi(datakit).get<Array<PipelineInfo>>(apiPath.datakit.pipelines)

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export type FilterInfo = {
  filePath: string
  content: string
}

export type PipelineDetail = {
  path: string
  content: string
}

export async function getPipelineDetail(datakit: IDatakit, fileName: string, category: string): Promise<DatakitResponse<PipelineDetail>> {
  const [err, data] = await datakitApi(datakit).get<PipelineDetail>(apiPath.datakit.pipelineDetail, { fileName, category })

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function createPipeline(datakit: IDatakit, pipelineInfo: PipelineInfo): Promise<DatakitResponse<PipelineInfo>> {
  const [err, data] = await datakitApi(datakit).post<PipelineInfo>(apiPath.datakit.pipelines, pipelineInfo)

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function updatePipeline(datakit: IDatakit, pipelineInfo: PipelineInfo): Promise<DatakitResponse<PipelineInfo>> {
  if (!pipelineInfo.fileName) {
    return ["invalid fileName", null]
  }
  const [err, data] = await datakitApi(datakit).patch<PipelineInfo>(apiPath.datakit.pipelines, pipelineInfo)

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function deletePipeline(datakit: IDatakit, { category, fileName }): Promise<DatakitResponse<any>> {
  const [err, data] = await datakitApi(datakit).delete(apiPath.datakit.pipelines, {
    category,
    fileName,
  })

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

type testPipelineParams = {
  category: string,
  data: string[],
  script_name: string,
  pipeline: Record<string, Record<string, string>>,
}

export async function testPipeline(datakit: IDatakit, params: testPipelineParams): Promise<DatakitResponse<string>> {
  const [err, data] = await datakitApi(datakit).post<string>(apiPath.datakit.pipelineTest, params)

  if (err) {
    return [getMsg(err), null]
  }

  return [null, data]
}

export async function getLogTail(datakit: IDatakit, type = "log", controller?: AbortController): Promise<any> {
  if (!controller) {
    controller = new AbortController()
  }
  return fetch(apiPath.datakit.logTail + "?datakit_id=" + datakit.id + "&type=" + type, {
    method: "GET",
    signal: controller.signal,
  }).then((response) => {
    if (Math.floor(response.status / 100) !== 2) {
      return [i18n.t("api.log.fetch.failed"), null] as [string, ReadableStreamDefaultReader<Uint8Array> | null | undefined]
    }
    return ["", { reader: response.body?.getReader(), controller }]
  }).catch((err) => {
    if (err && err.name === "AbortError") {
      return [FETCH_ABORT_ERROR, null]
    }
    console.error("getLogTail error: ", err)
    return ["get log failed", null]
  })
}

export async function downloadLogFile(datakit: IDatakit, type: string = "log"): Promise<string | null> {
  const downloadURL = apiPath.datakit.logDownload + "?datakit_id=" + datakit.id + "&type=" + type
  window.open(downloadURL)
  return ""
}

