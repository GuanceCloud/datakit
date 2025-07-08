import { Button, Flex } from "antd";

import dcaImg from "../../assets/login.png";
import dcaImgTW from "../../assets/login-truewatch.png";
import { useTranslation } from "react-i18next";
import config from "../../config";

let imgSRC = config.brandName === "truewatch" ? dcaImgTW : dcaImg;

export default function Login() {
  const { t } = useTranslation();
  return <div style={{
    height: "100%",
    display: "flex",
    justifyContent: "center",
    alignItems: "center"
  }}>

    <Flex gap="middle" style={{ width: "400px" }} vertical >
      <div style={{ textAlign: "center", marginBottom: "20px" }}><img alt="dca" src={imgSRC} width="250" /></div>
      <div style={{ textAlign: "center", fontSize: "16px", color: "#555555" }}>{t("login_dca_redirect")} </div>
      <Button type="primary"
        style={{
          backgroundColor: config.color.primary,
          borderColor: config.color.primary,
          fontSize: "14px",
          fontWeight: 600,
          fontFamily: "PingFangSC, PingFang SC;"
        }} block>
        <a href="/console/dca" >{t("go_forward")}</a>
      </Button>
    </Flex >
  </div >
}