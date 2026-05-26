# 사내 mirror registry 사용 시:
#   docker build --build-arg FRONTEND_BASE=internal-registry.example.com/node:20-alpine \
#                -f frontend/Dockerfile -t devhub/frontend:$TAG frontend
# node major (20) 는 host build (npm ci + npm run build) 의 환경 정합으로 고정.
# docs §1.1 host 사전조건 참조.
ARG FRONTEND_BASE=node:20-alpine
FROM ${FRONTEND_BASE}
WORKDIR /app
ENV NODE_ENV=production \
    PORT=3000 \
    HOSTNAME=0.0.0.0
COPY .next/standalone ./
COPY .next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
