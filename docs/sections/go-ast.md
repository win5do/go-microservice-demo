# 利用 go/ast 语法树做代码生成

## 需求概述

`go.uber.org/zap` 日志包性能很好，但是用起来很不方便，虽然新版本添加了 global 方法，但仍然别扭：`zap.S().Info()`。

现在我们的需求就是将 zap 的 sugaredLogger 封装成一个包，让它像 `logrus` 一样易用，直接调用包内函数：`log.Info()`。

我们只需要找到`SugaredLogger这个 type 拥有的 Exported 方法，将其改为函数，函数体调用其同名方法：

```go
func Info(args ...interface{}) {
	_globalS.Info(args)
}
```

此处 `var _globalS = zap.S()`，因为 `zap.S()` 每次调用都会调用 `RWMutex.RLock() `，改为全局变量提高性能。

这个需求很简单，黏贴复制一顿 replace 就可以搞定，但这太蠢，我们要用一种更 Geek 的方式：**代码生成**。

***完整代码：https://github.com/win5do/go-lib/blob/edc6813f5414f1251e91b670c3a9b89ed89e3525/logx/generator/zap.go***

## 代码实现

要获取某个 type 的方法，大家可能会想到 `reflect` 反射包，但是 reflect 只能知道参数类型，没法知道参数名。所以这里我们使用`go/ast`直接解析源码。

### 获取 ast 语法树

方法可能分散在包内不同 go 文件，所以必须解析整个包，而不是单个文件。

首先要找到 `go.uber.org/zap` 的源码路径，这里我们极客到底，通过 go/build 包获取其在 gomod 中的路径，不用手动填写：

```go
func getImportPkg(pkg string) (string, error) {
	p, err := gobuild.Import(pkg, "", gobuild.FindOnly)
	if err != nil {
		return "", err
	}

	return p.Dir, err
}
```

解析整个 zap 包，拿到 ast 语法树：
```go
func parseDir(dir, pkgName string) (*ast.Package, error) {
	pkgMap, err := goparser.ParseDir(
		token.NewFileSet(),
		dir,
		func(info os.FileInfo) bool {
			// skip go-test
			return !strings.Contains(info.Name(), "_test.go")
		},
		goparser.Mode(0), // no comment
	)
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	pkg, ok := pkgMap[pkgName]
	if !ok {
		err := errors.New("not found")
		return nil, errx.WithStackOnce(err)
	}

	return pkg, nil
}
```

### 遍历并修改 ast

遍历 ast，找到 SugaredLogger 的所有 Exported 方法：
```go
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Recv == nil ||
			!n.Name.IsExported() ||
			len(n.Recv.List) != 1 {
			return nil
		}
		t, ok := n.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			return nil
		}

		if t.X.(*ast.Ident).String() != "SugaredLogger" {
			return nil
		}

		log.Printf("func name: %s", n.Name.String())

		v.funcs = append(v.funcs, rewriteFunc(n))

	}
	return v
}
```


- 将方法 Recv 置空，变为函数。 
- 参数名不变，如果为可变参数，则加上展开符 `...`。
- 函数 body 改为调用全局变量 `_globalS` 的同名方法。 
- 如果有返回值则需要 return 语句。

```go
func rewriteFunc(fn *ast.FuncDecl) *ast.FuncDecl {
	fn.Recv = nil

	fnName := fn.Name.String()

	var args []string
	for _, field := range fn.Type.Params.List {
		for _, id := range field.Names {
			idStr := id.String()
			_, ok := field.Type.(*ast.Ellipsis)
			if ok {
				// Ellipsis args
				idStr += "..."
			}
			args = append(args, idStr)
		}
	}

	exprStr := fmt.Sprintf(`_globalS.%s(%s)`, fnName, strings.Join(args, ","))
	expr, err := goparser.ParseExpr(exprStr)
	if err != nil {
		panic(err)
	}

	var body []ast.Stmt
	if fn.Type.Results != nil {
		body = []ast.Stmt{
			&ast.ReturnStmt{
				// Return:
				Results: []ast.Expr{expr},
			},
		}
	} else {
		body = []ast.Stmt{
			&ast.ExprStmt{
				X: expr,
			},
		}
	}

	fn.Body.List = body

	return fn
}
```

上一步函数返回值中 `zap.SugaredLogger` 在目标包中需要改为 `zap.SugaredLogger`，这里使用 type alias 简单处理一下，当然修改 ast 同样能做到：

```go
// alias
type (
	Logger        = zap.Logger
	SugaredLogger = zap.SugaredLogger
)
```

### ast 转化为 go 代码

单个 func 的 ast 转化为 go 代码，使用 `go/format` 包：

```go
func astToGo(dst *bytes.Buffer, node interface{}) error {
	addNewline := func() {
		err := dst.WriteByte('\n') // add newline
		if err != nil {
			log.Panicln(err)
		}
	}

	addNewline()

	err := format.Node(dst, token.NewFileSet(), node)
	if err != nil {
		return err
	}

	addNewline()

	return nil
}
```

拼装成完整 go file：

```go
func writeGoFile(wr io.Writer, funcs []ast.Decl) error {
	// 输出Go代码
	header := `// Code generated by log-gen. DO NOT EDIT.
package log
`
	buffer := bytes.NewBufferString(header)

	for _, fn := range funcs {
		err := astToGo(buffer, fn)
		if err != nil {
			return errx.WithStackOnce(err)
		}
	}

	_, err := wr.Write(buffer.Bytes())
	return err
}
```

这个程序是输出到了 os.Stdout，通过 `go:generate` 将其重定向到 zap_sugar_generated.go 文件中：

```go
//go:generate sh -c "go run ./generator >zap_sugar_generated.go"
```

大功告成，输出代码示例：

```go
// Code generated by log-gen. DO NOT EDIT.
package log

func Desugar() *Logger {
	return _globalS.Desugar()
}

func Named(name string) *SugaredLogger {
	return _globalS.Named(name)
}

func With(args ...interface{}) *SugaredLogger {
	return _globalS.With(args...)
}

func Debug(args ...interface{}) {
	_globalS.Debug(args...)
}

func Info(args ...interface{}) {
	_globalS.Info(args...)
}

func Warn(args ...interface{}) {
	_globalS.Warn(args...)
}

func Error(args ...interface{}) {
	_globalS.Error(args...)
}

// ......
```

即使之后 zap 包升级了，方法有增改，修改 gomod 版本再次执行 gernerate 即可一键同步，告别手动复粘。

## 总结

Go 没法像 Java 那样做动态 AOP，但可以通过 go/ast 做代码生成，达成同样目标，而且不像 reflect 会影响性能和静态检查。用的好的话可以极大提高效率，更加自动化，减少手工复粘，也就降低犯错概率。

已在很多明星开源项目里广泛应用，如：

- 代码编辑工具 gomodifytags： https://github.com/fatih/gomodifytags、

- Go 编译时依赖注入 Wire： https://github.com/google/wire

- K8S 源码：https://github.com/kubernetes/code-generator

## Reference

https://github.com/kubernetes/gengo

https://juejin.cn/post/6844903982683389960
