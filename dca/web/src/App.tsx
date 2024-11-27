import { App as AntdApp, ConfigProvider } from "antd";
import { RouterProvider, } from "react-router-dom";

import router from './router'

import 'antd/dist/reset.css';

function App() {
  return (
    <ConfigProvider
      theme={{
        token: {
          fontSize: 12,
          colorPrimary: "#FF6600",
        },
        components: {
          Tabs: {
            fontSize: 14
          },
          Button: {
            colorPrimary: "#2F61CC",
            colorPrimaryHover: "#4d7ee6",
            colorPrimaryActive: "#2F61CC",
          },
          Divider: {
            colorSplit: "#F0F6FC"
          }
        }
      }}
    >
      <AntdApp style={{ height: "100%" }}>
        <RouterProvider router={router} />
      </AntdApp>
    </ConfigProvider>
  );
}

export default App;