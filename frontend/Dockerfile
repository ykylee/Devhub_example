FROM node:20-alpine AS builder
WORKDIR /app
ARG BACKEND_API_URL=http://localhost:8080
ENV BACKEND_API_URL=$BACKEND_API_URL
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ARG BACKEND_API_URL=http://localhost:8080
ENV BACKEND_API_URL=$BACKEND_API_URL
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/node_modules ./node_modules
EXPOSE 3000
CMD ["npm", "start"]
