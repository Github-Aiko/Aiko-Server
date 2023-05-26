# Build go
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod download && \
    go env -w GOFLAGS=-buildvcs=false && \
    go build -v -o Aiko-Server -trimpath -ldflags "-s -w -buildid=" .

# Release
FROM alpine:latest 
RUN apk --update --no-cache add tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && \
    mkdir /etc/Aiko-Server/
COPY --from=builder /app/Aiko-Server /usr/local/bin

ENTRYPOINT [ "Aiko-Server", "--config", "/etc/Aiko-Server/aiko.yml"]