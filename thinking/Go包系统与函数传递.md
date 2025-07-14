# Go包系统与函数传递

## Go中的包系统

### 包声明与导入
1. **包声明的作用**
   - `package` 声明当前文件属于哪个包
   - 同一目录下所有文件必须使用相同的包名
   - 组织代码、控制可见性和提供命名空间隔离

2. **导入机制**
   - 使用 `import` 导入其他包
   - 导入路径是完整的模块路径+包目录
   - 例如: `import "github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"`

3. **包名与导入路径的关系**
   - 包名通常是包含该包的目录名称
   - 导入后使用包名来引用其中的标识符: `rules.NewEngine()`
   - 不能用 `package rules` 来导入，`package` 只能声明当前文件的包名

### Go的可见性规则
- 首字母大写的标识符会被导出(公开)，可供包外访问
- 首字母小写的标识符仅包内可见
- 包内可以访问该包内所有标识符，无论首字母大小写

## Go中函数传递的方法

1. **函数作为参数**
```go
func process(f func(int) int, value int) {
    result := f(value)
    fmt.Println(result)
}
```

2. **函数类型定义**
```go
type Handler func(string) error

func executeHandler(h Handler, data string) {
    h(data)
}
```

3. **方法值和方法表达式**
```go
type Counter struct {
    count int
}
func (c *Counter) Increment() { c.count++ }

// 方法值
counter := &Counter{}
inc := counter.Increment  // 方法值
inc()  // 调用

// 方法表达式
inc := (*Counter).Increment
inc(counter)  // 需要显式传接收者
```

4. **闭包(匿名函数)**
```go
func createAdder(base int) func(int) int {
    return func(x int) int {
        return base + x
    }
}
```

## 关于项目中的 rules.NewEngine 函数

在K8s-Resource-Inspector项目中，我们看到类似 `rulesEngine, err := rules.NewEngine(rulesPath)` 的代码。这里的工作机制是:

1. `rules` 包定义在 `code/internal/rules/` 目录下
2. `NewEngine` 是在 `engine.go` 文件中定义的函数，首字母大写表示它是导出的
3. 当导入 `rules` 包后，`NewEngine` 函数自动成为该包的公共API
4. 这个函数不需要被"接收"，它是 `rules` 包的构造函数，用于创建 `Engine` 类型的实例

这是Go语言的常见模式：在包中定义类型（如`Engine`）和创建该类型实例的函数（通常以`New`开头）。

## Go语言的隐式约定

Go语言有许多隐式约定，这些约定使代码简洁明了，但需要熟悉才能完全理解:

- 首字母大写的标识符自动导出
- 没有显式继承但有组合和接口
- 接口的隐式实现(只要实现了接口所有方法即视为实现了接口)
- 自动调用的特殊方法(如`String()`方法用于格式化输出) 