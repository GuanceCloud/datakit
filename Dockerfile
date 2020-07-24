from ubuntu:18.04 as base

RUN mkdir -p /usr/local/cloudcare/dataflux/datakit
RUN mkdir -p /usr/local/cloudcare/dataflux/datakit/embed/linux-amd64
RUN mkdir -p /usr/local/cloudcare/dataflux/datakit/external

COPY build/datakit-linux-amd64/datakit /usr/local/cloudcare/dataflux/datakit/datakit
COPY build/datakit-linux-amd64/externals /usr/local/cloudcare/dataflux/datakit/externals
COPY embed/linux-amd64/agent /usr/local/cloudcare/dataflux/datakit/embed/linux-amd64/agent

ARG within_docker=1
ARG dataway=""
ARG uuid=""

env ENV_UUID=$uuid
env ENV_DATAWAY=$dataway
env ENV_WITHIN_DOCKER=$within_docker

CMD "/usr/local/cloudcare/dataflux/datakit/datakit"
