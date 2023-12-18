# Build go
FROM golang:1.21.4-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -v -o Aiko-Server -tags "sing xray with_reality_server with_quic with_grpc with_utls with_wireguard with_acme"

# Release
FROM  alpine
RUN  apk --update --no-cache add tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN mkdir /etc/Aiko-Server/
COPY --from=builder /app/Aiko-Server /usr/local/bin

ENTRYPOINT [ "Aiko-Server"]
