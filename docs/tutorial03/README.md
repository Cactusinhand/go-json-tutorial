# 第3章：字符串解析

在前两章中，我们已经实现了解析 JSON 中的布尔值和数字类型。本章我们将实现 JSON 字符串的解析，这是 JSON 中一个更复杂的数据类型，因为它涉及到转义字符和内存管理。

## 3.1 JSON 字符串语法

根据 JSON 标准，JSON 字符串是由双引号包围的零个或多个 Unicode 字符的序列，其中一些字符可以使用反斜杠转义序列表示。

JSON 字符串的语法可以表示为：
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
escape = %x5C              ; \
quotation-mark = %x22      ; "
unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
```

这意味着：
- 字符串必须由双引号包围
- 字符串可以包含以下转义序列：
  - `\"` - 双引号
  - `\\` - 反斜杠
  - `\/` - 正斜杠
  - `\b` - 退格符
  - `\f` - 换页符
  - `\n` - 换行符
  - `\r` - 回车符
  - `\t` - 制表符
  - `\uXXXX` - 4位十六进制表示的 Unicode 字符
- 不能直接包含控制字符（ASCII 码 0-31），必须使用转义序列表示

例如，以下是有效的 JSON 字符串：
- `""`
- `"Hello"`
- `"Hello\nWorld"`
- `"\"Quoted text\""`
- `"\u4E2D\u6587"` (表示 "中文")

## 3.2 Go 中的字符串表示

在 Go 中，字符串是只读的字节序列，通常包含 UTF-8 编码的文本。我们将使用 Go 的 `string` 类型来表示解析后的 JSON 字符串。

需要注意的是，Go 中的字符串和 JSON 字符串在转义处理上有所不同，我们需要在解析时进行相应的转换。

## 3.3 扩展 Value 结构

首先，我们需要扩展 `Value` 结构来支持字符串类型：

```go
// Value 表示一个 JSON 值
type Value struct {
    Type ValueType
    // 对于 TypeString，Data 存储为 string
    Data interface{}
}
```

## 3.4 实现字符串解析

字符串解析的主要挑战在于处理各种转义序列和 Unicode 字符。我们需要逐字符读取并处理转义字符。

首先，扩展 `parseValue` 函数以支持字符串解析：

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
    case '"':
        return p.parseString(v)
    default:
        return ParseInvalidValue
    }
}
```

然后，实现 `parseString` 函数：

```go
// 解析字符串
func (p *Parser) parseString(v *Value) ParseError {
    p.pos++ // 跳过开头的双引号
    start := p.pos
    
    // 用于构建字符串的缓冲区
    var builder strings.Builder
    
    for p.pos < len(p.json) {
        ch := p.json[p.pos]
        
        if ch == '"' {
            // 找到结束的双引号
            if builder.Len() == 0 {
                // 如果没有转义字符，直接使用原始字符串
                v.Type = TypeString
                v.Data = p.json[start:p.pos]
                p.pos++ // 跳过结束的双引号
                return ParseOK
            } else {
                // 如果有转义字符，使用构建的字符串
                v.Type = TypeString
                v.Data = builder.String()
                p.pos++ // 跳过结束的双引号
                return ParseOK
            }
        } else if ch == '\\' {
            // 处理转义字符
            p.pos++
            if p.pos >= len(p.json) {
                return ParseMissQuotationMark
            }
            
            // 将 start 到当前位置的字符添加到 builder
            if p.pos-1 > start {
                builder.WriteString(p.json[start : p.pos-1])
            }
            
            // 处理转义序列
            escape := p.json[p.pos]
            p.pos++
            
            switch escape {
            case '"':
                builder.WriteByte('"')
            case '\\':
                builder.WriteByte('\\')
            case '/':
                builder.WriteByte('/')
            case 'b':
                builder.WriteByte('\b')
            case 'f':
                builder.WriteByte('\f')
            case 'n':
                builder.WriteByte('\n')
            case 'r':
                builder.WriteByte('\r')
            case 't':
                builder.WriteByte('\t')
            case 'u':
                // 处理 Unicode 转义序列
                // 在此先简单实现，下一章将详细处理
                if p.pos+4 > len(p.json) {
                    return ParseInvalidUnicodeHex
                }
                
                // 解析4位十六进制数
                hexStr := p.json[p.pos : p.pos+4]
                p.pos += 4
                
                codePoint, err := strconv.ParseUint(hexStr, 16, 16)
                if err != nil {
                    return ParseInvalidUnicodeHex
                }
                
                // 将 Unicode 码点转换为 UTF-8
                builder.WriteRune(rune(codePoint))
            default:
                return ParseInvalidStringEscape
            }
            
            start = p.pos
        } else if ch < 0x20 {
            // 控制字符必须使用转义序列
            return ParseInvalidStringChar
        } else {
            // 普通字符
            p.pos++
        }
    }
    
    // 如果到达了 JSON 文本的末尾但没有找到结束的双引号
    return ParseMissQuotationMark
}
```

上面的实现有一个重要的优化：对于没有转义字符的简单字符串，我们直接使用原始字符串切片，而不是构建新字符串，这可以减少内存分配和复制。

## 3.5 扩展 API

接下来，我们需要提供访问和设置 JSON 字符串值的 API：

```go
// GetString 获取 JSON 字符串值
func GetString(v *Value) string {
    if v.Type != TypeString {
        return ""
    }
    return v.Data.(string)
}

// SetString 设置 JSON 字符串值
func SetString(v *Value, s string) {
    // 释放可能的旧数据（如果有必要）
    v.Type = TypeString
    v.Data = s
}
```

## 3.6 处理内存管理

在 Go 中，字符串是只读的，并且垃圾收集器会自动管理内存，所以我们不需要手动释放字符串内存。但是，我们仍然需要注意避免不必要的内存分配和复制，特别是对于大型字符串。

## 3.7 测试

为了确保我们的实现正确，我们需要编写一系列测试用例：

```go
func TestParseString(t *testing.T) {
    testCases := []struct {
        input    string
        expected string
    }{
        {`""`, ""},
        {`"Hello"`, "Hello"},
        {`"Hello\nWorld"`, "Hello\nWorld"},
        {`"\"Quoted\""`, "\"Quoted\""},
        {`"\u4E2D\u6587"`, "中文"},
        {`"Hello\\World"`, "Hello\\World"},
        {`"Hello\/World"`, "Hello/World"},
        {`"Hello\bWorld"`, "Hello\bWorld"},
        {`"Hello\fWorld"`, "Hello\fWorld"},
        {`"Hello\rWorld"`, "Hello\rWorld"},
        {`"Hello\tWorld"`, "Hello\tWorld"},
    }
    
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc.input); err != ParseOK {
            t.Errorf("解析 %q 失败，错误码：%v", tc.input, err)
            continue
        }
        if GetType(&v) != TypeString {
            t.Errorf("解析 %q 后类型应为 TypeString，实际为 %v", tc.input, GetType(&v))
            continue
        }
        actual := GetString(&v)
        if actual != tc.expected {
            t.Errorf("解析 %q 后值应为 %q，实际为 %q", tc.input, tc.expected, actual)
        }
    }
}

func TestParseInvalidString(t *testing.T) {
    testCases := []struct {
        input string
        error ParseError
    }{
        // 缺少结束双引号
        {`"`, ParseMissQuotationMark},
        {`"abc`, ParseMissQuotationMark},
        
        // 无效的转义序列
        {`"\v"`, ParseInvalidStringEscape},
        {`"\0"`, ParseInvalidStringEscape},
        {`"\x00"`, ParseInvalidStringEscape},
        
        // 无效的控制字符
        {string([]byte{'"', 0x01, '"'}), ParseInvalidStringChar},
        {string([]byte{'"', 0x1F, '"'}), ParseInvalidStringChar},
        
        // 无效的 Unicode 转义序列
        {`"\u"`, ParseInvalidUnicodeHex},
        {`"\u123"`, ParseInvalidUnicodeHex},
        {`"\uxyz"`, ParseInvalidUnicodeHex},
    }
    
    for _, tc := range testCases {
        v := Value{}
        if err := Parse(&v, tc.input); err != tc.error {
            t.Errorf("对于输入 %q 应该返回 %v，实际返回 %v", tc.input, tc.error, err)
        }
    }
}
```

## 3.8 完整实现

根据以上讨论，我们可以实现完整的字符串解析功能。完整的代码会包含更多的错误处理和边界检查，请参考示例代码并确保所有测试通过。

请注意，我们目前只实现了基本的 Unicode 支持。在下一章中，我们将详细处理 Unicode 字符的解析，包括代理对（surrogate pairs）的处理。

## 3.9 练习

1. 优化字符串解析的性能，尽量减少内存分配和复制。
2. 扩展 `parseString` 函数，支持更多的转义序列（如 `\xhh` 表示十六进制转义）。
3. 实现一个函数，将 Go 字符串转换为 JSON 字符串（包括转义特殊字符）。

## 3.10 下一步

在下一章中，我们将详细处理 JSON 字符串中的 Unicode 字符，包括 UTF-8 编码和代理对的处理。

## 参考资源

- [JSON 字符串语法](https://www.json.org/json-en.html)
- [RFC 8259：JSON 规范](https://tools.ietf.org/html/rfc8259)
- [Go 的 strings 包文档](https://golang.org/pkg/strings/)
- [Unicode 与 UTF-8](https://blog.golang.org/strings) 