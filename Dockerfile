FROM golang:1.15-alpine AS builder
WORKDIR /workspace
ENV GO111MODULE=on \
	GOPROXY=https://goproxy.cn,direct

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# src code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o main ./cmd/server

FROM alpine:3.12
COPY --from=builder /workspace/main /main
RUN chmod +x /main

ENV TZ=Asia/Shanghai
ENTRYPOINT ["/main"]
