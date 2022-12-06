# Build go
FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -v -o aiko-server -trimpath -ldflags "-s -w" ./cmd/server

# Release
FROM  alpine
RUN  apk --update --no-cache add tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime

COPY --from=builder /app/aiko-server /usr/local/bin
ENTRYPOINT aiko-server -api="$API" -token="$TOKEN" -node="$NODE"
