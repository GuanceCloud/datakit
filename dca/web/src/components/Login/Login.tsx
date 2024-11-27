import { Button, Flex } from "antd";

import dcaImg from "../../assets/login.png";

export default function Login() {
  return <div style={{
    height: "100%",
    display: "flex",
    justifyContent: "center",
    alignItems: "center"
  }}>

    <Flex gap="middle" style={{ width: "360px" }} vertical >
      <div style={{ textAlign: "center", marginBottom: "20px" }}><img alt="dca" src={dcaImg} width="250" /></div>
      <div style={{ textAlign: "center", fontSize: "16px", color: "#555555" }}>请登录观测云<br />通过【集成-DCA】入口跳转进入</div>
      <Button type="primary"
        style={{
          backgroundColor: "#FF6600",
          borderColor: "#FF6600",
          fontSize: "14px",
          fontWeight: 600,
          fontFamily: "PingFangSC, PingFang SC;"
        }} block>
        <a href="/console/dca" >立即前往</a>
      </Button>
    </Flex >
  </div >
}