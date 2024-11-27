import { Button } from "antd";
import TextArea from "antd/lib/input/TextArea";
import { useEffect, useState } from "react";
import { testPipeline } from "src/api/api";
import { JsonView, allExpanded, defaultStyles } from 'react-json-view-lite'
import 'react-json-view-lite/dist/index.css'
import './PipelineTest.scss'
import { alertError } from "src/helper/helper";
import { CloseOutlined, PlusOutlined } from "@ant-design/icons";

export default function PipelineTest({ datakit, fileName, category, pipeline }) {
  const [textArr, setTextArr] = useState<string[]>([""])
  const [parsedText, setParsedText] = useState<any>({})
  const [checkedSampleIndex, setCheckedSampleIndex] = useState<number>(0)

  const test = async (index?: number) => {
    if (!textArr) {
      return alertError("输入文本不能为空")
    }

    if (index === undefined) {
      index = checkedSampleIndex
    }

    if (textArr[index].length === 0) {
      return
    }

    let trueCatetory = category

    const [err, data] = await testPipeline(datakit, {
      category: trueCatetory,
      pipeline: {
        [trueCatetory]: { [fileName]: pipeline }
      },
      script_name: fileName,
      data: [textArr[index]]
    })
    if (err) {
      return alertError(err)
    }
    if (!data) {
      setParsedText({})
    } else {
      setParsedText(data)
    }
  }
  useEffect(() => {
    test(checkedSampleIndex)
  }, [checkedSampleIndex]) // eslint-disable-line react-hooks/exhaustive-deps

  const setText = (index, value) => {
    let arr = [...textArr]
    if (arr[index] !== undefined) {
      arr[index] = value
    }

    setTextArr(arr)
  }

  const addSample = () => {
    if (textArr.length >= 3) {
      return
    }

    textArr.push("")
    setTextArr([...textArr])
  }

  const deleteSample = (index) => {
    if (textArr.length === 1) {
      return
    }
    textArr.splice(index, 1)
    setTextArr([...textArr])
  }
  return (
    <div className="pipeline_test-container">
      <div className="text">
        {
          textArr.map((v, i) => {
            return <div className="text-content" key={i}>
              <TextArea
                value={v}
                autoSize={{ minRows: 3, maxRows: 3 }}
                onChange={({ target: { value } }) => setText(i, value)}
              ></TextArea>
              {
                textArr.length > 1 &&
                <div className="delete" onClick={() => deleteSample(i)}>
                  <CloseOutlined />
                </div>
              }
            </div>

          })
        }
      </div>

      <div className="setting-content">
        <div className={textArr.length >= 3 ? "add-sample disabled" : "add-sample"} onClick={addSample}>
          <PlusOutlined />添加样本
        </div>
        <div className="test-button" onClick={() => { test() }}>
          <Button size="small"><span className="fth-iconfont-play1"></span> &nbsp;开始测试</Button>
        </div>

      </div>
      <div className="result">
        <div className="header">
          <div className="title">基本信息</div>
          <ul className="sample-tabs">
            {
              textArr.map((v, i) => {
                return <li
                  key={i}
                  className={checkedSampleIndex === i ? "active" : ""}
                  onClick={() => setCheckedSampleIndex(i)}
                >样本{i + 1}</li>
              })
            }
          </ul>
        </div>
        <div className="result-content">
          <JsonView
            data={parsedText}
            shouldExpandNode={allExpanded}
            style={defaultStyles}
          />
        </div>
      </div>
    </div>
  )
}