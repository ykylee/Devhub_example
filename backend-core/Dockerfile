FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY bin/main ./main
COPY migrations ./migrations
EXPOSE 8080
CMD ["./main"]
