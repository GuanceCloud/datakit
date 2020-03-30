from ubuntu:latest

WORKDIR /usr/local/cloudcare/forethought/datakit
COPY ./datakit .
COPY ./agent embed/
RUN chmod +x embed/agent

ENTRYPOINT [ "./datakit", "--docker" ]