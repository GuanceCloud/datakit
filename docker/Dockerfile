from ubuntu:latest

WORKDIR /usr/local/cloudcare/dataflux/datakit
COPY ./datakit .
COPY ./agent embed/
RUN chmod +x embed/agent

ENTRYPOINT [ "./datakit", "--docker" ]