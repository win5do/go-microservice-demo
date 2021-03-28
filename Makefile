.PHONY: run

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

run-server:
	go run ./cmd/server --debug

run-client:
	go run ./cmd/client --service localhost:9020

IMG_SERVER := registry.cn-huhehaote.aliyuncs.com/feng-566/grpc-server:v1.0.0
IMG_CLIENT := registry.cn-huhehaote.aliyuncs.com/feng-566/grpc-client:v1.0.0


build-server: fmt vet
	docker build -t $(IMG_SERVER) . &&\
	docker push $(IMG_SERVER)


build-client: fmt vet
	docker build -f client.Dockerfile -t $(IMG_CLIENT) . &&\
	docker push $(IMG_CLIENT)

image: build-server build-client
