# 第2章：数字解析

在上一章中，我们已经实现了 JSON 库的基本框架和布尔值的解析。本章我们将实现 JSON 中最复杂的数据类型之一 —— 数字的解析。

## 2.1 JSON 数字语法

根据 JSON 标准（RFC 8259），JSON 数字的语法格式为：
```
number = [ "-" ] int [ frac ] [ exp ]
int = "0" / digit1-9 *digit
frac = "." 1*digit
exp = ("e" / "E") ["-" / "+"] 1*digit
```

这个格式可以表示为：
- 可选的负号
- 整数部分
  - 0
  - 非零数字后跟任意数量的数字
- 可选的小数部分（小数点后跟至少一个数字）
- 可选的指数部分（e 或 E 后跟可选的符号和至少一个数字）

举例来说，以下都是有效的 JSON 数字：
- `0`
- `-0`
- `123`
- `-123`
- `123.456`
- `-123.456`
- `1.23e+5`
- `-1.23e-5`
- `1.23E+5`
- `-1.23E-5`

JSON 数字的特点：
- 允许前导负号
- 不允许前导零（`0123`是无效的）
- 允许小数点后没有数字（`1.`是无效的）
- 允许指数表示法
- 不允许十六进制表示法（如`0xABC`）
- 不支持`NaN`和`Infinity`

## 2.2 Go 语言中的数字表示

在 Go 语言中，我们可以使用以下类型来表示数字：
- `int`、`int8`、`int16`、`int32`、`int64` - 带符号整数
- `uint`、`uint8`、`uint16`、`uint32`、`uint64` - 无符号整数
- `float32`、`float64` - 浮点数

对于 JSON 数字，我们将使用 `float64` 来存储，因为：
1. 它可以表示整数和小数
2. 它支持指数表示法
3. 范围足够大，可以满足大多数应用的需求

## 2.3 扩展 Value 结构

首先，我们需要扩展 `Value` 结构来存储数字值：

```go
// Value 表示一个 JSON 值
type Value struct {
    Type ValueType
    // 使用 interface{} 存储不同类型的值
    // 对于数字，存储为 float64
    // 后续会添加字符串、数组和对象类型
    Data interface{}
}
```

## 2.4 实现数字解析

数字解析比布尔值复杂得多，我们需要处理各种情况和边界条件。我们将使用状态机的思想来实现，逐字符解析并转换状态。

首先，让我们扩展 `parseValue` 函数以支持数字解析：

```go
// 解析 JSON 值
func (p *Parser) parseValue(v *Value) ParseError {
    if p.pos >= len(p.json) {
        return ParseExpectValue
    }
    
    switch p.json[p.pos] {
    case 'n':
        return p.parseNull(v)
    case 't':
        return p.parseTrue(v)
    case 'f':
        return p.parseFalse(v)
    case '-':
        return p.parseNumber(v)
    case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
        return p.parseNumber(v)
    default:
        return ParseInvalidValue
    }
}
```

然后，实现 `parseNumber` 函数：

```go
// 解析数字
func (p *Parser) parseNumber(v *Value) ParseError {
    // 记录开始位置
    start := p.pos
    
    // 检查负号
    if p.json[p.pos] == '-' {
        p.pos++
    }
    
    // 检查整数部分
    if p.pos < len(p.json) {
        if p.json[p.pos] == '0' {
            // 0 后面不能直接跟数字
            p.pos++
        } else if isDigit1to9(p.json[p.pos]) {
            p.pos++
            // 读取连续的数字
            for p.pos < len(p.json) && isDigit(p.json[p.pos]) {
                p.pos++
            }
        } else {
            // 无效的整数部分
            return ParseInvalidValue
        }
    } else {
        // 只有一个负号
        return ParseInvalidValue
    }
    
    // 检查小数部分
    if p.pos < len(p.json) && p.json[p.pos] == '.' {
        p.pos++
        
        if p.pos >= len(p.json) || !isDigit(p.json[p.pos]) {
            // 小数点后必须有至少一个数字
            return ParseInvalidValue
        }
        
        // 读取小数部分的数字
        for p.pos < len(p.json) && isDigit(p.json[p.pos]) {
            p.pos++
        }
    }
    
    // 检查指数部分
    if p.pos < len(p.json) && (p.json[p.pos] == 'e' || p.json[p.pos] == 'E') {
        p.pos++
        
        // 指数符号
        if p.pos < len(p.json) && (p.json[p.pos] == '+' || p.json[p.pos] == '-') {
            p.pos++
        }
        
        if p.pos >= len(p.json) || !isDigit(p.json[p.pos]) {
            // 指数部分必须有至少一个数字
            return ParseInvalidValue
        }
        
        // 读取指数部分的数字
        for p.pos < len(p.json) && isDigit(p.json[p.pos]) {
            p.pos++
        }
    }
    
    // 将字符串转换为数字
    numberStr := p.json[start:p.pos]
    number, err := strconv.ParseFloat(numberStr, 64)
    if err != nil {
        return ParseInvalidValue
    }
    
    v.Type = TypeNumber
    v.Data = number
    return ParseOK
}

// 检查字符是否为 1-9 的数字
func isDigit1to9(c byte) bool {
    return c >= '1' && c <= '9'
}

// 检查字符是否为 0-9 的数字
func isDigit(c byte) bool {
    return c >= '0' && c <= '9'
}
```

## 2.5 扩展 API

接下来，我们需要提供访问和设置 JSON 数字值的 API：

```go
// GetNumber 获取 JSON 数字值
func GetNumber(v *Value) float64 {
    // 断言函数，确保值类型正确
    // 在实际应用中可能需要进行类型检查
    if v.Type != TypeNumber {
        return 0.0
    }
    return v.Data.(float64)
}

// SetNumber 设置 JSON 数字值
func SetNumber(v *Value, n float64) {
    // 释放可能的旧数据（如果有必要）
    v.Type = TypeNumber
    v.Data = n
}
```

## 2.6 处理边界情况和错误

在数字解析中，我们需要特别注意以下边界情况：

1. 数字溢出（超出 float64 的表示范围）
2. 前导零问题（如 `01`）
3. 小数点问题（如 `1.` 或 `.1`）
4. 指数问题（如 `1e` 或 `1e+`）

在我们的实现中已经考虑了这些情况，但在一些极端情况下，可能还需要进一步处理。例如，数字 `1e309` 超出了 float64 的表示范围，这会导致 `strconv.ParseFloat` 返回错误。

## 2.7 测试

为了确保我们的实现正确，我们需要编写一系列测试用例：

```go
func TestParseNumber(t *testing.T) {
    testCases := []struct {
        input    string
        expected float64
    }{
        {"0", 0.0},
        {"-0", -0.0},
        {"123", 123.0},
        {"-123", -123.0},
        {"123.456", 123.456},
        {"-123.456", -123.456},
        {"1.23e+5", 123000.0},
        {"-1.23e-5", -0.0000123},
        {"1.23E+5", 123000.0},
        {"-1.23E-5", -0.0000123},
    }
    
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc.input); err != ParseOK {
            t.Errorf("解析 %q 失败，错误码：%v", tc.input, err)
            continue
        }
        if GetType(&v) != TypeNumber {
            t.Errorf("解析 %q 后类型应为 TypeNumber，实际为 %v", tc.input, GetType(&v))
            continue
        }
        actual := GetNumber(&v)
        if math.Abs(actual-tc.expected) > 1e-10 {
            t.Errorf("解析 %q 后值应为 %v，实际为 %v", tc.input, tc.expected, actual)
        }
    }
}

func TestParseInvalidNumber(t *testing.T) {
    testCases := []string{
        "+0",         // 不允许前导加号
        "0123",       // 不允许前导零
        ".123",       // 不允许省略整数部分
        "1.",         // 不允许省略小数部分
        "1e",         // 指数部分必须有数字
        "1e+",        // 指数部分必须有数字
        "1e-",        // 指数部分必须有数字
        "1e+a",       // 指数部分必须是数字
        "1.2.3",      // 不允许多个小数点
        "INF",        // 不支持无穷大
        "NAN",        // 不支持非数字
    }
    
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc); err != ParseInvalidValue {
            t.Errorf("对于无效输入 %q 应该返回 ParseInvalidValue，实际返回 %v", tc, err)
        }
    }
}
```

## 2.8 完整实现

根据以上讨论，我们可以实现完整的数字解析功能。完整的代码会包含更多的错误处理和边界检查，请参考示例代码并确保所有测试通过。

## 2.9 练习

1. 实现对超大数字的处理（超出 float64 范围的数字）。
2. 添加对十六进制数字的支持（如 `0xABC`），这不是 JSON 标准的一部分，但可以作为扩展功能。
3. 优化数字解析的性能，特别是避免使用 `strconv.ParseFloat` 的开销。

## 2.10 下一步

在下一章中，我们将实现字符串的解析，这将涉及到 Unicode 编码和转义序列的处理。

## 参考资源

- [JSON 数字语法](https://www.json.org/json-en.html)
- [RFC 8259：JSON 规范](https://tools.ietf.org/html/rfc8259)
- [IEEE 754 浮点数标准](https://en.wikipedia.org/wiki/IEEE_754)
- [Go 的 strconv 包文档](https://golang.org/pkg/strconv/) 