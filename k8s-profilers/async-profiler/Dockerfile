FROM eclipse-temurin:11.0.20_8-jdk as asyncprofiler
WORKDIR /opt
COPY ./async-profiler-2.9.2.tar.gz ./
RUN apt-get update && apt-get -y install make gcc g++
RUN tar zxf async-profiler-2.9.2.tar.gz && mv async-profiler-2.9.2 async-profiler && cd async-profiler && make

FROM eclipse-temurin:11.0.20_8-jdk
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
