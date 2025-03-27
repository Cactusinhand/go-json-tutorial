# 从零开始的 JSON 库教程（一）：启程 (Go语言版)

## 1. JSON 是什么

JSON（JavaScript Object Notation）是一个用于数据交换的文本格式，现时的标准为[ECMA-404](https://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf)。

虽然 JSON 源至于 JavaScript 语言，但它只是一种数据格式，可用于任何编程语言。现时具类似功能的格式有 XML、YAML，当中以 JSON 的语法最为简单。

JSON 是树状结构，而 JSON 只包含 6 种数据类型：

* null: 表示为 null
* boolean: 表示为 true 或 false
* number: 一般的浮点数表示方式
* string: 表示为 "..."
* array: 表示为 [ ... ]
* object: 表示为 { ... }

我们要实现的 JSON 库，主要是完成 3 个需求：

1. 把 JSON 文本解析为一个树状数据结构（parse）。
2. 提供接口访问该数据结构（access）。
3. 把数据结构转换成 JSON 文本（stringify）。

在本单元中，我们只实现最简单的 null 值解析。

## 2. Go语言项目结构

与C语言不同，Go语言有自己的项目组织方式和包管理系统。我们的JSON库将使用Go模块来组织代码。

我们的 JSON 库名为 leptjson，代码文件有：

1. `leptjson.go`：leptjson 的实现文件，包含类型定义和函数实现。
2. `leptjson_test.go`：测试文件，包含单元测试。

## 3. API 设计

Go语言有自己的类型系统和错误处理机制，我们的API设计会利用这些特性。

首先，我们定义JSON值的类型：

```go
type ValueType int

const (
	NULL ValueType = iota
	FALSE
	TRUE
	NUMBER
	STRING
	ARRAY
	OBJECT
)
```

然后，我们定义解析过程中可能出现的错误：

```go
type ParseError int

const (
	PARSE_OK ParseError = iota
	PARSE_EXPECT_VALUE
	PARSE_INVALID_VALUE
	PARSE_ROOT_NOT_SINGULAR
)
```

接着，我们定义JSON值的数据结构：

```go
type Value struct {
	Type ValueType
}
```

最后，我们提供两个主要的API函数：

```go
func Parse(v *Value, json string) ParseError
func GetType(v *Value) ValueType
```

## 4. JSON 语法子集

下面是此单元的 JSON 语法子集，使用 [ABNF](https://tools.ietf.org/html/rfc5234) 表示：

```
JSON-text = ws value ws
ws = *(%x20 / %x09 / %x0A / %x0D)
value = null 
null  = "null"
```

在这个语法子集下，我们定义了3种错误码：

* 若一个 JSON 只含有空白，返回 `PARSE_EXPECT_VALUE`。
* 若一个值之后，在空白之后还有其他字符，返回 `PARSE_ROOT_NOT_SINGULAR`。
* 若值不是 null，返回 `PARSE_INVALID_VALUE`。

## 5. 单元测试

Go语言内置了测试框架，我们使用它来编写单元测试。测试文件命名为 `xxx_test.go`，测试函数命名为 `TestXxx`。

我们的测试代码包含以下测试函数：

```go
func TestParseNull(t *testing.T)
func TestParseExpectValue(t *testing.T)
func TestParseInvalidValue(t *testing.T)
func TestParseRootNotSingular(t *testing.T)
```

## 6. 实现解析器

我们的解析器实现了以下几个函数：

```go
func parseWhitespace(c *context)
func parseNull(c *context, v *Value) ParseError
func parseValue(c *context, v *Value) ParseError
func Parse(v *Value, json string) ParseError
```

解析过程使用递归下降解析器（recursive descent parser）的方法。从最顶层的 `Parse` 函数开始，它会跳过空白字符，然后调用 `parseValue` 函数解析JSON值。

`parseValue` 函数会根据当前字符决定调用哪个具体的解析函数，例如 `parseNull`。

## 7. 运行测试

在Go语言中，测试是通过内置的`go test`命令来运行的。要运行我们的JSON库测试，可以使用以下命令：

```bash
# 在tutorial01目录下运行所有测试
go test

# 运行测试并显示详细输出
go test -v

# 运行特定的测试函数
go test -run TestParseNull
```

### 运行基准测试

基准测试用于测量代码的性能。要运行基准测试，可以使用以下命令：

```bash
# 运行所有基准测试
go test -bench=.

# 运行特定的基准测试
go test -bench=BenchmarkParseNull

# 运行基准测试并显示内存分配信息
go test -bench=. -benchmem
```

基准测试的输出示例：

```
BenchmarkParseNull-8       10000000               118 ns/op              0 B/op          0 allocs/op
BenchmarkParseTrue-8       10000000               119 ns/op              0 B/op          0 allocs/op
BenchmarkParseFalse-8      10000000               120 ns/op              0 B/op          0 allocs/op
```

输出结果解释：
- 第一列：基准测试名称和CPU核心数
- 第二列：测试运行的次数
- 第三列：每次操作的平均时间（纳秒）
- 第四列：每次操作分配的内存（字节）
- 第五列：每次操作的内存分配次数

## 8. 总结与练习

在本单元中，我们：

1. 设计了JSON库的API
2. 实现了解析null值的功能
3. 编写了单元测试
4. 学习了如何运行测试和基准测试

练习：

1. 完善解析器，增加对true和false的解析
2. 增加相应的单元测试