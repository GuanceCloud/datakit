import { useContext, useEffect, useState } from "react";
import { CopyOutlined, DownOutlined, EditOutlined, ExclamationCircleOutlined, PlusCircleOutlined, QuestionCircleOutlined, SaveOutlined, SnippetsOutlined, ToolOutlined } from "@ant-design/icons";
import { useBlocker } from "react-router-dom";
import { Button, Dropdown, Form, Input, MenuProps, message, Modal, Space, Tooltip, Typography } from "antd";

import { DkInfoContext } from "../DkInfo";
import { Item } from '../../Common/FileList/FileList'
import './Pipeline.scss'
import { createPipeline, deletePipeline, getPipelineDetail, getPipelineList, PipelineInfo, updatePipeline } from "src/api/api";
import PipelineTest from "./PipelineTest/PipelineTest";
import { alertError, isContainerMode } from "src/helper/helper";
import { useIsAdmin } from "src/hooks/useIsAdmin";
import DCAEditor, { DCAEditorConfiguration } from "src/components/DCAEditor/DCAEditor";

const { Text } = Typography

const DEFAULT_CATEGORY = "default"

const defaultMenu = <div>default <Tooltip placement="right" title="pipeline 目录下的文件，可用于日志数据处理。"><QuestionCircleOutlined /></Tooltip></div>
function isPipelineName(name?: string): boolean {
  if (!name) {
    return false
  }

  const fileName = name.split(/[/\\]/).pop()
  return !!(fileName && fileName.endsWith(".p"))
}

function getFilePath(item?: PipelineInfo | null, os: string = 'linux'): string {
  if (!item || !item.fileDir || !item.fileName) {
    return ""
  }
  const sep: string = os === "windows" ? "\\" : "/";

  if (item.category) {
    return `${item.fileDir}${sep}${item.category}${sep}${item.fileName}`
  }
  return `${item.fileDir}${sep}${item.fileName}`
}

type ValidateStatus = Parameters<typeof Form.Item>[0]['validateStatus'];

type selectedItem = {
  name: string
  category: string
  content: string
  path: string
}

export default function Pipeline() {
  const [fileData, setFileData] = useState<Record<string, any>>({})
  const [fileCategory, setFileCategory] = useState<MenuProps['items']>()
  const [selected, setSelected] = useState<selectedItem | null>(null)
  const [isEdit, setIsEdit] = useState<boolean>(false)
  const [code, setCode] = useState<string>("")
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isTest, setIsTest] = useState(false);
  const [editorOptions, setEditorOptions] = useState<DCAEditorConfiguration>({ readOnly: true })

  const isAdmin = useIsAdmin()

  const [isCopy, setIsCopy] = useState(false)
  const [currentCategory, setCurrentCategory] = useState("default")
  const [currentFiles, setCurrentFiles] = useState([])

  const { datakit } = useContext(DkInfoContext)

  let blocker = useBlocker(
    ({ currentLocation, nextLocation }) => {
      return currentLocation.pathname !== nextLocation.pathname && isEdit
    }
  )

  useEffect(() => {
    if (blocker.state === "blocked") {
      Modal.confirm({
        title: '确认',
        icon: <ExclamationCircleOutlined />,
        content: '当前处于修改状态，确定离开吗？',
        okText: '确认',
        cancelText: '取消',
        centered: true,
        onOk: () => {
          blocker.proceed()
        },
        onCancel() {
          blocker.reset()
        }
      })
    }
  }, [blocker])

  useEffect(() => {
    setEditorOptions((opt) => {
      return { ...opt, readOnly: !isEdit }
    })
  }, [isEdit])

  const [form] = Form.useForm()
  const [fileName, setFileName] = useState<{
    value: string;
    validateStatus?: ValidateStatus;
    errorMsg?: string | null;
  }>({
    value: '',
  })

  const validateFileName = (
    name: string,
  ): { validateStatus: ValidateStatus; errorMsg: string | null } => {
    if (isPipelineName(name)) {
      return {
        validateStatus: 'success',
        errorMsg: null,
      };
    }
    return {
      validateStatus: 'error',
      errorMsg: '名称不能为空，且需要以 .p 结尾',
    };
  }

  const newPipeline = async (category: string) => {
    setCurrentCategory(category)
    setIsCopy(false)
    setCode("")
    setIsModalVisible(true)
  }

  useEffect(() => {
    initFileList()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakit])

  useEffect(() => {
    let category = currentCategory === "default" ? DEFAULT_CATEGORY : currentCategory
    setCurrentFiles(fileData[category])
  }, [currentCategory, fileData])

  useEffect(() => {
    if (!selected) return
    setCode(selected.content || "")
  }, [selected])

  if (!datakit) {
    return
  }

  const doSelectFile = async (f) => {
    let dirName = currentCategory
    if (dirName === DEFAULT_CATEGORY) {
      dirName = ""
    }

    setIsTest(false)
    const [err, detail] = await getPipelineDetail(datakit, f.name, dirName)
    if (err) {
      alertError(err)
      return
    }
    setIsEdit(false)
    if (detail) {
      if (typeof detail === "string") { // deprecated.
        setCode(detail)
        setSelected({ name: f.name, category: currentCategory, path: "", content: detail })
      } else {
        setCode(detail.content)
        setSelected({ name: f.name, path: detail.path, content: detail.content, category: currentCategory })
      }
    }

  }

  const selectFile = async (f) => {
    if (isEdit) {
      Modal.confirm({
        title: '确认',
        icon: <ExclamationCircleOutlined />,
        content: '当前处于修改状态，确定离开吗？',
        okText: '确认',
        cancelText: '取消',
        centered: true,
        onOk: () => {
          doSelectFile(f)
        }
      })
    } else {
      doSelectFile(f)
    }
  }

  const onFileNameChange = (e) => {
    setFileName({
      ...validateFileName(e.target.value),
      value: e.target.value
    })
  }

  const copyFile = () => {
    setIsCopy(true)
    setIsModalVisible(true)
  }

  const save = async () => {
    if (!selected) return

    if (!selected.name || !code) {
      return alertError("文件名或文件内容不能为空")
    }
    const pipeline: PipelineInfo = { fileName: selected.name, content: code, category: selected.category === DEFAULT_CATEGORY ? "" : selected.category }
    const [err] = await updatePipeline(datakit, pipeline)
    if (err) {
      return alertError(err)
    }

    message.success("修改成功")
    setSelected({ ...selected, content: code })

    setIsEdit(false)
  }
  const edit = () => {
    setIsEdit(true)
  }
  const test = () => {
    setIsTest(!isTest)
  }

  const cancelEdit = () => {
    Modal.confirm({
      title: '确认',
      icon: <ExclamationCircleOutlined />,
      content: '当前处于修改状态，确定取消吗？',
      okText: '确认',
      cancelText: '取消',
      centered: true,
      onOk: () => {
        setIsEdit(false)
        setCode(selected ? selected.content : "")
      }
    })
  }

  const deletePipelineFile = async () => {
    if (!selected) {
      return
    }

    const handleDelete = async () => {
      let [err] = await deletePipeline(datakit, { category: selected.category === DEFAULT_CATEGORY ? "" : selected.category, fileName: selected.name })

      if (err) {
        alertError(err)
      } else {
        message.success("删除成功")
        setSelected(null)
        initFileList()
      }
    }

    Modal.confirm({
      title: "请确认",
      icon: <ExclamationCircleOutlined />,
      content: '确定要删除当前文件吗?',
      okText: '确认',
      cancelText: '取消',
      onOk: handleDelete,
    })
  }

  const handleOk = async () => {
    try {
      let { fileName } = await form.validateFields()
      if (isPipelineName(fileName)) {
        const name = (fileName as string).endsWith(".p") ? fileName : (fileName + ".p")
        const [err, data] = await createPipeline(datakit, { fileName: name, content: code, category: currentCategory })
        if (err) {
          return alertError(err)
        }
        message.success("保存成功")
        await initFileList()
        const newItem: selectedItem = {
          name,
          path: getFilePath(data, datakit.os),
          content: code,
          category: currentCategory === "" ? DEFAULT_CATEGORY : currentCategory
        }
        setSelected(newItem)
        setCode(newItem.content)
        setIsModalVisible(false)
        setFileName({ value: "" })
        setIsEdit(false)
        setCurrentCategory(currentCategory)
      } else {
        setFileName({
          ...validateFileName(fileName.value),
          value: fileName.value
        })
      }
    } catch (err) {
      console.error(err)
    }
  }
  const handleCancel = () => {
    setIsModalVisible(false)
    setFileName({ value: "" })
  }

  const selectCategory: MenuProps["onClick"] = ({ key }) => {
    if (fileCategory) {
      let item = fileCategory[key]
      setCurrentCategory(item.label)
    }
  }

  const initFileList = async () => {
    const [err, pipelineFileList] = await getPipelineList(datakit)
    if (err) {
      alertError(err)
      return
    }

    let fileData = {}

    pipelineFileList?.forEach((pipeline) => {
      const name = pipeline.fileName
      let category = pipeline.category
      if (!name && !category) {
        return
      }

      if (!category) {
        category = DEFAULT_CATEGORY
      }

      let files = fileData[category]
      if (!files) {
        fileData[category] = []
        files = fileData[category]
      }

      if (name) {
        const f: Item = {
          name,
          content: pipeline.content || "",
          path: getFilePath(pipeline, datakit.os), //pipeline.fileDir + "/" + name,
        }
        files.push(f)
      }
    })

    const fileItems: MenuProps['items'] = []
    let i = 0
    for (let k in fileData) {
      fileItems.push(
        {
          key: `${i++}`,
          label: k === DEFAULT_CATEGORY ? "default" : k,
        }
      )
    }

    setFileCategory(fileItems)
    setFileData(fileData)
    setCurrentFiles(fileData[currentCategory])
    setSelected(null)
  }

  return (
    <div className="pipeline-container">
      <div className="file-info tab-container">
        <div className="file-list" style={{ borderRightWidth: 1 }}>
          <div className="content"
            style={{
              width: 200,
              border: "none",
              overflowX: "hidden",
              transition: "width .5s",
            }}
          >
            <Dropdown
              menu={{ items: fileCategory, onClick: selectCategory }}
            >
              <div className="title">
                <Space>
                  <SnippetsOutlined />
                  {currentCategory === DEFAULT_CATEGORY ? defaultMenu : currentCategory}
                </Space>
                <DownOutlined />
              </div>
            </Dropdown>
            {isAdmin && !isContainerMode(datakit) &&
              <div className="new-button">
                <div className="wrap" onClick={() => newPipeline(currentCategory)}>
                  <Space>
                    <PlusCircleOutlined />
                    新建 Pipeline
                  </Space>

                </div>
              </div>
            }
            <div className="files">
              {
                currentFiles &&
                currentFiles.map((v: any) => {
                  return <div key={v.name} className={selected?.name === v.name ? "item selected" : "item"} onClick={() => selectFile(v)} >
                    <Text
                      style={{ width: 150 }}
                      ellipsis={{ tooltip: { placement: "rightBottom" } }}
                    >
                      {v.name}
                    </Text>
                  </div>
                })
              }
            </div>
          </div>
        </div>
        <div className="file-detail">
          {selected ?
            <>
              <div className="setting">
                <div className="path">
                  {`路径： ${selected?.path}`}
                </div>
                <div className="edit">
                  <div className="setting-left">
                    <Space>
                      {
                        isAdmin && !isContainerMode(datakit) &&

                        (

                          !isEdit ?
                            <>
                              <Button size="small" type="primary" className="button" onClick={edit}>
                                <EditOutlined />
                                编辑
                              </Button>
                              <Button size="small" className="button" onClick={copyFile}>
                                <CopyOutlined />
                                克隆
                              </Button>
                            </>
                            :
                            <>
                              <Button size="small" type="primary" className="button" onClick={save}>
                                <SaveOutlined />
                                保存
                              </Button>
                              <Button size="small" className="button" onClick={cancelEdit}>
                                取消
                              </Button>
                            </>)
                      }
                      {
                        isAdmin && !isContainerMode(datakit) &&
                        <>
                          <Button size="small" className="button" onClick={test}>
                            <ToolOutlined />
                            <span >{isTest ? "关闭测试" : "测试"}</span>
                          </Button>
                          <Button type="default" disabled={isEdit} size="small" onClick={() => deletePipelineFile()}>
                            <span className="fth-iconfont-trash size-14"></span>
                          </Button>
                        </>
                      }
                    </Space>
                  </div>
                </div>
              </div>
              <div className="content">
                <div className="content-body">
                  <DCAEditor value={code} setValue={setCode} editorOptions={editorOptions} />
                </div>
                {
                  isTest &&
                  <div className="test">
                    <PipelineTest
                      datakit={datakit}
                      fileName={selected?.name}
                      pipeline={code}
                      category={currentCategory}></PipelineTest>
                  </div>
                }

              </div>
            </>
            :
            <div></div>
          }
        </div>

      </div>

      <Modal
        title={`${isCopy ? "克隆" : "新建"} Pipeline`}
        open={isModalVisible}
        onOk={handleOk}
        onCancel={handleCancel}
        destroyOnClose={true}
        okText="确定"
        cancelText="取消"
      >
        <div style={{ padding: "50px 50px 10px 50px" }}>
          <div style={{ paddingBottom: "5px", fontSize: "14px" }}>Pipeline 名称</div>
          <Form
            autoComplete="off"
            name="basic"
            form={form}
            preserve={false}
            labelCol={{ span: 0 }}
            wrapperCol={{ span: 24 }}
          >
            <Form.Item
              label=""
              name="fileName"
              validateStatus={fileName.validateStatus}
              help={fileName.errorMsg || ""}
              initialValue={fileName.value}
            >
              <Input value={fileName.value} onChange={onFileNameChange} />
            </Form.Item>
          </Form>
        </div>
      </Modal>
    </div>
  )
}
