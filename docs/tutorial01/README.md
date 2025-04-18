# 第1章：基础结构与布尔值解析

在这一章中，我们将开始构建 JSON 库的基础架构，并实现最简单的 JSON 值 —— 布尔值的解析。

## 1.1 什么是 JSON？

JSON（JavaScript Object Notation）是一种轻量级的数据交换格式。它基于 JavaScript 的对象字面量语法，但是独立于任何编程语言。JSON 具有简单、可读性强的特点，广泛应用于 Web API、配置文件等场景。

JSON 支持以下数据类型：
- **null**：表示空值
- **布尔值**：`true` 或 `false`
- **数字**：整数或浮点数
- **字符串**：用双引号包围的字符序列
- **数组**：有序的值集合
- **对象**：键值对的集合，键必须是字符串

一个简单的 JSON 示例：
```json
{
  "name": "John",
  "age": 30,
  "is_student": false,
  "courses": ["Math", "Computer Science"],
  "address": {
    "city": "New York",
    "zip": "10001"
  }
}
```

## 1.2 Go 语言与 JSON

Go 标准库已经提供了处理 JSON 的包 `encoding/json`，但为了学习目的，我们将从零开始实现自己的 JSON 库。通过这个过程，我们可以深入理解 JSON 解析的原理和 Go 语言的特性。

## 1.3 库的设计目标

我们的 JSON 库将具有以下特点：
- 符合 JSON 标准（[RFC 8259](https://tools.ietf.org/html/rfc8259)）
- 使用递归下降解析器（recursive descent parser）
- 支持 UTF-8 编码的 JSON 文本
- 提供清晰简洁的 API
- 包含完整的错误处理
- 易于扩展

## 1.4 项目结构

首先，我们创建基本的项目结构：

```
tutorial01/
├── json.go     // 主要代码
└── json_test.go // 测试代码
```

## 1.5 定义数据结构

我们首先定义 JSON 值的类型和基本数据结构：

```go
// json.go
package json

// ValueType 表示 JSON 值的类型
type ValueType int

const (
    TypeNull ValueType = iota
    TypeFalse
    TypeTrue
    TypeNumber
    TypeString
    TypeArray
    TypeObject
)

// Value 表示一个 JSON 值
type Value struct {
    Type ValueType
    // 后续将添加其他字段以存储具体的值
}
```

## 1.6 解析器设计

我们的解析器将采用递归下降的方式，逐字符读取 JSON 文本并构建相应的数据结构。解析函数的基本框架如下：

```go
// Parser 表示 JSON 解析器的状态
type Parser struct {
    json string
    pos  int
}

// ParseError 表示解析错误的类型
type ParseError int

const (
    ParseOK ParseError = iota
    ParseExpectValue
    ParseInvalidValue
    // 后续将添加更多错误类型
)

// Parse 解析 JSON 文本并填充到 Value 结构中
func Parse(v *Value, json string) ParseError {
    parser := &Parser{json: json}
    parser.skipWhitespace()
    if err := parser.parseValue(v); err != ParseOK {
        return err
    }
    parser.skipWhitespace()
    if parser.pos < len(parser.json) {
        return ParseRootNotSingular
    }
    return ParseOK
}
```

## 1.7 实现布尔值解析

首先，我们需要实现一些辅助函数：

```go
// 跳过空白字符
func (p *Parser) skipWhitespace() {
    for p.pos < len(p.json) && isWhitespace(p.json[p.pos]) {
        p.pos++
    }
}

// 判断字符是否为空白字符
func isWhitespace(c byte) bool {
    return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
```

然后，实现解析 null 和布尔值的函数：

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
    default:
        return ParseInvalidValue
    }
}

// 解析 null
func (p *Parser) parseNull(v *Value) ParseError {
    if p.pos+3 >= len(p.json) || 
       p.json[p.pos:p.pos+4] != "null" {
        return ParseInvalidValue
    }
    p.pos += 4
    v.Type = TypeNull
    return ParseOK
}

// 解析 true
func (p *Parser) parseTrue(v *Value) ParseError {
    if p.pos+3 >= len(p.json) || 
       p.json[p.pos:p.pos+4] != "true" {
        return ParseInvalidValue
    }
    p.pos += 4
    v.Type = TypeTrue
    return ParseOK
}

// 解析 false
func (p *Parser) parseFalse(v *Value) ParseError {
    if p.pos+4 >= len(p.json) || 
       p.json[p.pos:p.pos+5] != "false" {
        return ParseInvalidValue
    }
    p.pos += 5
    v.Type = TypeFalse
    return ParseOK
}
```

## 1.8 访问 API

为了方便使用解析后的 JSON 值，我们添加一些访问函数：

```go
// GetType 返回 JSON 值的类型
func GetType(v *Value) ValueType {
    return v.Type
}

// IsBoolean 判断 JSON 值是否为布尔值
func IsBoolean(v *Value) bool {
    return v.Type == TypeTrue || v.Type == TypeFalse
}

// GetBoolean 获取 JSON 布尔值
func GetBoolean(v *Value) bool {
    // 断言函数，确保值类型正确
    // 在实际应用中可能需要进行类型检查
    return v.Type == TypeTrue
}

// SetBoolean 设置 JSON 布尔值
func SetBoolean(v *Value, b bool) {
    if b {
        v.Type = TypeTrue
    } else {
        v.Type = TypeFalse
    }
}

// SetNull 设置 JSON null 值
func SetNull(v *Value) {
    v.Type = TypeNull
}
```

## 1.9 测试

测试是确保代码正确性的重要环节。我们使用 Go 的测试框架创建一系列测试用例：

```go
// json_test.go
package json

import "testing"

func TestParseNull(t *testing.T) {
    v := Value{}
    if err := Parse(&v, "null"); err != ParseOK {
        t.Errorf("解析 null 失败，错误码：%v", err)
    }
    if GetType(&v) != TypeNull {
        t.Errorf("期望类型为 null，实际为 %v", GetType(&v))
    }
}

func TestParseTrue(t *testing.T) {
    v := Value{}
    if err := Parse(&v, "true"); err != ParseOK {
        t.Errorf("解析 true 失败，错误码：%v", err)
    }
    if GetType(&v) != TypeTrue {
        t.Errorf("期望类型为 true，实际为 %v", GetType(&v))
    }
    if !GetBoolean(&v) {
        t.Errorf("GetBoolean 应该返回 true")
    }
}

func TestParseFalse(t *testing.T) {
    v := Value{}
    if err := Parse(&v, "false"); err != ParseOK {
        t.Errorf("解析 false 失败，错误码：%v", err)
    }
    if GetType(&v) != TypeFalse {
        t.Errorf("期望类型为 false，实际为 %v", GetType(&v))
    }
    if GetBoolean(&v) {
        t.Errorf("GetBoolean 应该返回 false")
    }
}

func TestParseExpectValue(t *testing.T) {
    testCases := []string{"", " "} 
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc); err != ParseExpectValue {
            t.Errorf("对于空白输入应该返回 ParseExpectValue，实际返回 %v", err)
        }
    }
}

func TestParseInvalidValue(t *testing.T) {
    testCases := []string{"nul", "truee", "falss", "?"}
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc); err != ParseInvalidValue {
            t.Errorf("对于无效输入 %q 应该返回 ParseInvalidValue，实际返回 %v", tc, err)
        }
    }
}

func TestParseRootNotSingular(t *testing.T) {
    testCases := []string{"null x", "true false"}
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc); err != ParseRootNotSingular {
            t.Errorf("对于非单一根值输入 %q 应该返回 ParseRootNotSingular，实际返回 %v", tc, err)
        }
    }
}
```

## 1.10 完整实现

根据以上设计，我们可以实现完整的第一章代码。请参考示例代码并确保所有测试通过。

## 1.11 练习

1. 尝试自己实现 Parse 函数，支持解析 null, true 和 false 值。
2. 增加错误处理，当遇到 `"True"` 或 `"NULL"` 这样的输入时，应该返回错误。
3. 修改代码，使其能够处理 JSON 值前后的空白字符。

## 1.12 下一步

在下一章中，我们将实现数字的解析，这将涉及更复杂的状态转换和错误处理。

## 参考资源

- [JSON 官方网站](https://www.json.org/)
- [RFC 8259：JSON 规范](https://tools.ietf.org/html/rfc8259)
- [Go 编程语言规范](https://golang.org/ref/spec)
- [递归下降解析器介绍](https://en.wikipedia.org/wiki/Recursive_descent_parser) 