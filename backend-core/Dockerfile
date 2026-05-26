# 사내 mirror registry 사용 시:
#   docker build --build-arg BACKEND_CORE_BASE=internal-registry.example.com/alpine:3.21 \
#                -f backend-core/Dockerfile -t devhub/backend-core:$TAG backend-core
# 자세한 절차: docs/setup/docker-packaging-deployment-guide.md §5.1.
ARG BACKEND_CORE_BASE=alpine:3.21
FROM ${BACKEND_CORE_BASE}
WORKDIR /app
COPY bin/main ./main
COPY migrations ./migrations
EXPOSE 8080
CMD ["./main"]
