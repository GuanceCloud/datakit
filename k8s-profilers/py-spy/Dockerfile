FROM python:3.10.11-bullseye as py-spy-for-datakit
RUN pip install py-spy-for-datakit -i https://mirrors.aliyun.com/pypi/simple/
RUN which py-spy-for-datakit

FROM ubuntu:22.04
COPY --from=py-spy-for-datakit /usr/local/bin/py-spy-for-datakit /usr/bin/py-spy-for-datakit
WORKDIR /app/py-spy
RUN apt-get update && apt-get -y install cron
COPY --chmod=0755 profiling.sh ./
COPY --chmod=0755 cmd.sh /usr/bin/
RUN apt-get clean && rm -rf /var/lib/apt/lists
CMD ["cron", "-f"]
