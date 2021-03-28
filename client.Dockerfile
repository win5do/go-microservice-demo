FROM golang:1.15-alpine AS builder
WORKDIR /workspace
ENV GO111MODULE=on \
	GOPROXY=https://goproxy.cn
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o main ./cmd/client

FROM alpine:3.12
COPY --from=builder /workspace/main /main
RUN chmod +x /main

ENV TZ=Asia/Shanghai
ENTRYPOINT ["/main"]
