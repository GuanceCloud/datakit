FROM ubuntu:latest AS base

RUN mkdir -p /usr/local/cloudcare/dataflux/datakit
RUN mkdir -p /usr/local/cloudcare/dataflux/datakit/embed/linux-amd64
RUN mkdir -p /usr/local/cloudcare/dataflux/datakit/externals
RUN mkdir -p /opt/oracle

COPY build/datakit-linux-amd64/datakit /usr/local/cloudcare/dataflux/datakit/datakit
COPY build/datakit-linux-amd64/externals /usr/local/cloudcare/dataflux/datakit/externals
COPY embed/linux-amd64/agent /usr/local/cloudcare/dataflux/datakit/embed/linux-amd64/agent
COPY plugins/externals/oraclemonitor/instantclient-basic-linux.x64-19.6.0.0.0dbru.zip /usr/local/cloudcare/dataflux/datakit/externals/instantclient-basic-linux.x64-19.6.0.0.0dbru.zip

RUN apt-get update
#RUN apt-get install -y libaio-dev libaio1 unzip vim
RUN apt-get install -y libaio-dev libaio1 unzip
RUN unzip /usr/local/cloudcare/dataflux/datakit/externals/instantclient-basic-linux.x64-19.6.0.0.0dbru.zip -d /opt/oracle

ARG dataway=""
ARG uuid=""
ARG loglevel=""
ARG enable_inputs=""
ARG global_tags=""
ARG hostname=""

env ENV_UUID=$uuid
env ENV_DATAWAY=$dataway
env ENV_LOG_LEVEL=$loglevel
env ENV_ENABLE_INPUTS=$enable_inputs
env ENV_GLOBAL_TAGS=$global_tags
env ENV_HOSTNAME=$hostname

CMD ["/usr/local/cloudcare/dataflux/datakit/datakit", "-docker"]
