# 编写可测试的 Go 代码

## 测试分类

开发涉及的测试主要有以下三种，其他范围更广的测试如 系统测试、功能测试 这里就不介绍了。

- unit test（单元测试）：针对程序模块来进行正确性检验的测试工作，程序单元是应用的最小可测试部件。单元测试要求没有外部依赖。
- integration test（集成测试）：也叫组装测试或联合测试。 在单元测试的基础上，将所有模块按照设计要求组装成为子系统或系统，进行集成测试。集成测试通常会接入db，mq，后端接口等真实依赖。
- e2e test（端到端测试）：从头到尾验证整个软件及其与外部接口的集成。例如使用 Postman 测试 server 接口。

## 测试前提

- 模块化：逻辑清晰，各司其职
- 细粒度：函数式，关注点在输入输出
- 分层：MVC 边界清晰，而不是带球一条龙
- 抽象：依赖接口而不是具体实现
- 依赖注入：方便替换为 mock 实现

_这些概念看起来比较 Java，但的确是宝贵的工程实践，要学会站在巨人的肩膀上。_

## 测试方法
### mock 接口
对接口进行 mock，只关注接口函数的输入和输出，无需实现细节。

实例：[gomock](https://github.com/golang/mock)

gomock 根据代码中的 interface 接口生产 stub 代码。注入依赖后，使用 EXPECT 提前布局输入和输出，测试调用方行为与预期是否一致。

user.go:
```go
package user

//go:generate mockgen -destination mock_user/mock_user.go github.com/win5do/golang-microservice-demo/docs/sample/gomock/user Index
type Index interface {
	Get(key string) interface{}
}

func GetIndex(in Index, key string) interface{} {
  // business logic
  r := in.Get(key)
  // handle output
  return r
}
```

mock_user.go:
```go
// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/win5do/golang-microservice-demo/docs/sample/gomock/user (interfaces: Index)

// Package mock_user is a generated GoMock package.
package mock_user

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockIndex is a mock of Index interface
type MockIndex struct {
	ctrl     *gomock.Controller
	recorder *MockIndexMockRecorder
}

// MockIndexMockRecorder is the mock recorder for MockIndex
type MockIndexMockRecorder struct {
	mock *MockIndex
}

// NewMockIndex creates a new mock instance
func NewMockIndex(ctrl *gomock.Controller) *MockIndex {
	mock := &MockIndex{ctrl: ctrl}
	mock.recorder = &MockIndexMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIndex) EXPECT() *MockIndexMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockIndex) Get(arg0 string) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockIndexMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockIndex)(nil).Get), arg0)
}
```

user_test.go:
```go
func TestGetIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	index := mock_user.NewMockIndex(ctrl)
	in := "uid"
	out := "username"

	index.EXPECT().Get(in).Return(out)

	r :=  GetIndex(index, "uid")
	require.Equal(t, out, r)
}
```

这是一个最最简单的例子，可能看不出来什么。试想一下，如果 `GetIndex`  逻辑非常复杂，这里 mock 掉依赖调用 `Get` 的输入输出，就能用单测 100% 覆盖 `GetIndex` 代码。

示例代码：https://github.com/win5do/go-microservice-demo/tree/main/docs/sample/gomock/user

特点：很聪明的做法，mock 后调用接口完全透明，准备好输入输出，通过[表格驱动测试](https://feixiao.github.io/testing/table_driven_test.html) ，可以覆盖各种边界值。只能在 test 中使用，每次调用前都需要 mock 输入输出，使用略为繁琐。

### fake 实现

fake 是指模拟真实依赖做一套简化的实现，屏蔽外部系统的依赖。

实例：[client-go fake client](https://github.com/kubernetes/client-go/tree/master/kubernetes/fake)

client-go 依赖于 etcd，其模拟实现 fake-client 包实现了在内存中对资源进行增删改查，与真实依赖 etcd-client 高度一致。

特点：模拟实现编码工作量较大，fake 代码也需要测试来保证正确性，但完善后使用方便。

### 集成测试

直接使用测试环境的真实依赖。现在利用 docker 等容器技术部署依赖非常方便，进行数据初始化后，配置测试代码连接测试 db 运行集成测试。

使用 docker-compose 本地启动 mysql，mongodb，etcd 等依赖可参考我的另一个项目： [db-local](https://github.com/win5do/db-local)。

特点：真实依赖跟线上环境一致性高，更容易发现 bug。但比较笨重，启动慢，测试运行时间长。
