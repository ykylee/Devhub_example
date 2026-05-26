# 사내 mirror registry 사용 시:
#   docker build --build-arg BACKEND_AI_BASE=internal-registry.example.com/python:3.12-slim \
#                -f backend-ai/Dockerfile -t devhub/backend-ai:$TAG backend-ai
# python minor (3.12) 는 host build 의 ABI 정합으로 고정 — 변경 시 backend-ai/.build/
# site-packages 도 동일 minor 로 재빌드 필요. docs §1.1 host 사전조건 참조.
ARG BACKEND_AI_BASE=python:3.12-slim
FROM ${BACKEND_AI_BASE}
WORKDIR /app
ENV PYTHONUNBUFFERED=1 \
    PYTHONPATH=/opt/devhub/site-packages
COPY .build/site-packages /opt/devhub/site-packages
COPY main.py .
EXPOSE 8000
CMD ["python", "main.py"]
