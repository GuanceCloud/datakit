import { App as AntdApp, ConfigProvider } from "antd";
import { RouterProvider, } from "react-router-dom";

import router from './router'

import 'antd/dist/reset.css';
import config from './config'

const { theme } = config

function App() {
  return (
    <ConfigProvider
      theme={
        theme
      }
    >
      <AntdApp style={{ height: "100%" }}>
        <RouterProvider router={router} />
      </AntdApp>
    </ConfigProvider>
  );
}

export default App;