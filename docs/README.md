# Golang-microservice-demo

本项目为 Golang 开发的一个微服务 server demo，展示了 grpc、gorm 等常用库的使用，以及 k8s、opentracing 等流行生态的适配，错误处理、Context、Chan 等编码技巧，测试驱动开发的尝试。是作者对自己 Go web
 开发经验的记录和总结，配套文档可点击链接查看。

## Dependencies

- grpc
- [grpc-gateway](./sections/grpc-gateway.md)
- grpc-middleware
- [gorm v2](./sections/gorm.md)
- opentracing / jaeger
- gin
- gomock

## Design

- 文档生成
- [简化的 DDD 架构](./sections/ddd-lite.md)
- [DB 分布式锁](./sections/db-lock.md)
- Repository pattern
- 高可用、横向扩展
- [Error with stackTrace](./sections/error-handle.md)
- Context
- [Unit test / Integrate test](./sections/go-test.md)
- [Deploy with Docker / Kubernetes](./sections/deploy.md)
