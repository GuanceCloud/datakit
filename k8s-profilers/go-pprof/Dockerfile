FROM golang:1.19-bullseye as builder
RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go install github.com/guancecloud/pprof@latest
RUN cp -f $(which pprof) /usr/local/bin


FROM ubuntu:22.04
LABEL authors="guance.com" email="zhangyi905@guance.com"
WORKDIR /app/go-pprof
COPY --from=builder /usr/local/bin/pprof /bin/
RUN apt-get update && apt-get -y install ca-certificates && sed -i.backup 's/http:\/\/deb.debian.org/https:\/\/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && apt-get -y install curl procps net-tools cron iproute2
COPY --chmod=0755 profiling.sh ./
COPY --chmod=0755 cmd.sh /usr/bin/
RUN apt-get clean && rm -rf /var/lib/apt/lists
CMD ["cron", "-f"]