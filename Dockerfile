FROM pubrepo.guance.com/base/ubuntu:20.04 AS base
ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN mkdir -p /usr/local/datakit \
  && mkdir -p /usr/local/datakit/externals \
  && mkdir -p /opt/oracle

COPY dist/datakit-linux-${TARGETARCH}/ /usr/local/datakit/

RUN sed -i 's/\(archive\|security\|ports\).ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list \
  && apt-get update \
  && apt-get --no-install-recommends install -y redis-tools libaio-dev libaio1 unzip wget curl python3 python3-pip libxml2 alien \
  && pip3 install requests -i http://mirrors.aliyun.com/pypi/simple/ --trusted-host mirrors.aliyun.com \
  && rm -rf /var/lib/apt/lists/*

# download 3rd party libraries
RUN \
  case "$TARGETARCH" in \
  amd64) \
  wget -q https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip \
  -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
  && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
  && mv /opt/oracle/instantclient_21_10 /opt/oracle/instantclient \
  && rm /usr/local/datakit/externals/instantclient-basiclite-linux.zip; \
  wget -q https://static.guance.com/otn_software/db2/linuxx64_odbc_cli.tar.gz \
  -O /usr/local/datakit/externals/linuxx64_odbc_cli.tar.gz \
  && mkdir /opt/ibm \
  && tar zxf /usr/local/datakit/externals/linuxx64_odbc_cli.tar.gz -C /opt/ibm \
  && rm /usr/local/datakit/externals/linuxx64_odbc_cli.tar.gz; \
  wget -q https://static.guance.com/oceanbase/x86/libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm \
  -O /usr/local/datakit/externals/libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm \
  && alien -i /usr/local/datakit/externals/libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm \
  && rm /usr/local/datakit/externals/libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm; \
  wget -q https://static.guance.com/oceanbase/x86/obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm \
  -O /usr/local/datakit/externals/obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm \
  && alien -i /usr/local/datakit/externals/obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm \
  && rm /usr/local/datakit/externals/obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm; \
  ;; \
  esac;

RUN \
  case "$TARGETARCH" in \
  "arm64") \
  wget -q https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip \
  -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
  && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
  && mv /opt/oracle/instantclient_19_19 /opt/oracle/instantclient \
  && rm /usr/local/datakit/externals/instantclient-basiclite-linux.zip; \
  wget -q https://static.guance.com/oceanbase/arm/libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm \
  -O /usr/local/datakit/externals/libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm \
  && alien --target=arm64 -i /usr/local/datakit/externals/libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm \
  && rm /usr/local/datakit/externals/libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm; \
  wget -q https://static.guance.com/oceanbase/arm/obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm \
  -O /usr/local/datakit/externals/obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm \
  && alien --target=arm64 -i /usr/local/datakit/externals/obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm \
  && rm /usr/local/datakit/externals/obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm; \
  ;; \
  esac;

# download data files required by datakit
RUN wget -q -O data.tar.gz https://static.guance.com/datakit/data.tar.gz \
  && tar -xzf data.tar.gz -C /usr/local/datakit && rm -rf data.tar.gz

CMD ["/usr/local/datakit/datakit", "run", "-C"]
