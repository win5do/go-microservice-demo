# Deploy

## Docker build

Golang 可以直接编译成二进制文件，可以使用两段式构建减少容器体积。

Dockfile：https://github.com/win5do/go-microservice-demo/blob/main/Dockerfile

使用 Makefile 运行镜像构建命令：
```sh
make image
```

## K8s yaml

Deployment 运行三节点高可用服务，Service 暴露服务端口。

yaml：https://github.com/win5do/go-microservice-demo/tree/main/deploy

部署：
```sh
kubectl apply -f ./deploy
```
