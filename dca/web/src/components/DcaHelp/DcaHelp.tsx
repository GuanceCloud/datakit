import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import {CopyToClipboard} from 'react-copy-to-clipboard'
import {xonokai} from 'react-syntax-highlighter/dist/esm/styles/prism'
import {Prism as SyntaxHighlighter} from 'react-syntax-highlighter'
import gfm from 'remark-gfm'
import "./DcaHelp.scss"
import { Button, message, Spin } from 'antd'
import { CopyOutlined } from '@ant-design/icons'

export default function DcaHelp(){
  const [doc, setDoc] = useState<string>("")
  const components = {
    code ({...props})   {
      let {inline, className, children} = props
      const content = String(children).replace(/\n$/, '')
      const match = /language-(\w+)/.exec(className || 'language-toml')
      return !inline && match ? (
        <div className="highlight-code">
           <div className="copy">
            <CopyToClipboard text={content} onCopy={() => message.success("复制成功")}>
              <Button size="small" type='text'><CopyOutlined />复制代码</Button>
            </CopyToClipboard>
          </div>
          <SyntaxHighlighter 
            showLineNumbers={true} 
            style={xonokai} 
            customStyle={{paddingTop: '30px', borderRadius: "10px"}}
            language={match[1]}
            PreTag="div" 
            children={content} 
            {...props} 
          />
        </div>
      ) : (
        <code className={className + ' inline-code'} {...props}>
          {children}
        </code>
      )
    }
  }

  const initDoc = async () => {
    let helpDoc = "暂无数据"
    setDoc(helpDoc)
  }

  useEffect(() => {
    initDoc()
  }, [])

  return (
    <div className="dca-help-container">
      {
        !doc ?
          <div className="loading">
            <Spin/>
          </div>
        :
          <ReactMarkdown 
            components={components} 
            remarkPlugins={[[gfm, {singleTilde: false}]]}
            children={doc} 
            className="markdown"
          />
      }
    </div>
  )
}