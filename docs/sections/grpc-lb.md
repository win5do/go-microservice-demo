# k8s 服务发现 以及 gRPC 长连接负载均衡

众所周知 gRPC 是基于 HTTP/2，而 HTTP/2 是基于 TCP 长连接的。

k8s 自带一套基于 DNS 的服务发现机制 —— Service。

但基于 Service 的服务发现对于 gRPC  来说并不是开箱即用的，这里面有很多坑。

## 错误姿势 ClusterIP Service

gRPC client 直接使用 ClusterIP Service 会导致负载不均衡。因为 HTTP/2 多个请求可以复用一条连接，并发达到最大值才会创建新的连接。这个最大值由 MaxConcurrentStreams 控制，Golang client 默认是100。

除非运气比较好，例如并发数是 (200, 300]，正好与三个不同 pod 建立了三条长连接。所以使用 ClusterIP Service 不太靠谱。

### 为什么 HTTP/1.1 不受影响

HTTP/1.1 默认开启 Keepalive，也会保持长连接。 但是 HTTP/1.1 多个请求不会共享一个连接，如果连接池里没有空闲连接则会新建一个，经过 Service 的负载均衡，各个 pod 上的连接是相对均衡的。

## 正确姿势 

长连接负载均衡的原理是与后端每个 pod 都建立一个长连接，LB 算法选择一个写入请求。

### gRPC client LB 配合 Headless Service

创建 Headless Service 后，k8s 会生成 DNS 记录，访问 Service 会返回后端多个 pod IP 的 A 记录，这样应用就可以基于 DNS 自定义负载均衡。

在 grpc-client 指定 headless service 地址为 `dns:///` 协议，DNS resolver 会通过 DNS 查询后端多个 pod IP，然后通过 client LB 算法来实现负载均衡。这些 grpc-go 这个库都帮你做了。

```sh
conn, err := grpc.DialContext(ctx, "dns:///"+headlessSvc,
    grpc.WithInsecure(),
    grpc.WithBalancerName(roundrobin.Name),
    grpc.WithBlock(),
)
```

完整代码参考：https://github.com/win5do/go-microservice-demo/tree/main/cmd/client

### Proxy LB 或 ServiceMesh

如果不想在 client 代码中来做 LB，可以使用 Envoy 或 Nginx 反向代理。

Proxy Server 与后端多个 pod 保持长连接，需配置对应模块来识别 HTTP/2 请求，根据 LB 算法选择一个 pod 转发请求。

> Istio 做长连接 LB 不要求 Headless Service，因为网格控制器不会直接使用 Service，而是获取 Service 背后 Endpoint（IP）配置到 Envoy 中。

Proxy 如果自身有多个 replicas，则 proxy 与 client 之间也有长连接的问题，这就套娃了。

但 Istio 等服务网格会为每个 pod 注入一个专用的 Envoy 代理作为 sidecar，所以问题不大。

gRPC 使用 Envoy proxy 可参考 Google Cloud 的教程：

https://github.com/GoogleCloudPlatform/grpc-gke-nlb-tutorial

## 对比

client 和 proxy 两种方式的对比其实就是 `侵入式服务治理 vs 网格化服务治理`。

|      | 侵入式                                                       | 网格化                                                       |
| ---- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 优点 | - 性能好，没有多次转发<br />- 逻辑清晰，开发人员能很快定位问题 | - 基础设施下沉，逻辑一致，方便跨语言异构<br />- 非侵入，应用无感知<br />- 服务治理生态繁荣 |
| 缺点 | - 多种编程语言需要分别开发 client 库<br />- 对接监控告警、链路追踪等基础设施工作量大 | - 链路长，有性能损耗<br />- 下层逻辑复杂，不透明，出问题抓瞎<br />- 还不够成熟 |

## Reference

https://kubernetes.io/blog/2018/11/07/grpc-load-balancing-on-kubernetes-without-tears/

https://zhuanlan.zhihu.com/p/336676373

https://zhuanlan.zhihu.com/p/100200985
