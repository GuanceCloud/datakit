FROM pubrepo.guance.com/datakit-operator/async-profiler:0.6.0-alpine

WORKDIR /app/async-profiler

RUN rm -rf ./build ./profiler.sh

COPY --chmod=0755 profiling.sh ./

CMD ["cron", "-f"]
