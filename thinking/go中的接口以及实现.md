其实类似于java,也就是多模态,可以把实现的接口 赋予成 接口的类型.有点抽象,见例子.

比如我们现在有一个接口  
type Animal interface {
    Speak() string
}

然后实现了  Dog和cat

type Dog struct{}
func (d *Dog) Speak() string {
    return "汪汪"
}

type Cat struct{}
func (c *Cat) Speak() string {
    return "喵喵"
}

那么我们在主流程中,并不关注到底实现的是dog还是cat,我们直接

var a Animal

那假如是dog/cat
a = &Dog{}
fmt.Println(a.Speak()) // 输出：汪汪

a = &Cat{}
fmt.Println(a.Speak()) // 输出：喵喵

也就是说，假如我们原先是 a=&Dog{}
后面想用其他的了，只需要改
a=&cat{}

其余主流程代码都一个不用动，因为他们是多模态的接口，实现的功能都是一样的。

拿我们项目里的例子来说

接口声明了两个方法
type Generator interface {
    GenerateNodeReport(...)
    GeneratePodReport(...)
}

我们在生产环境，声明了 DefaultGenerator 这个结构体，然后指针接受者，该结构体实现了这两个方法，那他就实现接口了
type DefaultGenerator struct{}
func (g *DefaultGenerator) GenerateNodeReport(...) {...}
func (g *DefaultGenerator) GeneratePodReport(...) {...}

测试环境也是同理
type MockGenerator struct{}
func (m *MockGenerator) GenerateNodeReport(...) {...}
func (m *MockGenerator) GeneratePodReport(...) {...}


那么我们在主流程中，只需要定义 该 接口类型的 变量
var gen Generator

然后就可以给它赋予不同实现的值
// 生产环境
gen = &DefaultGenerator{}

// 测试环境，mock就代表着测试环境
gen = &MockGenerator{}


实现接口实际上也就是 给某个结构体赋予某功能。也就是招聘需要xx技能， 接口是 技能，结构体是人，只要结构体会这个技能，那他就实现了该技能，就可以入职。岗位并不关心 这个结构体到底是谁，只要招聘 会xx技能的人。