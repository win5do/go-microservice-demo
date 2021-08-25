.PHONY: run

# Run go fmt against code
fmt:
	go fmt ./...

lint:
	golangci-lint run -v

staticcheck: fmt lint

run-server:
	go run -race ./cmd/server --debug

run-client:
	go run  -race ./cmd/client --service localhost:9020

IMG_SERVER := registry.cn-huhehaote.aliyuncs.com/feng-566/grpc-server:v1.0.0
IMG_CLIENT := registry.cn-huhehaote.aliyuncs.com/feng-566/grpc-client:v1.0.0


build-server: check
	docker build -t $(IMG_SERVER) . &&\
	docker push $(IMG_SERVER)


build-client: check
	docker build -f client.Dockerfile -t $(IMG_CLIENT) . &&\
	docker push $(IMG_CLIENT)

image: build-server build-client
