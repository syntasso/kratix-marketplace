FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY approval-gate.py /app/approval-gate.py

ENTRYPOINT ["python", "/app/approval-gate.py"]
