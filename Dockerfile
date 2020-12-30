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
    && apt-get install -y libaio-dev libaio1 unzip wget

# TODO: we should host the file on OSS
RUN wget -q https://download.oracle.com/otn_software/linux/instantclient/19800/instantclient-basiclite-linux.x64-19.8.0.0.0dbru.zip?xd_co_f=6a6ddc80-4750-4aca-bd5f-ffd0b3fbd9aa -O /usr/local/cloudcare/dataflux/datakit/externals/instantclient-basiclite-linux.zip \
    && unzip /usr/local/cloudcare/dataflux/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle

ARG dataway=""
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
    ENV_HOSTNAME=$hostname

CMD ["/usr/local/cloudcare/dataflux/datakit/datakit", "-docker"]