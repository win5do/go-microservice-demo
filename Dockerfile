FROM golang:1.15-alpine AS builder
ENV GO111MODULE=on \
	GOPROXY=https://goproxy.cn
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o main ./cmd/server

FROM alpine:3.12
COPY --from=builder /go/src/app/main /main
RUN chmod +x /main

ENV TZ=Asia/Shanghai
ENTRYPOINT ["/main"]
