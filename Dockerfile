FROM ubuntu:18.04 AS base

RUN mkdir -p /usr/local/cloudcare/dataflux/datakit \
    && mkdir -p /usr/local/cloudcare/dataflux/datakit/embed/linux-amd64 \
    && mkdir -p /usr/local/cloudcare/dataflux/datakit/externals \
    && mkdir -p /opt/oracle

COPY dist/datakit-linux-amd64/datakit /usr/local/cloudcare/dataflux/datakit/datakit
COPY dist/datakit-linux-amd64/externals /usr/local/cloudcare/dataflux/datakit/externals
COPY embed/linux-amd64/agent /usr/local/cloudcare/dataflux/datakit/embed/linux-amd64/agent
COPY iploc.bin /usr/local/cloudcare/dataflux/datakit/data/iploc.bin

RUN sed -i 's/\(archive\|security\).ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list \
    && apt-get update \
    && apt-get install -y libaio-dev libaio1 unzip wget curl

# TODO: we should host the file on OSS
RUN wget -q https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/otn_software/instantclient/instantclient-basiclite-linux.x64-19.8.0.0.0dbru.zip -O /usr/local/cloudcare/dataflux/datakit/externals/instantclient-basiclite-linux.zip \
    && unzip /usr/local/cloudcare/dataflux/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle

ARG dataway=""
ARG dataway_ws_port=""
ARG uuid=""
ARG loglevel=""
ARG enable_inputs=""
ARG global_tags=""
ARG hostname=""

ENV ENV_UUID=$uuid \
    ENV_DATAWAY=$dataway \
    ENV_LOG_LEVEL=$loglevel \
    ENV_ENABLE_INPUTS=$enable_inputs \
    ENV_GLOBAL_TAGS=$global_tags \
		ENV_DATAWAY_WSPORT=$dataway_ws_port \
    ENV_HOSTNAME=$hostname

CMD ["/usr/local/cloudcare/dataflux/datakit/datakit", "-docker"]
