FROM golang:1.25.5-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/hls-server ./src/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin .
COPY .upload ./.upload

EXPOSE 8080
CMD ["./hls-server"]

