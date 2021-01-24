# Golang `github.com/pkg/errors` 包使用的正确姿势

Golang 的 error 不会像 Java 那样打印 stackTrack 信息。回溯 err 非常不方便。

之前见过比较蠢的的做法是层层 log，写起来贼费劲。

大家应该都知道可以通过 `github.com/pkg/errors` 这个包来处理 err，`WithStack(err)` 函数可以打印 stack。

注意，使用 `log.Errorf("%+v", err)` 才会打印 stackTrack，使用 `%v %s` 不行。

但是如果多次使用 `WithStack(err)`，会将 stack 打印多遍，err 信息可能非常长。像这样：
```log
err_test.go:35: err: normal error
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.errMulti
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:29
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.TestOriginWithStack
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:35
    testing.tRunner
        /usr/local/Cellar/go/1.15.5/libexec/src/testing/testing.go:1123
    runtime.goexit
        /usr/local/Cellar/go/1.15.5/libexec/src/runtime/asm_amd64.s:1374
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.errMulti
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:30
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.TestOriginWithStack
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:35
    testing.tRunner
        /usr/local/Cellar/go/1.15.5/libexec/src/testing/testing.go:1123
    runtime.goexit
        /usr/local/Cellar/go/1.15.5/libexec/src/runtime/asm_amd64.s:1374
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.errMulti
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:31
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.TestOriginWithStack
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:35
    testing.tRunner
        /usr/local/Cellar/go/1.15.5/libexec/src/testing/testing.go:1123
    runtime.goexit
        /usr/local/Cellar/go/1.15.5/libexec/src/runtime/asm_amd64.s:1374
```

可以看到这个 stack 信息重复了三次。可以人肉去 check 下层有没有使用 `WithStack(err)`，如果下层用了上层就不用。但这样会增加心智负担，容易出错。

我们可以在调用是使用一个 wrap 函数，判断一下是否已经执行 `WithStack(err)`。

但是 `github.com/pkg/errors` 自定义的 error 类型 `withStack` 是私有类型，如何去判断是否已经执行 `WithStack(err)` 呢？

好在 `StackTrace` 不是私有类型，所以我们可以使用 interface 的一个小技巧，自己定义一个 interface，如果拥有 `StackTrace()` 方法则不再执行 `WithStack(err)`。 像这样：

```sh
type stackTracer interface {
	StackTrace() errors2.StackTrace
}

func WithStackOnce(err error) error {
	if !stackFlag {
		return err
	}

	_, ok := err.(stackTracer)
	if ok {
		return err
	}

	return errors2.WithStack(err)
}
```
有人可能要问 `StackTrace` 也是私有类型咋办？那就 fork 然后直接改源码吧。


现在使用这个 wrap 函数打印出来的 stackTrace 就不会重复和冗长。像这样：
```log
err_test.go:21: err: normal error
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.WithStackOnce
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err.go:27
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.errOnce
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:13
    github.com/win5do/golang-microservice-demo/pkg/lib/errx.TestWithStackOnce
        /Users/wufeng/code/codebase/self/gorm-demo/pkg/lib/errx/err_test.go:19
    testing.tRunner
        /usr/local/Cellar/go/1.15.5/libexec/src/testing/testing.go:1123
    runtime.goexit
        /usr/local/Cellar/go/1.15.5/libexec/src/runtime/asm_amd64.s:1374
```

完整代码参考：https://github.com/win5do/go-microservice-demo/blob/main/pkg/lib/errx/err.go
