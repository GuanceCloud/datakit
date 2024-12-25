#!/usr/bin/env node

const shell = require("shelljs")
const inquirer = require('inquirer')
const fs = require("fs")
const path = require("path")
const OSS = require("ali-oss")
const _ = require("lodash")
const axios = require("axios").default
const { ArgumentParser } = require('argparse');

const actions = {
  cache: {},
  register(action, method) {
    if (this.cache[action]) {
      throw `action ${action} existed`
    } else {
      this.cache[action] = method
    }
  }
}

const parser = new ArgumentParser({
  description: 'Build DCA client'
})

parser.add_argument('--upload-app', { help: "upload app only", nargs: "?" })
parser.add_argument('--build-app', { help: "build app only", nargs: "?" })
parser.add_argument('--image-tag', { help: "image tag name", action: "append", nargs: "?" })
parser.add_argument('--image-url', { help: "image url address", nargs: "?" })
parser.add_argument('--upload-addr', { help: "upload oss address", nargs: "?" })
parser.add_argument('--download-cdn', { help: "downlad cdn address", nargs: "?" })
parser.add_argument('--release', { help: "release type", nargs: "?" })

parser.add_argument('action', { help: "action type, build_image only", nargs: "?" })

const ARGS = parser.parse_args()

const packageInfo = require("./package.json")
let releaseType = ARGS["release"]
if (!releaseType) {
  releaseType = "LOCAL"
} else {
  releaseType = releaseType.toUpperCase()
}

const LOCAL_OSS_BUCKET = process.env[`${releaseType}_OSS_BUCKET`]
const LOCAL_OSS_SECRET_KEY = process.env[`${releaseType}_OSS_SECRET_KEY`]
const LOCAL_OSS_ACCESS_KEY = process.env[`${releaseType}_OSS_ACCESS_KEY`]
const LOCAL_OSS_REGION = process.env[`${releaseType}_OSS_REGION`]

const { version: VERSION } = packageInfo

if (!VERSION) {
  console.error("can't get version from 'package.json'")
  shell.exit(1)
}

const ossClient = new OSS({
  region: LOCAL_OSS_REGION || 'oss-cn-hangzhou',
  accessKeyId: LOCAL_OSS_ACCESS_KEY,
  accessKeySecret: LOCAL_OSS_SECRET_KEY,
  bucket: LOCAL_OSS_BUCKET
})

const DIST_DIR = "dist"
const APP_DIR = "app"
const RENDERER_DIR = path.join(APP_DIR, "client")

const MODE = process.env.MODE || "prod" // prod default

const configPath = path.join(RENDERER_DIR, `src/config/config.${MODE}.ts`)
const destConfigPath = path.join(RENDERER_DIR, `src/config.ts`)

const packAll = async () => {
  if (!checkGit()) {
    const isValid = await inquirer
      .prompt([{
        type: "input",
        name: "git",
        message: "your working tree is not clean, continue to pack? [Y/N]",
        default: "N"
      }])
      .then((answers) => {
        return answers.git.toLowerCase() === "y"
      })
      .catch((error) => {
        console.log(error)
        return false
      })

    if (!isValid) {
      shell.exit(1)
    }
  }

  let res = shell.rm("-rf", DIST_DIR)
  if (res.code != 0) {
    shell.exit()
  }

  res = shell.mkdir(DIST_DIR)
  if (res.code != 0) {
    shell.exit()
  }

  res = shell.cp(configPath, destConfigPath)
  if (res.code != 0) {
    console.error("copy config failed", res)
    shell.exit()
  }
  console.log("cp config success")

  // build mac
  pack("mac")

  //build win
  pack("win")

}

async function uploadAll() {
  console.log("upload to oss....")
  await Promise.allSettled([
    uploadToOss(`./dist/mac/DCA-v${VERSION}.dmg`),
    uploadToOss(`./dist/win/DCA-v${VERSION}-x86.exe`)
  ])
}

async function uploadToOss(filePath) {
  const client = new OSS({
    region: LOCAL_OSS_REGION || 'oss-cn-hangzhou',
    accessKeyId: LOCAL_OSS_ACCESS_KEY,
    accessKeySecret: LOCAL_OSS_SECRET_KEY,
    bucket: LOCAL_OSS_BUCKET
  })

  let baseName = path.basename(filePath)
  const result = await client.putStream(`zhengbo/dca/v${VERSION}/${baseName}`, fs.createReadStream(filePath), { timeout: 600000 }).catch((err) => {
    console.error(err)
    return false
  })
  if (!result || !result.res || result.res.status != 200) {
    return false
  }
  console.log(`upload successfully to oss: ${result.name} `)
  return true
}

async function buildImg() {
  // let imageURL = "registry.jiagouyun.com/cloudcare-tools-pub/dca"
  let imageURL = "pubrepo.guance.com/image-repo-for-testing/dca"
  let [_, author] = runCmd("git log --pretty=format:%an -1")
  let tags = ARGS["image_tag"]
  if (ARGS["image_url"]) {
    imageURL = ARGS["image_url"]
  }
  if (!tags || tags.length === 0) {
    tags = ["latest"]
  }

  tags.push(VERSION)

  let images = tags.map((tag) => {
    return `-t ${imageURL}:${tag}`
  })

  let imageStr = images.join(" ")

  let { CI_COMMIT_BRANCH } = process.env
  let buildCmd = `
    docker buildx build \
      --platform linux/arm64,linux/amd64 \
      ${imageStr}\
      . \
      --push \
  `
  let [err] = runCmd(buildCmd)
  console.info(`build: ${buildCmd}`)
  if (err) {
    console.error(err)
    await notify("DCA CI Failed", `${author} 推送分支 ${CI_COMMIT_BRANCH}，构建 DCA 镜像失败！`)
    shell.exit(-1)
  }
  let imageListStr = ""
  tags.forEach(t => {
    imageListStr += `> ${imageURL}:${t}\n\n`
  });
  let image = `${imageURL}:${tags[0]}`

  let successText = `
## DCA 镜像发布
**${author}** 推送分支 ${CI_COMMIT_BRANCH}, 发布了新的 DCA 镜像:\n
${imageListStr}\n
**1. 下载镜像**\n
\`\`\`shell\ndocker pull  ${image}\n\`\`\`\n
**2. 运行容器**\n\n
\`\`\`shell\ndocker run --rm --name dca -p 8000:80  ${image}\n\`\`\`\n
**3. 访问网站**\n\n
http://127.0.0.1:8000\n  
`
  await notify("DataKit DCA 镜像发布", successText)
  await writeVersion()
  await writeYaml()
}

async function notify(title, text) {
  let { DINGDING_TOKEN: token } = process.env
  if (!token) {
    console.error("notify token is missing")
    return
  }
  await axios.post(`https://oapi.dingtalk.com/robot/send?access_token=${token}`, {
    msgtype: "markdown",
    markdown: {
      title,
      text
    }
  }).then((res) => {
    if (!res.data || res.data.errcode !== 0) {
      console.error("send ding message failed", res.data.errmsg)
      process.exit(res.data.errcode)
    }
  }).catch((err) => {
    console.error("send ding message failed", err)
  })
}

function pack(target) {
  console.info(`pack ${target} starting...`)
  const res = shell.exec(`npm run pack:${target}`)
  if (res.code != 0) {
    shell.exit()
  }
  console.info(`pack ${target} success`)
  shell.mv("release", path.join(DIST_DIR, target))
}

// write version
async function writeVersion() {
  const versionInfo = {
    "version": VERSION,
    "git": {
      "hash": ""
    }
  }

  const res = shell.exec("git rev-parse HEAD")
  if (res.code != 0) {
    console.error("get git branch hash error")
  } else {
    versionInfo["git"]["hash"] = res.trim()
  }

  const result = await ossClient.put(`datakit/dca/version`, Buffer.from(JSON.stringify(versionInfo)), { timeout: 600000 }).catch((err) => {
    console.error(err)
    return false
  })
  if (!result || !result.res || result.res.status != 200) {
    return false
  }
  console.log(`upload successfully to oss: ${result.name} `)
  return true
}

async function writeYaml() {
  fs.readFileSync("./dca.yaml")
  try {
    const template = fs.readFileSync("./dca.yaml", "utf8")
    const data = _.template(template)({
      version: VERSION,
    })

    const result = await ossClient.put(`datakit/dca/dca.yaml`, Buffer.from(data), { timeout: 600000 }).catch((err) => {
      console.error(err)
      return false
    })
    if (!result || !result.res || result.res.status != 200) {
      return false
    }
    console.log(`upload successfully to oss: ${result.name} `)
    return true
  } catch (error) {
    console.error(error)
    return false
  }

  return true
}

// check git status
function checkGit() {
  const res = shell.exec("git status -s")
  if (res.code != 0) {
    console.error("git status error")
  } else {
    console.log(res.toString())
    return res.trim() === ""
  }

  return false
}

function runCmd(cmd) {
  const res = shell.exec(cmd)
  if (res.code != 0) {
    return [res.stderr]
  } else {
    return [null, res.trim()]
  }
}

actions.register("build_app", packAll)
actions.register("upload_app", uploadAll)
actions.register("build_image", buildImg)
actions.register("test_ci", function () {
  console.log("test ci")
})

const run = async () => {
  console.log(ARGS)
  let action = ARGS["action"]

  if (typeof actions.cache[action] == "function") {
    await actions.cache[action]()
  } else {
    console.error(`invalid action: ${action}`)
  }
}

run().then(() => {
  console.log("Done!!")
}).catch((err) => {
  console.error(err)
  shell.exit(-1)
})