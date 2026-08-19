import { message } from "antd"
import CryptoJS from 'crypto-js'
import { getMsg } from "src/store/baseApi"
import { DCA_STATUS, IDatakit, IDatakitVersionLine, ILatestDatakitVersions } from "src/store/type"

export async function sleep(time: number) {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(undefined)
    }, time)
  })
}

export var alertError = function () {
  let timer
  return (msg: any) => {
    if (!msg) {
      return
    }

    if (typeof msg === "object") {
      msg = getMsg(msg)
    }

    if (timer) {
      clearTimeout(timer)
      timer = setTimeout(() => {
        message.error(msg)
      }, 1000)
    } else {
      timer = setTimeout(() => {
        message.error(msg)
      }, 1000)
    }
  }
}()

export function aesEncrypt(word, keyWord = 'XwKsGlMcdPMEhR1B') {
  var key = CryptoJS.enc.Utf8.parse(keyWord)
  var srcs = CryptoJS.enc.Utf8.parse(word)
  var encrypted = CryptoJS.AES.encrypt(srcs, key, { mode: CryptoJS.mode.ECB, padding: CryptoJS.pad.Pkcs7 })
  return encrypted.toString()
}

export function isPhoneNumber(phone: string): boolean {
  let reg = /^((0\d{2,3}-\d{7,8})|(1[3456789]\d{9}))$/
  return reg.test(phone)
}

// time: ns
export function showDuration(time: number) {
  if (!time) {
    return ""
  }
  let unit = ['s', 'ms', 'µs', 'ns']
  let index = unit.length - 1
  while (index > 0 && time > 1000) {
    time = time / 1000
    index--
  }
  return time.toFixed(2) + unit[index]
}

export function isValidIP(ip: string): boolean {
  var reg = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
  return reg.test(ip);
}

export function isDatakitManagement(dk: IDatakit): boolean {
  return dk.status === DCA_STATUS.RUNNING
}

export function isLoadingStatus(dk: IDatakit): boolean {
  return [DCA_STATUS.UPGRADING, DCA_STATUS.RESTARTING].includes((dk.status as DCA_STATUS))
}

type ParsedDatakitVersion = {
  major: number
  minor: number
  patch: number
  suffix: string
}

function parseDatakitVersion(version: string): ParsedDatakitVersion | undefined {
  const normalized = version.trim().replace(/^v/, '').split('_', 1)[0]
  const match = normalized.match(/^(\d+)\.(\d+)\.(\d+)(?:-(.+))?$/)
  if (!match) {
    return undefined
  }

  return {
    major: Number(match[1]),
    minor: Number(match[2]),
    patch: Number(match[3]),
    suffix: match[4] || '',
  }
}

export function getDatakitVersionLine(version: string): IDatakitVersionLine | undefined {
  const parsed = parseDatakitVersion(version)
  if (parsed?.major === 1) {
    return "v1"
  }
  if (parsed?.major === 2) {
    return "v2"
  }
  return undefined
}

export function getLatestDatakitVersion(version: string, latestVersions: ILatestDatakitVersions): string | undefined {
  const line = getDatakitVersionLine(version)
  return line ? latestVersions[line] : undefined
}

function compareVersionSuffix(current: string, latest: string): number | undefined {
  if (current === latest) {
    return 0
  }

  const currentRC = current.match(/^rc(\d+)$/)
  const latestRC = latest.match(/^rc(\d+)$/)
  if (currentRC && latestRC) {
    return Number(currentRC[1]) - Number(latestRC[1])
  }
  if (currentRC && !latest) {
    return -1
  }
  if (!current && latestRC) {
    return 1
  }

  const currentBuild = current.match(/^(\d+)-g[0-9a-f]+$/i)
  const latestBuild = latest.match(/^(\d+)-g[0-9a-f]+$/i)
  if (currentBuild && latestBuild) {
    return Number(currentBuild[1]) - Number(latestBuild[1])
  }
  if (currentBuild && !latest) {
    return 1
  }
  if (!current && latestBuild) {
    return -1
  }

  return undefined
}

export function compareDatakitVersions(currentVersion: string, latestVersion: string): number | undefined {
  const current = parseDatakitVersion(currentVersion)
  const latest = parseDatakitVersion(latestVersion)
  if (!current || !latest) {
    return undefined
  }

  for (const field of ['major', 'minor', 'patch'] as const) {
    if (current[field] !== latest[field]) {
      return current[field] - latest[field]
    }
  }

  return compareVersionSuffix(current.suffix, latest.suffix)
}

export function isNewerDatakitVersionAvailable(version: string, latestVersions: ILatestDatakitVersions): boolean {
  const latestVersion = getLatestDatakitVersion(version, latestVersions)
  if (!latestVersion) {
    return false
  }
  const comparison = compareDatakitVersions(version, latestVersion)
  return comparison !== undefined && comparison < 0
}

export function isDatakitUpgradeable(dk: IDatakit, latestVersions: ILatestDatakitVersions): boolean {
  return isDatakitManagement(dk)
    && isNewerDatakitVersionAvailable(dk.version, latestVersions)
    && dk.status !== DCA_STATUS.OFFLINE
    && !isContainerMode(dk)
}

export async function runJob(limit, arr, fn) {
  let res: any = new Array(arr.length)
  const f = async function (i) {
    res[i] = fn(arr[i]).finally(() => limit++)
  }

  for (let i = 0; i < arr.length;) {
    let current = i
    if (limit > 0) {
      limit--
      f(current)
      i++
    } else {
      await sleep(1000)
    }
  }

  return Promise.allSettled(res)
}

export function isContainerMode(dk?: IDatakit): boolean {
  return dk?.run_in_container || false
}
