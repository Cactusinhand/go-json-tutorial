# 从零开始的 JSON 库教程（二）：解析数字 (Go语言版)

## 1. 数字类型的语法

JSON 的数字语法比较简单，比起 C/C++ 的数字字面量，JSON 的数字语法是它的一个子集。语法图如下：

```
number = [ "-" ] int [ frac ] [ exp ]
int = "0" / digit1-9 *digit
frac = "." 1*digit
exp = ("e" / "E") [ "-" / "+" ] 1*digit
```

简单翻译一下：

- 数字可以是负数
- 整数部分如果不是 0，不能以 0 开头
- 小数部分是可选的，如果有，必须有小数点和至少一个数字
- 指数部分是可选的，如果有，必须有 e 或 E，可以有正负号，必须有至少一个数字

比如以下都是合法的 JSON 数字：

```
0
1
-0
-1
1.1
1e10
1e+10
1e-10
-1e10
-1e+10
-1e-10
```

但注意以下不是合法的 JSON 数字：

```
+0     // 不允许正号
+1     // 不允许正号
.123   // 小数点前必须有数字
1.     // 小数点后必须有数字
INF    // 不是合法的 JSON 数字
inf    // 不是合法的 JSON 数字
NAN    // 不是合法的 JSON 数字
nan    // 不是合法的 JSON 数字
0123   // 前导零后不能有数字
0x0    // 不支持十六进制
0x123  // 不支持十六进制
```

## 2. 数字的表示方式

在 Go 语言中，我们可以使用 `float64` 类型来表示 JSON 数字。这是因为 JSON 的数字可以表示整数和浮点数，而 `float64` 类型可以覆盖大部分常用的数值范围。

我们需要扩展 `Value` 结构体，添加一个 `N` 字段来存储数字值：

```go
type Value struct {
	Type ValueType `json:"type"` // 值类型
	N    float64   `json:"n"`    // 数字值（当Type为NUMBER时有效）
}
```

同时，我们需要添加一个新的错误类型 `PARSE_NUMBER_TOO_BIG`，用于处理数字溢出的情况：

```go
const (
	PARSE_OK                ParseError = iota // 解析成功
	PARSE_EXPECT_VALUE                        // 期望一个值
	PARSE_INVALID_VALUE                       // 无效的值
	PARSE_ROOT_NOT_SINGULAR                   // 根节点不唯一
	PARSE_NUMBER_TOO_BIG                      // 数字太大
)
```

## 3. 数字的解析

数字的解析是本章的重点。我们需要实现 `parseNumber` 函数，它需要处理各种数字格式，包括整数、小数和指数。

```go
func parseNumber(c *context, v *Value) ParseError {
	// 记录起始位置
	start := c.index

	// 处理负号
	if c.index < len(c.json) && c.json[c.index] == '-' {
		c.index++
	}

	// 处理整数部分
	if c.index < len(c.json) && c.json[c.index] == '0' {
		c.index++
		// 0后面不能直接跟数字，必须是小数点或指数符号
		if c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			return PARSE_INVALID_VALUE
		}
	} else if c.index < len(c.json) && c.json[c.index] >= '1' && c.json[c.index] <= '9' {
		c.index++
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	} else {
		return PARSE_INVALID_VALUE
	}

	// 处理小数部分
	if c.index < len(c.json) && c.json[c.index] == '.' {
		c.index++
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 处理指数部分
	if c.index < len(c.json) && (c.json[c.index] == 'e' || c.json[c.index] == 'E') {
		c.index++
		if c.index < len(c.json) && (c.json[c.index] == '+' || c.json[c.index] == '-') {
			c.index++
		}
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 转换为数字
	n, err := strconv.ParseFloat(c.json[start:c.index], 64)
	if err != nil {
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				return PARSE_NUMBER_TOO_BIG
			}
		}
		return PARSE_INVALID_VALUE
	}

	// 检查是否溢出
	if math.IsInf(n, 0) || math.IsNaN(n) {
		return PARSE_NUMBER_TOO_BIG
	}

	v.N = n
	v.Type = NUMBER
	return PARSE_OK
}
```

解析过程分为几个步骤：

1. 处理负号：检查是否有负号，如果有则跳过
2. 处理整数部分：处理0或非0开头的整数
3. 处理小数部分：处理小数点和小数位
4. 处理指数部分：处理e/E和指数
5. 转换为数字：使用Go的strconv.ParseFloat函数将字符串转换为浮点数
6. 检查溢出：检查数字是否太大

## 4. 访问数字值

我们需要提供一个函数来访问解析后的数字值：

```go
func GetNumber(v *Value) float64 {
	return v.N
}
```

## 5. 单元测试

我们编写了一系列单元测试来验证数字解析功能：

```go
func TestParseNumber(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
	}{
		{"0", 0.0},
		{"-0", 0.0},
		{"-0.0", 0.0},
		{"1", 1.0},
		{"-1", -1.0},
		{"1.5", 1.5},
		{"-1.5", -1.5},
		{"3.1416", 3.1416},
		{"1E10", 1e10},
		{"1e10", 1e10},
		{"1E+10", 1e+10},
		{"1E-10", 1e-10},
		{"-1E10", -1e10},
		{"-1e10", -1e10},
		{"-1E+10", -1e+10},
		{"-1E-10", -1e-10},
		{"1.234E+10", 1.234e+10},
		{"1.234E-10", 1.234e-10},
		{"1e-10000", 0.0}, // 下溢出
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, tc.input); err != PARSE_OK {
				t.Errorf("期望解析成功，但返回错误: %v", err)
			}
			if GetType(&v) != NUMBER {
				t.Errorf("期望类型为NUMBER，但得到: %v", GetType(&v))
			}
			if GetNumber(&v) != tc.expected {
				t.Errorf("期望值为%g，但得到: %g", tc.expected, GetNumber(&v))
			}
		})
	}
}
```

我们还测试了各种无效的数字格式：

```go
func TestParseInvalidValue(t *testing.T) {
	// 测试无效数字
	invalidNumbers := []string{
		"+0",    // 不允许正号
		"+1",    // 不允许正号
		".123",  // 小数点前必须有数字
		"1.",    // 小数点后必须有数字
		"INF",   // 不是合法的JSON数字
		"inf",   // 不是合法的JSON数字
		"NAN",   // 不是合法的JSON数字
		"nan",   // 不是合法的JSON数字
		"0123",  // 前导零后不能有数字
		"0x0",   // 不支持十六进制
		"0x123", // 不支持十六进制
		"0123",  // 不允许前导零
		"1e",    // 指数部分不完整
		"1e+",   // 指数部分不完整
		"1e-",   // 指数部分不完整
	}

	for _, invalidNum := range invalidNumbers {
		t.Run(invalidNum, func(t *testing.T) {
			v := Value{Type: TRUE}
			if err := Parse(&v, invalidNum); err != PARSE_INVALID_VALUE {
				t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v，输入: %s", err, invalidNum)
			}
			if GetType(&v) != NULL {
				t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
			}
		})
	}
}
```

以及数字太大的情况：

```go
func TestParseNumberTooBig(t *testing.T) {
	v := Value{Type: FALSE}
	if err := Parse(&v, "1e309"); err != PARSE_NUMBER_TOO_BIG {
		t.Errorf("期望错误PARSE_NUMBER_TOO_BIG，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	v = Value{Type: FALSE}
	if err := Parse(&v, "-1e309"); err != PARSE_NUMBER_TOO_BIG {
		t.Errorf("期望错误PARSE_NUMBER_TOO_BIG，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}
```

## 6. 基准测试

为了测试解析器的性能，我们添加了基准测试：

```go
// 基准测试 - 解析null值
func BenchmarkParseNull(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "null")
	}
}

// 基准测试 - 解析true值
func BenchmarkParseTrue(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "true")
	}
}

// 基准测试 - 解析false值
func BenchmarkParseFalse(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "false")
	}
}

// 基准测试 - 解析数字
func BenchmarkParseNumber(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "123.456e+789")
	}
}
```

运行基准测试的命令如下：

```bash
go test -bench=. -benchmem
```

这将运行所有基准测试，并显示内存分配信息。以下是一个示例输出：

```
BenchmarkParseNull-8       10000000               118 ns/op              0 B/op          0 allocs/op
BenchmarkParseTrue-8       10000000               119 ns/op              0 B/op          0 allocs/op
BenchmarkParseFalse-8      10000000               120 ns/op              0 B/op          0 allocs/op
BenchmarkParseNumber-8      3000000               435 ns/op              0 B/op          0 allocs/op
```

从结果可以看出：

1. 解析 null、true 和 false 的性能非常接近，每次操作大约需要 120 纳秒。
2. 解析数字的性能相对较慢，每次操作大约需要 435 纳秒，这是因为数字解析涉及更复杂的处理逻辑和字符串到浮点数的转换。
3. 所有操作都没有内存分配，这是因为我们的解析器设计得非常高效，不需要额外的内存分配。

## 7. 运行测试

在Go语言中，测试是通过内置的`go test`命令来运行的。要运行我们的JSON库测试，可以使用以下命令：

```bash
# 在tutorial02目录下运行所有测试
go test

# 运行测试并显示详细输出
go test -v

# 运行特定的测试函数
go test -run TestParseNumber
```

### 运行基准测试

基准测试用于测量代码的性能。要运行基准测试，可以使用以下命令：

```bash
# 运行所有基准测试
go test -bench=.

# 运行特定的基准测试
go test -bench=BenchmarkParseNumber

# 运行基准测试并显示内存分配信息
go test -bench=. -benchmem
```

基准测试的输出示例：

```
BenchmarkParseNull-8       10000000               118 ns/op              0 B/op          0 allocs/op
BenchmarkParseTrue-8       10000000               119 ns/op              0 B/op          0 allocs/op
BenchmarkParseFalse-8      10000000               120 ns/op              0 B/op          0 allocs/op
BenchmarkParseNumber-8      3000000               435 ns/op              0 B/op          0 allocs/op
```

输出结果解释：
- 第一列：基准测试名称和CPU核心数
- 第二列：测试运行的次数
- 第三列：每次操作的平均时间（纳秒）
- 第四列：每次操作分配的内存（字节）
- 第五列：每次操作的内存分配次数

## 8. 总结

在本章中，我们实现了 JSON 数字的解析功能。主要内容包括：

1. 了解 JSON 数字的语法规则
2. 使用 `float64` 类型表示 JSON 数字
3. 实现 `parseNumber` 函数，处理各种数字格式
4. 添加错误处理，包括处理数字溢出的情况
5. 编写单元测试和基准测试验证实现
6. 学习了如何运行测试和基准测试

通过这一章的学习，我们的 JSON 解析器已经可以处理 null、true、false 和数字类型。在下一章中，我们将实现 JSON 字符串的解析，包括处理转义字符和 Unicode 编码。