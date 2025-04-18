# 第4章：Unicode 支持

在上一章中，我们实现了基本的 JSON 字符串解析。本章我们将深入 Unicode 的世界，完善对 Unicode 字符的支持，特别是处理转义序列 `\uXXXX` 和 UTF-8 编码。

## 4.1 Unicode 简介

Unicode 是一个国际标准，为世界上几乎所有的文字系统中的每个字符分配了唯一的编码，使计算机能够一致地表示和处理文本。每个 Unicode 字符都有一个唯一的编号，称为代码点（code point），通常表示为 U+XXXX 格式，其中 XXXX 是十六进制数。

例如：
- U+0041 表示字符 'A'
- U+4E2D 表示汉字 '中'
- U+1F600 表示emoji 😀

### Unicode 编码方式

Unicode 可以通过多种编码方式存储，最常见的是：

1. **UTF-8**：使用 1 到 4 个字节编码一个 Unicode 字符。ASCII 字符（U+0000 到 U+007F）只需要 1 个字节，其他字符需要 2-4 个字节。
2. **UTF-16**：使用 2 或 4 个字节编码一个 Unicode 字符。
3. **UTF-32**：使用固定的 4 个字节编码每个 Unicode 字符。

Go 语言内部使用 UTF-8 编码表示字符串，这也是 JSON 规范推荐的编码方式。

### Unicode 代理对

Unicode 标准最初设计为 16 位编码，可以表示 65536 个字符（U+0000 到 U+FFFF）。但随着更多字符的加入，Unicode 扩展到了 21 位（U+0000 到 U+10FFFF）。

为了在 UTF-16 编码中表示超过 U+FFFF 的字符，引入了代理对（surrogate pairs）的概念。代理对使用两个 16 位码元（code unit）组合表示一个高于 U+FFFF 的 Unicode 字符：

- 高代理（high surrogate）：范围为 U+D800 到 U+DBFF
- 低代理（low surrogate）：范围为 U+DC00 到 U+DFFF

当 JSON 中使用 `\uXXXX` 表示字符时，如果遇到代理对，需要将它们正确解析为对应的 Unicode 字符。

## 4.2 JSON 中的 Unicode

在 JSON 中，Unicode 字符可以通过以下方式表示：

1. 直接使用 UTF-8 编码的字符
2. 使用 `\uXXXX` 转义序列

例如，汉字"中国"可以表示为：
- 直接在 JSON 中使用 UTF-8 编码："中国"
- 使用转义序列："\u4E2D\u56FD"

JSON 规范要求解析器能够处理所有有效的 Unicode 字符，包括通过代理对表示的字符。例如，emoji 😀 (U+1F600) 在 JSON 中可以表示为 "\uD83D\uDE00"。

## 4.3 UTF-8 编码规则

UTF-8 是一种变长编码，使用 1 到 4 个字节表示一个 Unicode 字符：

| Unicode 范围         | UTF-8 编码格式                  | 字节数 |
|---------------------|--------------------------------|-------|
| U+0000 - U+007F     | 0xxxxxxx                       | 1     |
| U+0080 - U+07FF     | 110xxxxx 10xxxxxx              | 2     |
| U+0800 - U+FFFF     | 1110xxxx 10xxxxxx 10xxxxxx     | 3     |
| U+10000 - U+10FFFF  | 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx | 4 |

例如，汉字"中"的 Unicode 代码点是 U+4E2D，它的 UTF-8 编码为 E4 B8 AD（二进制表示为 11100100 10111000 10101101）。

## 4.4 实现 Unicode 支持

现在我们来修改我们的 JSON 库，增加 Unicode 支持。主要工作包括：

1. 处理 `\uXXXX` 转义序列
2. 处理代理对
3. 将解析的 Unicode 字符转换为 UTF-8 编码

### 4.4.1 修改 leptjson.h

首先，我们需要添加新的错误码，用于表示 Unicode 相关的错误：

```c
typedef enum {
    LEPT_PARSE_OK,
    LEPT_PARSE_EXPECT_VALUE,
    LEPT_PARSE_INVALID_VALUE,
    LEPT_PARSE_ROOT_NOT_SINGULAR,
    LEPT_PARSE_NUMBER_TOO_BIG,
    LEPT_PARSE_MISS_QUOTATION_MARK,
    LEPT_PARSE_INVALID_STRING_ESCAPE,
    LEPT_PARSE_INVALID_STRING_CHAR,
    LEPT_PARSE_INVALID_UNICODE_HEX,
    LEPT_PARSE_INVALID_UNICODE_SURROGATE
} lept_parse_error;
```

我们添加了两个新的错误类型：
- `LEPT_PARSE_INVALID_UNICODE_HEX`：表示 `\uXXXX` 中的 XXXX 不是有效的十六进制数
- `LEPT_PARSE_INVALID_UNICODE_SURROGATE`：表示代理对不正确

### 4.4.2 解析 Unicode 转义序列

现在，我们需要修改字符串解析函数，增加对 `\uXXXX` 转义序列的处理。首先，我们编写一个函数来解析 4 位十六进制数：

```c
static const char* lept_parse_hex4(const char* p, unsigned* u) {
    int i;
    *u = 0;
    for (i = 0; i < 4; i++) {
        char ch = *p++;
        *u <<= 4;
        if      (ch >= '0' && ch <= '9')  *u |= ch - '0';
        else if (ch >= 'A' && ch <= 'F')  *u |= ch - ('A' - 10);
        else if (ch >= 'a' && ch <= 'f')  *u |= ch - ('a' - 10);
        else return NULL;
    }
    return p;
}
```

这个函数接收一个字符串指针 `p` 和一个无符号整数指针 `u`，解析 `p` 开始的 4 个字符作为十六进制数，并将结果存储在 `u` 中。如果解析成功，返回解析后的位置；如果失败，返回 `NULL`。

### 4.4.3 处理代理对和 UTF-8 编码

然后，我们需要将 Unicode 代码点转换为 UTF-8 编码。我们定义一个函数 `lept_encode_utf8` 来完成这个工作：

```c
static void lept_encode_utf8(lept_context* c, unsigned u) {
    if (u <= 0x7F)
        PUTC(c, u & 0xFF);
    else if (u <= 0x7FF) {
        PUTC(c, 0xC0 | ((u >> 6) & 0xFF));
        PUTC(c, 0x80 | (u & 0x3F));
    }
    else if (u <= 0xFFFF) {
        PUTC(c, 0xE0 | ((u >> 12) & 0xFF));
        PUTC(c, 0x80 | ((u >>  6) & 0x3F));
        PUTC(c, 0x80 | (u & 0x3F));
    }
    else {
        PUTC(c, 0xF0 | ((u >> 18) & 0xFF));
        PUTC(c, 0x80 | ((u >> 12) & 0x3F));
        PUTC(c, 0x80 | ((u >>  6) & 0x3F));
        PUTC(c, 0x80 | (u & 0x3F));
    }
}
```

这个函数根据 Unicode 代码点的范围，使用不同的 UTF-8 编码规则，将字符添加到 context 的缓冲区中。

### 4.4.4 修改 parseString 函数

现在，我们修改 `lept_parse_string` 函数，增加对 Unicode 转义序列的处理：

```c
static int lept_parse_string(lept_context* c, lept_value* v) {
    unsigned u, u2;
    // ... existing code ...
    
    for (;;) {
        char ch = *p++;
        switch (ch) {
            // ... existing code for other escape characters ...
            
            case '\\':
                switch (*p++) {
                    // ... existing switch cases ...
                    case 'u':
                        if (!(p = lept_parse_hex4(p, &u)))
                            STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_HEX);
                        
                        // 处理代理对
                        if (u >= 0xD800 && u <= 0xDBFF) {
                            // 高代理项，需要后面跟着低代理项
                            if (*p++ != '\\' || *p++ != 'u')
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_SURROGATE);
                            if (!(p = lept_parse_hex4(p, &u2)))
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_HEX);
                            if (u2 < 0xDC00 || u2 > 0xDFFF)
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_SURROGATE);
                                
                            // 计算实际的 Unicode 代码点
                            u = 0x10000 + (((u - 0xD800) << 10) | (u2 - 0xDC00));
                        }
                        
                        lept_encode_utf8(c, u);
                        break;
                    // ... other cases ...
                }
                break;
            // ... existing code ...
        }
    }
}
```

在这个修改中，我们：
1. 解析 `\uXXXX` 格式的 Unicode 字符
2. 检查是否是高代理项，如果是，则继续解析下一个转义序列，期望它是一个低代理项
3. 如果确实是一个代理对，则计算实际的 Unicode 代码点
4. 将解析得到的 Unicode 代码点编码为 UTF-8，并添加到结果字符串中

## 4.5 测试 Unicode 支持

为了测试我们的 Unicode 支持，我们需要编写几个测试用例：

```c
static void test_parse_unicode() {
    TEST_STRING("Hello\\u0000World", "Hello\0World");
    TEST_STRING("\\u4E2D\\u56FD", "中国");  // 测试中文
    TEST_STRING("\\uD834\\uDD1E", "𝄞");    // 测试代理对 (G clef 符号 U+1D11E)
    TEST_STRING("\\u00A2", "¢");           // 美分符号 U+00A2
    TEST_STRING("\\u2028", "\xe2\x80\xa8");// 行分隔符 U+2028
}

static void test_parse_invalid_unicode() {
    // 无效的十六进制数
    TEST_ERROR("\"\\u123\"", LEPT_PARSE_INVALID_UNICODE_HEX);
    // 高代理项后面不是低代理项
    TEST_ERROR("\"\\uD834\"", LEPT_PARSE_INVALID_UNICODE_SURROGATE);
    // 高代理项后面不是转义序列
    TEST_ERROR("\"\\uD834X\"", LEPT_PARSE_INVALID_UNICODE_SURROGATE);
    // 高代理项后面的转义序列不是 \u
    TEST_ERROR("\"\\uD834\\\"", LEPT_PARSE_INVALID_UNICODE_SURROGATE);
    // 高代理项后面的 \u 转义序列不是有效的十六进制数
    TEST_ERROR("\"\\uD834\\uXXXX\"", LEPT_PARSE_INVALID_UNICODE_HEX);
    // 高代理项后面的 \u 转义序列不是低代理项
    TEST_ERROR("\"\\uD834\\u1234\"", LEPT_PARSE_INVALID_UNICODE_SURROGATE);
}
```

## 4.6 性能考虑

处理 Unicode 可能会导致性能下降，特别是在解析大量包含 Unicode 字符的 JSON 文本时。为了优化性能，我们可以考虑以下几点：

1. 预分配足够的内存，避免频繁的内存重新分配
2. 使用查找表加速 Unicode 到 UTF-8 的转换
3. 对于特定平台，考虑使用 SIMD 指令集加速处理

## 4.7 练习

1. 实现一个函数 `lept_stringify`，将 lept_value 转换为 JSON 字符串。该函数需要正确处理 Unicode 字符，包括将非 ASCII 字符转换为 `\uXXXX` 格式。

2. 扩展错误处理，检测更多的 Unicode 相关错误，如无效的 UTF-8 编码序列。

3. 实现一个更高效的 UTF-8 编码函数，使用查找表或其他优化技术。

## 4.8 下一步

在下一章中，我们将实现 JSON 数组的解析，这将使我们的 JSON 库更接近完整。 