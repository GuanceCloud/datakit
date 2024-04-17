FROM pubrepo.guance.com/base/ubuntu:20.04 AS base
ARG TARGETARCH

RUN mkdir -p /usr/local/datakit-tools

COPY v1-write-linux-${TARGETARCH}.out /usr/local/datakit-tools/

RUN sed -i 's/\(archive\|security\|ports\).ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list \
    && apt-get update \
    && apt-get --no-install-recommends install -y wget curl

CMD ["/usr/local/datakit-tools/v1-write-linux-arm64.out", "-gin-log", "-listen", "0.0.0.0:54321"]
