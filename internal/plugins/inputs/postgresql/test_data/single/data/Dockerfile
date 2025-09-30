FROM pubrepo.guance.com/image-repo-for-testing/python:3-alpine-datakit_framework

WORKDIR /app

COPY data/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY data/run_test.py .

CMD ["sh", "-c", "python run_test.py || exit 1"]
