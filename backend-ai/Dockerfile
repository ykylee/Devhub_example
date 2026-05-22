FROM python:3.12-slim
WORKDIR /app
ENV PYTHONUNBUFFERED=1 \
    PYTHONPATH=/opt/devhub/site-packages
COPY .build/site-packages /opt/devhub/site-packages
COPY main.py .
EXPOSE 8000
CMD ["python", "main.py"]
