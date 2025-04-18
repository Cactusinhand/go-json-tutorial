# 第9章：增强的错误处理

在前面的章节中，我们已经实现了基本的错误处理，但它只能返回有限的错误类型，无法提供具体的错误位置和上下文信息。在本章中，我们将增强我们的 JSON 库的错误处理能力，提供更详细的错误信息，帮助用户更容易地调试和修复问题。

## 9.1 错误处理的需求

在一个优秀的库中，错误处理是非常重要的一部分。良好的错误处理应该满足以下需求：

1. **明确的错误类型**：用户应该能够知道发生了什么类型的错误，如语法错误、类型错误等。
2. **详细的错误信息**：错误信息应该尽可能详细，包括错误位置、错误上下文等。
3. **易于理解的错误消息**：错误消息应该易于理解，帮助用户快速定位问题。
4. **可追踪的错误路径**：对于嵌套结构，应该能够知道错误发生在哪个层级。

## 9.2 增强错误信息结构

首先，我们需要定义一个更丰富的错误信息结构，用于保存错误的详细信息：

```c
typedef struct {
    size_t line;          /* 错误发生的行号 */
    size_t column;        /* 错误发生的列号 */
    const char* position; /* 错误发生的位置 */
    const char* message;  /* 错误消息 */
    const char* path;     /* 错误路径（如 "obj.arr[0].key"） */
    int code;             /* 错误代码 */
} lept_error;
```

这个结构包含了错误的位置、消息、路径和代码，可以帮助用户更好地理解和定位错误。

## 9.3 修改解析上下文

接下来，我们需要修改解析上下文，加入行号和列号的跟踪：

```c
typedef struct {
    const char* json;
    char* stack;
    size_t size, top;
    size_t line, column;  /* 当前解析的行号和列号 */
    lept_error error;     /* 错误信息 */
} lept_context;
```

我们需要在解析过程中更新行号和列号：

```c
static void lept_parse_whitespace(lept_context* c) {
    const char* p = c->json;
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') {
        if (*p == '\n') {
            c->line++;
            c->column = 0;
        }
        else {
            c->column++;
        }
        p++;
    }
    c->json = p;
}
```

这样，当遇到换行符时，我们增加行号并重置列号；对于其他字符，我们只增加列号。

## 9.4 设置错误信息

我们需要一个函数来设置错误信息：

```c
static void lept_set_error(lept_context* c, int code, const char* message) {
    c->error.code = code;
    c->error.line = c->line;
    c->error.column = c->column;
    c->error.position = c->json;
    c->error.message = message;
    c->error.path = NULL;  /* 路径信息需要另外设置 */
}
```

然后，我们修改所有的错误处理代码，使用 `lept_set_error` 函数设置详细的错误信息：

```c
static int lept_parse_null(lept_context* c, lept_value* v) {
    EXPECT(c, 'n');
    if (c->json[0] != 'u' || c->json[1] != 'l' || c->json[2] != 'l') {
        lept_set_error(c, LEPT_PARSE_INVALID_VALUE, "Invalid value: expected 'null'");
        return LEPT_PARSE_INVALID_VALUE;
    }
    c->json += 3;
    v->type = LEPT_NULL;
    return LEPT_PARSE_OK;
}
```

对于每种错误情况，我们都提供一个具体的错误消息，帮助用户理解发生了什么错误。

## 9.5 错误路径跟踪

为了提供错误的路径信息，我们需要在解析过程中跟踪当前的路径。我们可以使用一个栈来存储路径信息：

```c
typedef struct {
    const char* name;  /* 当前元素的名称 */
    size_t index;      /* 当前元素的索引（对于数组） */
    int is_array;      /* 当前元素是否是数组 */
} lept_path_entry;

#define LEPT_CONTEXT_PATH_STACK_INIT_SIZE 256

/* 在解析上下文中添加路径栈 */
typedef struct {
    const char* json;
    char* stack;
    size_t size, top;
    size_t line, column;
    lept_error error;
    lept_path_entry* path_stack;  /* 路径栈 */
    size_t path_size, path_top;   /* 路径栈的大小和顶部位置 */
} lept_context;
```

我们需要一些函数来管理路径栈：

```c
static void lept_context_push_path(lept_context* c, const char* name, size_t index, int is_array) {
    if (c->path_top + 1 >= c->path_size) {
        if (c->path_size == 0)
            c->path_size = LEPT_CONTEXT_PATH_STACK_INIT_SIZE;
        else
            c->path_size += c->path_size >> 1;  /* 扩展为 1.5 倍 */
        c->path_stack = (lept_path_entry*)realloc(c->path_stack, c->path_size * sizeof(lept_path_entry));
    }
    lept_path_entry* entry = &c->path_stack[c->path_top++];
    entry->name = name;
    entry->index = index;
    entry->is_array = is_array;
}

static void lept_context_pop_path(lept_context* c) {
    assert(c->path_top > 0);
    c->path_top--;
}

static void lept_context_build_path(lept_context* c, char* buffer, size_t size) {
    size_t i, offset = 0;
    buffer[0] = '\0';
    for (i = 0; i < c->path_top; i++) {
        const lept_path_entry* entry = &c->path_stack[i];
        if (entry->is_array) {
            offset += snprintf(buffer + offset, size - offset, "[%zu]", entry->index);
        }
        else {
            if (i > 0)
                offset += snprintf(buffer + offset, size - offset, ".");
            if (entry->name)
                offset += snprintf(buffer + offset, size - offset, "%s", entry->name);
        }
    }
}
```

当我们解析数组或对象时，我们需要记录当前的路径：

```c
static int lept_parse_array(lept_context* c, lept_value* v) {
    size_t size = 0;
    int ret;
    EXPECT(c, '[');
    lept_parse_whitespace(c);
    if (*c->json == ']') {
        c->json++;
        v->type = LEPT_ARRAY;
        v->u.a.size = 0;
        v->u.a.e = NULL;
        return LEPT_PARSE_OK;
    }
    for (;;) {
        lept_value e;
        lept_init(&e);
        lept_context_push_path(c, NULL, size, 1);  /* 记录数组元素的路径 */
        if ((ret = lept_parse_value(c, &e)) != LEPT_PARSE_OK) {
            lept_context_pop_path(c);
            break;
        }
        lept_context_pop_path(c);
        memcpy(lept_context_push(c, sizeof(lept_value)), &e, sizeof(lept_value));
        size++;
        lept_parse_whitespace(c);
        if (*c->json == ',') {
            c->json++;
            lept_parse_whitespace(c);
        }
        else if (*c->json == ']') {
            c->json++;
            v->type = LEPT_ARRAY;
            v->u.a.size = size;
            size *= sizeof(lept_value);
            memcpy(v->u.a.e = (lept_value*)malloc(size), lept_context_pop(c, size), size);
            return LEPT_PARSE_OK;
        }
        else {
            lept_set_error(c, LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET, "Expected ',' or ']'");
            ret = LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET;
            break;
        }
    }
    /* ... */
    return ret;
}
```

同样，对于对象的解析，我们也需要记录路径：

```c
static int lept_parse_object(lept_context* c, lept_value* v) {
    size_t size = 0;
    lept_member m;
    int ret;
    EXPECT(c, '{');
    lept_parse_whitespace(c);
    if (*c->json == '}') {
        c->json++;
        v->type = LEPT_OBJECT;
        v->u.o.m = NULL;
        v->u.o.size = 0;
        return LEPT_PARSE_OK;
    }
    for (;;) {
        char* str;
        lept_init(&m.v);
        /* parse key */
        if (*c->json != '"') {
            lept_set_error(c, LEPT_PARSE_MISS_KEY, "Expected '\"' at the beginning of object key");
            ret = LEPT_PARSE_MISS_KEY;
            break;
        }
        if ((ret = lept_parse_string_raw(c, &str, &m.klen)) != LEPT_PARSE_OK)
            break;
        m.k = (char*)malloc(m.klen + 1);
        memcpy(m.k, str, m.klen);
        m.k[m.klen] = '\0';
        /* parse ws colon ws */
        lept_parse_whitespace(c);
        if (*c->json != ':') {
            lept_set_error(c, LEPT_PARSE_MISS_COLON, "Expected ':' after object key");
            ret = LEPT_PARSE_MISS_COLON;
            break;
        }
        c->json++;
        lept_parse_whitespace(c);
        /* parse value */
        lept_context_push_path(c, m.k, 0, 0);  /* 记录对象成员的路径 */
        if ((ret = lept_parse_value(c, &m.v)) != LEPT_PARSE_OK) {
            lept_context_pop_path(c);
            break;
        }
        lept_context_pop_path(c);
        memcpy(lept_context_push(c, sizeof(lept_member)), &m, sizeof(lept_member));
        size++;
        m.k = NULL; /* ownership is transferred to member on stack */
        /* parse ws [comma | right-curly-brace] ws */
        lept_parse_whitespace(c);
        if (*c->json == ',') {
            c->json++;
            lept_parse_whitespace(c);
        }
        else if (*c->json == '}') {
            c->json++;
            v->type = LEPT_OBJECT;
            v->u.o.size = size;
            size *= sizeof(lept_member);
            memcpy(v->u.o.m = (lept_member*)malloc(size), lept_context_pop(c, size), size);
            return LEPT_PARSE_OK;
        }
        else {
            lept_set_error(c, LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET, "Expected ',' or '}'");
            ret = LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET;
            break;
        }
    }
    /* ... */
    return ret;
}
```

## 9.6 获取错误信息

最后，我们需要提供一个函数，让用户可以获取详细的错误信息：

```c
const lept_error* lept_get_error(const lept_value* v) {
    return &v->error;
}

const char* lept_get_error_message(const lept_error* e) {
    static char buffer[1024];
    if (e->path)
        snprintf(buffer, sizeof(buffer), "Line %zu Column %zu: %s (at %s)", e->line, e->column, e->message, e->path);
    else
        snprintf(buffer, sizeof(buffer), "Line %zu Column %zu: %s", e->line, e->column, e->message);
    return buffer;
}
```

这样，用户可以获取一个格式化的错误消息，包含行号、列号、错误消息和路径信息。

## 9.7 测试增强的错误处理

为了测试我们的增强错误处理，我们可以编写一些测试用例：

```c
static void test_parse_error_position() {
    lept_value v;
    
    /* 测试行号和列号跟踪 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_INVALID_VALUE, lept_parse(&v, "\n\n\n  null x"));
    const lept_error* e = lept_get_error(&v);
    EXPECT_EQ_SIZE_T(4, e->line);
    EXPECT_EQ_SIZE_T(9, e->column);
    lept_free(&v);
    
    /* 测试错误路径 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_INVALID_VALUE, lept_parse(&v, "{\"obj\":{\"array\":[1,2,true x]}}"));
    e = lept_get_error(&v);
    EXPECT_STREQ("obj.array[2]", e->path);
    lept_free(&v);
}
```

## 9.8 错误处理的最佳实践

在实际应用中，良好的错误处理是很重要的。以下是一些错误处理的最佳实践：

1. **提供详细的错误信息**：错误信息应该尽可能详细，包括错误类型、位置、上下文等。

2. **使用适当的错误代码**：为不同类型的错误定义不同的错误代码，方便用户区分和处理。

3. **避免泄露资源**：在错误处理过程中，应该确保释放所有分配的资源，避免内存泄漏。

4. **提供错误恢复机制**：如果可能，提供错误恢复机制，让程序在发生错误后能够继续运行。

5. **记录错误信息**：在适当的情况下，将错误信息记录到日志中，方便后续分析和调试。

## 9.9 练习

1. 实现一个函数 `lept_get_error_position`，用于获取错误的位置信息。

2. 改进 `lept_parse_string` 函数，提供更详细的错误信息，如"Expected '\"' at the end of string"。

3. 为 `lept_stringify` 函数添加错误处理，当生成 JSON 文本失败时，提供详细的错误信息。

4. 实现一个函数 `lept_error_to_string`，将错误信息转换为适合打印的字符串。

5. 改进 `lept_patch` 函数，当应用 JSON Patch 失败时，提供详细的错误信息。

## 9.10 下一步

在本章中，我们增强了 JSON 库的错误处理能力，提供了更详细的错误信息。这使得用户可以更容易地调试和修复问题。在下一章中，我们将实现 JSON 指针，它是一种用于在 JSON 文档中定位特定位置的标准化语法。 