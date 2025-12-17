FROM pubrepo.guance.com/image-repo-for-testing/python:3-alpine-datakit_framework

WORKDIR /app

RUN pip install --no-cache-dir pymongo

COPY . /app

ENV MONGO_HOST=mongodb
ENV MONGO_PORT=27017
ENV MONGO_USER=admin
ENV MONGO_PASSWORD=admin
ENV MONGO_DATABASE=test_db
ENV NUM_THREADS=5
ENV RUN_DURATION=300

CMD ["python", "run_test.py"]