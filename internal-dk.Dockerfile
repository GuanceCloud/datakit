# docker file used to install datakit within k8s
# date: Thu May 14 07:56:59 UTC 2020

from ubuntu:latest

RUN mkdir -p /usr/local/cloudcare/DataFlux/datakit/embed/linux-amd64
ADD build/datakit-linux-amd64/datakit /usr/local/cloudcare/DataFlux/datakit
ADD embed/linux-amd64/agent           /usr/local/cloudcare/DataFlux/datakit/embed/linux-amd64
CMD /usr/local/cloudcare/DataFlux/datakit/datakit
