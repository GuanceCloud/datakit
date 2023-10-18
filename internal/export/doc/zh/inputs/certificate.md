<!-- markdownlint-disable MD025 -->
# OpenSSL 生成自签证书 {#self-signed-certificate-with-OpenSSL}
<!-- markdownlint-enable -->

---

本文介绍如何通过 OpenSSL cli 生成自签证书用于开启安全连接层(ssl)

## 生成步骤 {#certificate-gen-steps}

- step1: 生成私钥，私钥用于开启加密和身份验证

  ```shell
  openssl genrsa -out domain.key 2048
  ```

    - `genrsa`: 生成 rsa 密钥
    - `out`: 私钥文件
    - `2048`: 位数

- step2: 生成证书签名请求，证书签名请求(CSR)用于对证书发起签名，文件中包括公钥信息和生成证书的必要信息

  ```shell
  openssl req -key domain.key -new -out domain.csr
  ```

    - `req`: 证书请求命令
    - `key`: 私钥文件
    - `new`: 新建
    - `out`: csr 文件

  生成过程将以命令交互的方式进行，如下：

  ```shell
  Enter pass phrase for domain.key:
  You are about to be asked to enter information that will be incorporated
  into your certificate request.
  What you are about to enter is what is called a Distinguished Name or a DN.
  There are quite a few fields but you can leave some blank
  For some fields there will be a default value,
  If you enter '.', the field will be left blank.
  -----
  Country Name (2 letter code) [AU]:AU
  State or Province Name (full name) [Some-State]:stateA
  Locality Name (eg, city) []:cityA
  Organization Name (eg, company) [Internet Widgits Pty Ltd]:companyA
  Organizational Unit Name (eg, section) []:sectionA
  Common Name (e.g. server FQDN or YOUR name) []:domain
  Email Address []:email@email.com

  Please enter the following 'extra' attributes
  to be sent with your certificate request
  A challenge password []:
  An optional company name []:
  ```

  > 注意：Common Name 提示符下输入实际使用的合法域名

- step3: 生成自签证书，在不需要进行授信链检查的环境下可以直接生成自签证书用于开启证书服务

  ```shell
  openssl x509 -signkey domain.key -in domain.csr -req -days 365 -out domain.crt
  ```

    - `x509`: 生成 x509 规范的证书
    - `signkey`: 私钥文件
    - `in`: 输入 csr 文件
    - `days`: 有效期
    - `out`: 证书文件

- step4: 生成自签 CA

  生成 CA 跟证书和私钥

  ```shell
  openssl req -x509 -sha256 -days 1825 -newkey rsa:2048 -keyout rootCA.key -out rootCA.crt
  ```

    - `req`: 证书请求命令
    - `x509`: 生成 x509 规范的证书
    - `sha256`: sha256 哈希算法
    - `days`: 过期天数
    - `newkey`: 新建 key
    - `rsa:2048`: rsa 加密算法 2048 位
    - `keyout`: 私钥文件
    - `out`: 证书文件

  使用根 CA 为证书签名

  创建 ext 文件

  > 注意：domain 处输入实际使用的合法域名

  ```shell
  authorityKeyIdentifier=keyid,issuer
  basicConstraints=CA:FALSE
  subjectAltName = @alt_names
  [alt_names]
  DNS.1 = domain
  ```

  使用跟证书和私钥签发证书

  ```shell
  openssl x509 -req -CA rootCA.crt -CAkey rootCA.key -in domain.csr -out domain.crt -days 365 -CAcreateserial -extfile domain.ext
  ```

## 参考文献 {#certificate-references}

- [Generating Private Key](https://www.herongyang.com/Cryptography/keytool-Import-Key-openssl-genrsa-Command.html){:target="_blank"}
- [Certificate With OpenSSL](https://www.baeldung.com/openssl-self-signed-cert){:target="_blank"}
