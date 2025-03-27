# 从零开始的 JSON 库教程（四）：解析字符串 (Go语言版)


本文是《从零开始的 JSON 库教程》的第四个单元。本单元的源代码位于 [json-tutorial/go-json-tutorial/tutorial04](https://github.com/miloyip/json-tutorial/blob/master/go-json-tutorial/tutorial04/)。

本单元内容：

1. [JSON 字符串语法](#1-json-字符串语法)
2. [字符串表示](#2-字符串表示)
3. [内存管理](#3-内存管理)
4. [解析字符串](#4-解析字符串)
5. [总结与练习](#5-总结与练习)
6. [参考](#6-参考)

## 1. JSON 字符串语法

JSON 字符串语法比较简单，但有几个转义符（escape）要处理。完整的语法如下：

```
string = quotation-mark *char quotation-mark
char = unescaped /
   escape (
       %x22 /          ; "    quotation mark  U+0022
       %x5C /          ; \    reverse solidus U+005C
       %x2F /          ; /    solidus         U+002F
       %x62 /          ; b    backspace       U+0008
       %x66 /          ; f    form feed       U+000C
       %x6E /          ; n    line feed       U+000A
       %x72 /          ; r    carriage return U+000D
       %x74 /          ; t    tab             U+0009
       %x75 4HEXDIG )  ; uXXXX                U+XXXX
escape = %x5C          ; \
quotation-mark = %x22  ; "
unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
```

简单来说，JSON 字符串是由双引号包围的任意数量字符。字符分为无转义字符或转义序列。转义序列以反斜线开始，紧接着是以下字符：

* `"`: 双引号
* `\`: 反斜线
* `/`: 正斜线
* `b`: 退格符
* `f`: 换页符
* `n`: 换行符
* `r`: 回车符
* `t`: 制表符
* `uXXXX`: 4 位十六进制数字表示的 Unicode 码点

无转义字符是除了双引号、反斜线和控制字符（U+0000 至 U+001F）以外的所有字符。

## 2. 字符串表示

在 Go 语言中，我们可以直接使用内置的 `string` 类型来表示 JSON 字符串。这是因为 Go 的字符串本身就是 UTF-8 编码的，可以很好地表示 Unicode 字符。

我们在 `Value` 结构体中添加了一个 `S` 字段来存储字符串值：

```go
// Value 表示一个JSON值
type Value struct {
    Type   ValueType `json:"type"`   // 值类型
    N      float64   `json:"n"`      // 数字值（当Type为NUMBER时有效）
    S      string    `json:"s"`      // 字符串值（当Type为STRING时有效）
}
```

同时，我们添加了一个 `GetString` 函数来获取字符串值：

```go
// GetString 获取JSON字符串值
func GetString(v *Value) string {
    return v.S
}
```

## 3. 内存管理

Go 语言的字符串是不可变的，当我们需要构建字符串时，通常使用 `[]byte` 切片来累积字符，然后在最后转换为字符串。这种方式既高效又安全，因为：

1. 切片可以动态增长，避免了频繁的内存分配
2. Go 的垃圾回收器会自动管理内存，无需手动释放

在我们的实现中，使用了 `[]byte` 切片来构建字符串：

```go
var sb []byte
// ... 添加字符到 sb ...
*s = string(sb)
```

## 4. 解析字符串

解析 JSON 字符串的主要挑战在于处理各种转义序列，特别是 Unicode 转义序列。我们的实现分为以下几个步骤：

1. 跳过开头的双引号
2. 逐字符解析，直到遇到结束的双引号
3. 处理转义序列
4. 特别处理 Unicode 转义序列，包括代理对（surrogate pairs）

以下是解析字符串的核心代码：

```go
// parseString 解析字符串值
func parseString(c *context, v *Value) ParseError {
    return parseStringRaw(c, &v.S, v)
}

// parseStringRaw 解析字符串，可选择是否设置值
func parseStringRaw(c *context, s *string, v *Value) ParseError {
    var sb []byte
    c.index++ // 跳过开头的引号
    
    for c.index < len(c.json) {
        ch := c.json[c.index]
        if ch == '"' {
            c.index++
            if v != nil {
                v.Type = STRING
            }
            *s = string(sb)
            return PARSE_OK
        } else if ch == '\\' {
            c.index++
            if c.index >= len(c.json) {
                return PARSE_INVALID_STRING_ESCAPE
            }
            
            switch c.json[c.index] {
            case '"', '\\', '/':
                sb = append(sb, c.json[c.index])
            case 'b':
                sb = append(sb, '\b')
            case 'f':
                sb = append(sb, '\f')
            case 'n':
                sb = append(sb, '\n')
            case 'r':
                sb = append(sb, '\r')
            case 't':
                sb = append(sb, '\t')
            case 'u':
                // 解析4位十六进制数字
                // ... 处理Unicode转义序列 ...
            default:
                return PARSE_INVALID_STRING_ESCAPE
            }
        } else if ch < 0x20 {
            // 控制字符必须转义
            return PARSE_INVALID_STRING_CHAR
        } else {
            sb = append(sb, ch)
        }
        c.index++
    }
    
    return PARSE_MISS_QUOTATION_MARK
}
```

### 处理 Unicode 转义序列

Unicode 转义序列的处理是最复杂的部分，特别是当涉及到代理对（surrogate pairs）时。代理对是一种在 UTF-16 编码中表示超出基本多语言平面（BMP）字符的方式。

例如，音乐符号 𝄞 的 Unicode 码点是 U+1D11E，它超出了 BMP 的范围（U+0000 至 U+FFFF）。在 UTF-16 中，它被编码为两个 16 位值：0xD834 和 0xDD1E，分别称为高代理项和低代理项。在 JSON 中，这个字符可以表示为 `"\uD834\uDD1E"`。

我们的实现需要正确处理这种情况，将代理对转换为正确的 Unicode 码点，然后编码为 UTF-8：

```go
// 处理UTF-16代理对
if codepoint >= 0xD800 && codepoint <= 0xDBFF {
    // 高代理项，需要后面跟着低代理项
    if c.index+6 >= len(c.json) || 
        c.json[c.index+1] != '\\' || 
        c.json[c.index+2] != 'u' {
        return PARSE_INVALID_UNICODE_SURROGATE
    }
    
    c.index += 2 // 跳过 \u
    var codepoint2 int
    // ... 解析低代理项 ...
    
    // 检查是否为有效的低代理项
    if codepoint2 < 0xDC00 || codepoint2 > 0xDFFF {
        return PARSE_INVALID_UNICODE_SURROGATE
    }
    
    // 计算Unicode码点
    codepoint = 0x10000 + ((codepoint - 0xD800) << 10) + (codepoint2 - 0xDC00)
}

// 将Unicode码点编码为UTF-8
buf := make([]byte, 4)
n := utf8.EncodeRune(buf, rune(codepoint))
sb = append(sb, buf[:n]...)
```

## 5. 总结与练习

本单元我们实现了 JSON 字符串的解析，包括处理各种转义序列和 Unicode 字符。Go 语言的内置字符串类型和 UTF-8 支持使得实现相对简单，但我们仍然需要小心处理各种边界情况。

练习：

1. 修改 `parseStringRaw` 函数，使用 `strings.Builder` 代替 `[]byte` 切片来构建字符串，比较两种实现的性能差异。
2. 实现一个 `SetString` 函数，允许设置 `Value` 的字符串值。
3. 扩展测试用例，测试更多的 Unicode 字符和边界情况。
4. 实现字符串的生成功能，将 `Value` 中的字符串值转换为 JSON 字符串，包括适当的转义处理。
5. 优化内存使用，考虑使用预分配的缓冲区来减少内存分配次数。

## 6. 参考

1. [RFC 7159 - The JavaScript Object Notation (JSON) Data Interchange Format](https://tools.ietf.org/html/rfc7159)
2. [The Unicode Standard](https://www.unicode.org/standard/standard.html)
3. [UTF-8, a transformation format of ISO 10646](https://tools.ietf.org/html/rfc3629)
4. [UTF-16, an encoding of ISO 10646](https://tools.ietf.org/html/rfc2781)
5. [Go 标准库 - unicode/utf8 包](https://golang.org/pkg/unicode/utf8/)
6. [Go 标准库 - strings 包](https://golang.org/pkg/strings/)

## 7. 运行测试

### 单元测试

在Go语言中，测试是通过内置的`go test`命令来运行的。要运行我们的JSON库测试，可以使用以下命令：

```bash
# 在tutorial04目录下运行所有测试
go test

# 运行测试并显示详细输出
go test -v

# 运行特定的测试函数
go test -run TestParseString
```

### 基准测试

基准测试用于测量代码的性能。要运行基准测试，可以使用以下命令：

```bash
# 运行所有基准测试
go test -bench=.

# 运行特定的基准测试
go test -bench=BenchmarkParseString

# 运行基准测试并显示内存分配信息
go test -bench=. -benchmem
```

基准测试的输出示例：

```
BenchmarkParseNull-8           10000000               118 ns/op              0 B/op          0 allocs/op
BenchmarkParseTrue-8           10000000               119 ns/op              0 B/op          0 allocs/op
BenchmarkParseFalse-8          10000000               120 ns/op              0 B/op          0 allocs/op
BenchmarkParseNumber-8          3000000               435 ns/op              0 B/op          0 allocs/op
BenchmarkParseString-8          2000000               756 ns/op             32 B/op          1 allocs/op
BenchmarkParseUnicodeString-8   1000000              1254 ns/op             48 B/op          2 allocs/op
```

从结果可以看出：

1. 解析字符串比解析基本类型（null、true、false）和数字要慢，这是因为字符串解析涉及更复杂的处理逻辑和内存分配。
2. 解析包含Unicode转义序列的字符串更慢，需要额外的处理步骤。
3. 字符串解析会导致内存分配，这是因为我们需要为字符串内容分配内存。

## 8. 常见问题

1. **为什么使用`[]byte`而不是`strings.Builder`来构建字符串？**

   在Go 1.10之前，`strings.Builder`还不存在，使用`[]byte`是一种常见的字符串构建方式。在现代Go版本中，`strings.Builder`通常是更好的选择，因为它专门为字符串构建优化，并且API更友好。

2. **为什么需要特别处理Unicode代理对？**

   JSON规范基于JavaScript，而JavaScript字符串使用UTF-16编码。在UTF-16中，超出基本多语言平面的字符（U+10000至U+10FFFF）需要使用代理对表示。为了兼容性，JSON也采用了这种表示方式。

3. **Go语言的字符串是UTF-8编码的，为什么还需要处理UTF-16代理对？**

   虽然Go语言内部使用UTF-8编码，但JSON规范要求支持UTF-16代理对。我们需要将JSON中的UTF-16代理对正确转换为UTF-8编码的字符。

4. **为什么控制字符（U+0000至U+001F）必须转义？**

   控制字符在文本中可能导致显示和解析问题，JSON规范要求这些字符必须使用转义序列表示，以确保JSON文本的可读性和安全性。
```