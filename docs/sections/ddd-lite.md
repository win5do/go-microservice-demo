# DDD Lite
DDD 领域驱动设计的大名大家应该都有所耳闻，但是实际项目完整落地 DDD 的很少。因为 DDD 概念繁杂，领域、子域、核心子域、通用子域、实体、值对象、领域服务、应用服务、领域事件、上下文等一大堆概念，直接把人绕晕，对应到实际业务模型时，横看成岭侧成峰，开发人员内部都难以达成一致。

因为 DDD 设计之初目标是作为复杂软件解决之道，但我们大部分应用并没有那么复杂，一个简单的应用使用这么一套复杂的概念，有点弄巧成拙。在微服务时代，设计原则就是根据领域划分上下文，单体应用复杂度大大降低，微服务需要一种精简的架构。

这里我提出轻量级 DDD 架构： **DDD Lite**，让其更好的契合微服务。

分层架构在 MVC 中就被广泛采用，DDD 中也有四层架构、五层架构、六边形架构等多种流派。 DDD Lite 采用较为简单的四层架构，自上而下为：
- Interface
- Service
- Model
- Infrastructure

## 分层
### Interface
接口层。

对外提供 HTTP，RPC 等接口，参数校验、编解码等逻辑在本层处理，业务状态码、对外数据结构在本层定义。

### Service
服务层。

主要业务逻辑层，调用 Model 层 Repository 接口实现领域业务逻辑，事务组装。一个 Service 对应一个领域，领域是指相似业务逻辑的归类。一个微服务一般只有一到两个 Service，不宜过多。

### Model
数据模型层。

定义数据结构，以及数据模型对应的 Repository 接口，但不包含具体实现。多个相关 Model 组成领域，此处对领域进行简化，不再纠结于子域、实体、值对象等细粒度概念。围绕 Model 定义 Repository 接口，不掺杂业务逻辑，供 Service 层调用，方便在 Service 层组成事务。

#### Repository pattern
这里单独讲一下 Repository 设计模式，这是设计好数据模型的重中之重。

Repository 模式即是将对数据结构的操作抽象成接口，将业务逻辑（business logic）和数据处理（data access）分离，加入了一个抽象层。

有了 Repository 接口，依赖就从具体变成抽象，和 DB 等基础层依赖解耦，方便实施 TDD（测试驱动开发）。

Repository 接口应该保持细粒度，设计得足够通用，减少业务属性，方便复用。

### Infrastructure
基础设施层。

实现 Model 层 Repository 接口，对接 DB、MQ 等数据持久化，或者 RPC 远程调用后端服务。

同时也可以提供 Repository fake 实现，供 Service 层单元测试使用。

## 总结
DDD 中的很多方法论，其中大部分大家在项目工程中都见过或者实践过。但 DDD 要求的前提过于理想，比如什么领域专家，通用语言等前提，在民工式敏捷开发，需求朝令夕改，996加速的大环境下，过于虚幻，难以落地。 所以我们没必要全盘照搬，应该因地制宜，去取精华去其糟粕，提炼出适合自己的应用架构。DDD Lite 对 DDD 设计模式进行了简化，是作者对 DDD 的理论结合实际工程经验的一些总结和思考。

*DDD Lite 架构仅为个人理解，如有高见，欢迎交流。*

## Reference

DDD 分层架构：https://www.yyang.io/2015/12/31/DDD-and-Layered-Architecture/

国外大佬提出的 DDD Lite 架构： https://threedots.tech/post/ddd-lite-in-go-introduction/

国内大佬完整遵循 DDD 设计的 Go demo，个人觉得略显繁琐：https://github.com/agiledragon/ddd-sample-in-golang

