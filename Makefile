.PHONY: run

run:
	go run ./cmd/server --debug


IMG := registry.cn-huhehaote.aliyuncs.com/feng-566/go-microservice-demo:v1.0.0

build:
	docker build -t $(IMG) .

push:
	docker push $(IMG)

image: build push