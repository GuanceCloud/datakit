# Unless explicitly stated otherwise all files in this repository are licensed
# under the MIT License.
# This product includes software developed at Guance Cloud (https://www.guance.com/).
# Copyright 2021-present Guance, Inc.

FROM eclipse-temurin:11.0.20_8-jdk as asyncprofiler
WORKDIR /opt
COPY ./async-profiler-2.9.2.tar.gz ./
RUN apt-get update && apt-get -y install make gcc g++
RUN tar zxf async-profiler-2.9.2.tar.gz && mv async-profiler-2.9.2 async-profiler && cd async-profiler && make

FROM eclipse-temurin:11.0.20_8-jdk
LABEL authors="guance.com" \
      email="zhangyi905@guance.com"
RUN mkdir -p /app/async-profiler/build && if [ -f "$JAVA_HOME/bin/jps" ] && [ ! -f "/usr/bin/jps" ]; then \
     cp "$JAVA_HOME/bin/jps" /usr/bin/jps; \
    fi
WORKDIR /app/async-profiler
COPY --from=asyncprofiler /opt/async-profiler/build ./build
COPY --from=asyncprofiler /opt/async-profiler/profiler.sh ./
COPY --chmod=0755 profiling.sh ./
COPY --chmod=0755 cmd.sh /usr/bin/
RUN apt-get update && apt-get -y install curl cron
RUN apt-get clean && rm -rf /var/lib/apt/lists
CMD ["cron", "-f"]