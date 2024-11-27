
FROM pubrepo.guance.com/base/ubuntu:20.04 AS base
ARG TARGETARCH
ARG APP_DIR="/usr/local/dca/"

ENV DEBIAN_FRONTEND=noninteractive

RUN mkdir -p ${APP_DIR}

WORKDIR ${APP_DIR}

RUN sed -i 's/\(archive\|security\|ports\).ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list \
  && apt-get update \
  && apt-get --no-install-recommends install -y lsb-release gpg libaio-dev libaio1 unzip wget curl python3 python3-pip libxml2 alien

COPY ./dist/dca-linux-${TARGETARCH}/ .

COPY ./web/build ${APP_DIR}public/

CMD ["/usr/local/dca/dca"]

EXPOSE 80
