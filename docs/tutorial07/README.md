# 第7章：生成器

在前面的章节中，我们已经实现了一个完整的 JSON 解析器，可以将 JSON 文本解析为内存中的数据结构。在本章中，我们将实现相反的功能——JSON 生成器（generator/stringifier），将内存中的数据结构转换为 JSON 文本。

## 7.1 JSON 生成器的设计

生成器的主要功能是将 `lept_value` 结构体转换为 JSON 文本。我们需要为每种 JSON 类型实现相应的生成函数，然后将这些函数组合起来，形成一个完整的生成器。

我们先定义生成器的接口函数：

```c
char* lept_stringify(const lept_value* v, size_t* length);
```

这个函数接收一个 `lept_value` 指针和一个用于返回生成的 JSON 文本长度的指针，返回新分配的字符串，保存生成的 JSON 文本。调用方需要负责释放这个字符串的内存。

## 7.2 字符串化的实现

我们首先需要一个上下文（context）来存储生成的 JSON 文本：

```c
typedef struct {
    char* buf;
    size_t size;
    size_t capacity;
} lept_stringify_context;
```

其中，`buf` 是一个动态分配的字符缓冲区，用于存放生成的 JSON 文本；`size` 是当前已使用的缓冲区大小；`capacity` 是当前缓冲区的容量。

接下来，我们实现一个函数用于向缓冲区添加字符：

```c
static void lept_stringify_context_push(lept_stringify_context* c, const char* s, size_t len) {
    if (c->size + len > c->capacity) {
        if (c->capacity == 0)
            c->capacity = LEPT_STRINGIFY_INIT_SIZE;
        while (c->size + len > c->capacity)
            c->capacity += c->capacity >> 1;  /* 扩展容量为当前的 1.5 倍 */
        c->buf = (char*)realloc(c->buf, c->capacity);
    }
    memcpy(c->buf + c->size, s, len);
    c->size += len;
}
```

这个函数会自动扩展缓冲区的容量，以确保能够容纳新添加的内容。

有了上下文和添加字符的函数，我们就可以开始实现各种 JSON 类型的生成函数了。

### 7.2.1 生成 null, true 和 false

对于 null, true 和 false 这三种 JSON 值，生成函数非常简单：

```c
static void lept_stringify_value(lept_stringify_context* c, const lept_value* v);

static void lept_stringify_literal(lept_stringify_context* c, const char* literal, size_t len) {
    lept_stringify_context_push(c, literal, len);
}

static void lept_stringify_null(lept_stringify_context* c) {
    lept_stringify_literal(c, "null", 4);
}

static void lept_stringify_boolean(lept_stringify_context* c, int boolean) {
    if (boolean)
        lept_stringify_literal(c, "true", 4);
    else
        lept_stringify_literal(c, "false", 5);
}
```

### 7.2.2 生成数字

对于数字类型，我们需要将 `double` 类型的值转换为字符串。C 语言提供了 `sprintf` 函数，但它不会告诉我们生成的字符串长度，因此我们需要使用 `snprintf`：

```c
static void lept_stringify_number(lept_stringify_context* c, double number) {
    char buffer[32];
    int length = snprintf(buffer, sizeof(buffer), "%.17g", number);
    lept_stringify_context_push(c, buffer, length);
}
```

这里我们使用 `%.17g` 作为格式化字符串，它可以精确地表示双精度浮点数，同时避免不必要的数字。

### 7.2.3 生成字符串

生成字符串是最复杂的部分，因为我们需要处理转义字符：

```c
static void lept_stringify_string(lept_stringify_context* c, const char* s, size_t len) {
    size_t i;
    lept_stringify_context_push(c, "\"", 1);  /* 添加开头的双引号 */
    for (i = 0; i < len; i++) {
        unsigned char ch = (unsigned char)s[i];
        switch (ch) {
            case '\"': lept_stringify_context_push(c, "\\\"", 2); break;
            case '\\': lept_stringify_context_push(c, "\\\\", 2); break;
            case '\b': lept_stringify_context_push(c, "\\b", 2); break;
            case '\f': lept_stringify_context_push(c, "\\f", 2); break;
            case '\n': lept_stringify_context_push(c, "\\n", 2); break;
            case '\r': lept_stringify_context_push(c, "\\r", 2); break;
            case '\t': lept_stringify_context_push(c, "\\t", 2); break;
            default:
                if (ch < 0x20) {
                    char buffer[7];
                    sprintf(buffer, "\\u%04X", ch);
                    lept_stringify_context_push(c, buffer, 6);
                }
                else
                    lept_stringify_context_push(c, &s[i], 1);
        }
    }
    lept_stringify_context_push(c, "\"", 1);  /* 添加结尾的双引号 */
}
```

在这个函数中，我们处理了 JSON 中的所有转义字符，将它们转换为对应的转义序列。对于小于 0x20 的控制字符，我们使用 `\uXXXX` 的形式表示。

### 7.2.4 生成数组

生成数组需要处理数组中的每个元素，并在元素之间添加逗号：

```c
static void lept_stringify_array(lept_stringify_context* c, const lept_value* v) {
    size_t i;
    lept_stringify_context_push(c, "[", 1);  /* 添加开头的方括号 */
    for (i = 0; i < v->u.a.size; i++) {
        if (i > 0)
            lept_stringify_context_push(c, ",", 1);  /* 在元素之间添加逗号 */
        lept_stringify_value(c, &v->u.a.e[i]);
    }
    lept_stringify_context_push(c, "]", 1);  /* 添加结尾的方括号 */
}
```

### 7.2.5 生成对象

生成对象与生成数组类似，但需要处理键值对：

```c
static void lept_stringify_object(lept_stringify_context* c, const lept_value* v) {
    size_t i;
    lept_stringify_context_push(c, "{", 1);  /* 添加开头的花括号 */
    for (i = 0; i < v->u.o.size; i++) {
        if (i > 0)
            lept_stringify_context_push(c, ",", 1);  /* 在键值对之间添加逗号 */
        lept_stringify_string(c, v->u.o.m[i].k, v->u.o.m[i].klen);  /* 生成键 */
        lept_stringify_context_push(c, ":", 1);  /* 在键和值之间添加冒号 */
        lept_stringify_value(c, &v->u.o.m[i].v);  /* 生成值 */
    }
    lept_stringify_context_push(c, "}", 1);  /* 添加结尾的花括号 */
}
```

### 7.2.6 整合

最后，我们需要一个函数来整合所有的生成函数：

```c
static void lept_stringify_value(lept_stringify_context* c, const lept_value* v) {
    switch (v->type) {
        case LEPT_NULL:   lept_stringify_null(c); break;
        case LEPT_FALSE:  lept_stringify_boolean(c, 0); break;
        case LEPT_TRUE:   lept_stringify_boolean(c, 1); break;
        case LEPT_NUMBER: lept_stringify_number(c, v->u.n); break;
        case LEPT_STRING: lept_stringify_string(c, v->u.s.s, v->u.s.len); break;
        case LEPT_ARRAY:  lept_stringify_array(c, v); break;
        case LEPT_OBJECT: lept_stringify_object(c, v); break;
        default: break;
    }
}

char* lept_stringify(const lept_value* v, size_t* length) {
    lept_stringify_context c;
    assert(v != NULL);
    c.buf = NULL;
    c.size = c.capacity = 0;
    lept_stringify_value(&c, v);
    if (length)
        *length = c.size;
    lept_stringify_context_push(&c, "\0", 1);  /* 添加结尾的空字符 */
    return c.buf;
}
```

## 7.3 格式化选项

在一些应用场景中，我们可能希望生成的 JSON 文本具有良好的可读性，例如添加缩进和换行。为此，我们可以为生成器添加格式化选项：

```c
typedef struct {
    int indent_size;
    int pretty_format;
} lept_stringify_options;
```

其中，`indent_size` 表示缩进的空格数，`pretty_format` 表示是否使用美化格式（添加缩进和换行）。

然后，我们修改生成器的接口函数：

```c
char* lept_stringify_ex(const lept_value* v, size_t* length, const lept_stringify_options* options);
```

并提供一个简化的接口，使用默认选项：

```c
char* lept_stringify(const lept_value* v, size_t* length) {
    lept_stringify_options default_options = { 0, 0 };  /* 默认不美化，缩进为 0 */
    return lept_stringify_ex(v, length, &default_options);
}
```

美化格式的实现略复杂，这里不详细展开。感兴趣的读者可以自行尝试实现。

## 7.4 测试生成器

为了测试我们的生成器，我们需要编写一些测试用例：

```c
static void test_stringify_null() {
    lept_value v;
    lept_init(&v);
    lept_set_null(&v);
    EXPECT_STREQ("null", lept_stringify(&v, NULL));
    lept_free(&v);
}

static void test_stringify_boolean() {
    lept_value v;
    lept_init(&v);
    lept_set_boolean(&v, 1);
    EXPECT_STREQ("true", lept_stringify(&v, NULL));
    lept_set_boolean(&v, 0);
    EXPECT_STREQ("false", lept_stringify(&v, NULL));
    lept_free(&v);
}

static void test_stringify_number() {
    lept_value v;
    lept_init(&v);
    lept_set_number(&v, 123.0);
    EXPECT_STREQ("123", lept_stringify(&v, NULL));
    lept_set_number(&v, 0.0);
    EXPECT_STREQ("0", lept_stringify(&v, NULL));
    lept_set_number(&v, -0.0);
    EXPECT_STREQ("0", lept_stringify(&v, NULL));
    lept_set_number(&v, 1.2345);
    EXPECT_STREQ("1.2345", lept_stringify(&v, NULL));
    lept_set_number(&v, -1.2345);
    EXPECT_STREQ("-1.2345", lept_stringify(&v, NULL));
    lept_set_number(&v, 1e20);
    EXPECT_STREQ("1e+20", lept_stringify(&v, NULL));
    lept_set_number(&v, 1.234e+20);
    EXPECT_STREQ("1.234e+20", lept_stringify(&v, NULL));
    lept_set_number(&v, 1.234e-20);
    EXPECT_STREQ("1.234e-20", lept_stringify(&v, NULL));
    lept_free(&v);
}

static void test_stringify_string() {
    lept_value v;
    lept_init(&v);
    lept_set_string(&v, "", 0);
    EXPECT_STREQ("\"\"", lept_stringify(&v, NULL));
    lept_set_string(&v, "Hello", 5);
    EXPECT_STREQ("\"Hello\"", lept_stringify(&v, NULL));
    lept_set_string(&v, "Hello\nWorld", 11);
    EXPECT_STREQ("\"Hello\\nWorld\"", lept_stringify(&v, NULL));
    lept_set_string(&v, "\" \\ / \b \f \n \r \t", 13);
    EXPECT_STREQ("\"\\\" \\\\ \\/ \\b \\f \\n \\r \\t\"", lept_stringify(&v, NULL));
    lept_set_string(&v, "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1A\x1B\x1C\x1D\x1E\x1F", 31);
    EXPECT_STREQ("\"\\u0001\\u0002\\u0003\\u0004\\u0005\\u0006\\u0007\\b\\t\\n\\u000B\\f\\r\\u000E\\u000F\\u0010\\u0011\\u0012\\u0013\\u0014\\u0015\\u0016\\u0017\\u0018\\u0019\\u001A\\u001B\\u001C\\u001D\\u001E\\u001F\"", lept_stringify(&v, NULL));
    lept_free(&v);
}

static void test_stringify_array() {
    lept_value v, e;
    lept_init(&v);
    lept_set_array(&v, 0);
    EXPECT_STREQ("[]", lept_stringify(&v, NULL));
    
    lept_init(&e);
    lept_set_number(&e, 1.0);
    lept_push_array_element(&v, &e);
    lept_free(&e);
    EXPECT_STREQ("[1]", lept_stringify(&v, NULL));
    
    lept_init(&e);
    lept_set_number(&e, 2.0);
    lept_push_array_element(&v, &e);
    lept_free(&e);
    EXPECT_STREQ("[1,2]", lept_stringify(&v, NULL));
    
    lept_init(&e);
    lept_set_number(&e, 3.0);
    lept_push_array_element(&v, &e);
    lept_free(&e);
    EXPECT_STREQ("[1,2,3]", lept_stringify(&v, NULL));
    
    lept_free(&v);
}

static void test_stringify_object() {
    lept_value v, s, a;
    lept_init(&v);
    lept_set_object(&v, 0);
    EXPECT_STREQ("{}", lept_stringify(&v, NULL));
    
    lept_init(&s);
    lept_set_string(&s, "name", 4);
    lept_set_object_value(&v, "name", 4, &s);
    lept_free(&s);
    EXPECT_STREQ("{\"name\":\"name\"}", lept_stringify(&v, NULL));
    
    lept_init(&a);
    lept_set_array(&a, 0);
    lept_set_object_value(&v, "array", 5, &a);
    lept_free(&a);
    EXPECT_STREQ("{\"name\":\"name\",\"array\":[]}", lept_stringify(&v, NULL));
    
    lept_free(&v);
}

static void test_stringify() {
    test_stringify_null();
    test_stringify_boolean();
    test_stringify_number();
    test_stringify_string();
    test_stringify_array();
    test_stringify_object();
}
```

这些测试用例覆盖了所有 JSON 数据类型的生成，包括各种特殊情况。

## 7.5 优化

### 7.5.1 内存优化

在当前的实现中，每次向缓冲区添加内容时，如果容量不足，我们会重新分配内存。这可能会导致频繁的内存分配和拷贝，影响性能。我们可以尝试优化内存分配策略，例如预先分配足够的内存：

```c
char* lept_stringify(const lept_value* v, size_t* length) {
    lept_stringify_context c;
    size_t estimated_size = lept_stringify_estimate_size(v);  /* 估计生成的 JSON 文本大小 */
    assert(v != NULL);
    c.buf = (char*)malloc(estimated_size);
    c.size = 0;
    c.capacity = estimated_size;
    lept_stringify_value(&c, v);
    if (length)
        *length = c.size;
    lept_stringify_context_push(&c, "\0", 1);  /* 添加结尾的空字符 */
    c.buf = (char*)realloc(c.buf, c.size);  /* 缩小内存到实际大小 */
    return c.buf;
}
```

其中，`lept_stringify_estimate_size` 函数用于估计生成的 JSON 文本大小。它的实现可以是简单的，例如返回一个固定大小（如 1024 或更大），也可以是复杂的，根据 `lept_value` 的类型和内容进行估计。

### 7.5.2 性能优化

在生成字符串时，我们对每个字符进行了处理，这可能导致频繁的函数调用和内存操作。我们可以对特定情况进行优化，例如对于不包含需要转义的字符的字符串，直接进行内存拷贝：

```c
static void lept_stringify_string(lept_stringify_context* c, const char* s, size_t len) {
    size_t i;
    int need_escape = 0;
    
    for (i = 0; i < len; i++) {
        unsigned char ch = (unsigned char)s[i];
        if (ch < 0x20 || ch == '\"' || ch == '\\') {
            need_escape = 1;
            break;
        }
    }
    
    lept_stringify_context_push(c, "\"", 1);  /* 添加开头的双引号 */
    
    if (need_escape) {
        for (i = 0; i < len; i++) {
            unsigned char ch = (unsigned char)s[i];
            switch (ch) {
                case '\"': lept_stringify_context_push(c, "\\\"", 2); break;
                case '\\': lept_stringify_context_push(c, "\\\\", 2); break;
                case '\b': lept_stringify_context_push(c, "\\b", 2); break;
                case '\f': lept_stringify_context_push(c, "\\f", 2); break;
                case '\n': lept_stringify_context_push(c, "\\n", 2); break;
                case '\r': lept_stringify_context_push(c, "\\r", 2); break;
                case '\t': lept_stringify_context_push(c, "\\t", 2); break;
                default:
                    if (ch < 0x20) {
                        char buffer[7];
                        sprintf(buffer, "\\u%04X", ch);
                        lept_stringify_context_push(c, buffer, 6);
                    }
                    else
                        lept_stringify_context_push(c, &s[i], 1);
            }
        }
    }
    else {
        lept_stringify_context_push(c, s, len);  /* 直接拷贝整个字符串 */
    }
    
    lept_stringify_context_push(c, "\"", 1);  /* 添加结尾的双引号 */
}
```

这样，对于不包含需要转义的字符的字符串，我们可以一次性将整个字符串拷贝到缓冲区中，避免了对每个字符的处理。

## 7.6 练习

1. 实现 `lept_stringify_ex` 函数，支持格式化选项。添加缩进和换行，使生成的 JSON 文本更易读。

2. 实现 `lept_stringify_estimate_size` 函数，根据 `lept_value` 的类型和内容估计生成的 JSON 文本大小，用于优化内存分配。

3. 添加更多的测试用例，确保生成器能够正确处理各种 JSON 数据类型和特殊情况。

4. 实现一个函数 `lept_stringify_file`，将 `lept_value` 直接写入文件，而不是生成字符串。这对于大型 JSON 数据特别有用。

## 7.7 下一步

在本章中，我们实现了 JSON 生成器，可以将内存中的数据结构转换为 JSON 文本。现在，我们的 JSON 库已经具备了解析和生成的基本功能。在下一章中，我们将为这个库添加更多实用的功能，如访问元素、修改内容等，使其更加完整和易用。 