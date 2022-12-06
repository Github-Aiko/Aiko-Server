# Build go
FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -v -o server-trojan -trimpath -ldflags "-s -w" ./cmd/server

# Release
FROM  alpine
# 安装必要的工具包
RUN  apk --update --no-cache add tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /app/server-trojan /usr/local/bin
ENTRYPOINT server-trojan -api="$API" -token="$TOKEN" -node="$NODE"
