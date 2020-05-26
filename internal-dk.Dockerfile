# docker file used to install within k8s
# date: Thu May 14 07:56:59 UTC 2020

from ubuntu:latest

# install dependencies
RUN apt-get update && apt-get install -y \
	libpcap-dev

RUN mkdir -p /usr/local/cloudcare/DataFlux/datakit/embed/linux-amd64
ADD build/datakit-linux-amd64/datakit /usr/local/cloudcare/DataFlux/datakit
ADD embed/linux-amd64/agent           /usr/local/cloudcare/DataFlux/datakit/embed/linux-amd64
CMD /usr/local/cloudcare/DataFlux/datakit/datakit --cfg /usr/local/cloudcare/DataFlux/datakit/datakit.conf
