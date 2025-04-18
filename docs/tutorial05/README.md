# 第5章：数组解析

在完成了 Unicode 支持后，我们的 JSON 解析器已经可以处理字符串类型的值了。本章将实现 JSON 数组的解析，这是 JSON 的重要数据结构之一。

## 5.1 JSON 数组

JSON 数组是一个有序的值的集合，使用方括号 `[]` 表示，值之间用逗号 `,` 分隔。数组中的元素可以是任意 JSON 数据类型，包括数字、字符串、布尔值、null、数组和对象。数组甚至可以包含不同类型的元素。

例如：
```json
[1, 2, 3]
["apple", "banana", "orange"]
[true, false, null, 42, "text", [1, 2], {"key": "value"}]
[]
```

最后一个例子是一个空数组，它不包含任何元素。

## 5.2 修改数据结构

首先，我们需要在 `lept_value` 结构体中添加数组类型的支持。修改 `lept_value` 结构体如下：

```c
typedef struct lept_value lept_value;

struct lept_value {
    union {
        struct { lept_value* e; size_t size; }a; /* array */
        struct { char* s; size_t len; }s;         /* string */
        double n;                                 /* number */
    } u;
    lept_type type;
};
```

上面的代码中，我们添加了一个结构体 `a` 用于表示数组，其中：
- `e` 是指向数组元素的指针
- `size` 是数组中元素的个数

同时，我们也需要在 `lept_type` 枚举类型中添加数组类型：

```c
typedef enum {
    LEPT_NULL,
    LEPT_FALSE,
    LEPT_TRUE,
    LEPT_NUMBER,
    LEPT_STRING,
    LEPT_ARRAY,
    LEPT_OBJECT
} lept_type;
```

## 5.3 访问数组的接口

在实现数组解析之前，我们先定义一些用于访问数组的函数：

```c
size_t lept_get_array_size(const lept_value* v);
lept_value* lept_get_array_element(const lept_value* v, size_t index);
```

这些函数的实现也很简单：

```c
size_t lept_get_array_size(const lept_value* v) {
    assert(v != NULL && v->type == LEPT_ARRAY);
    return v->u.a.size;
}

lept_value* lept_get_array_element(const lept_value* v, size_t index) {
    assert(v != NULL && v->type == LEPT_ARRAY && index < v->u.a.size);
    return &v->u.a.e[index];
}
```

`lept_get_array_size` 返回数组中元素的个数，`lept_get_array_element` 返回指定索引的元素。

## 5.4 解析 JSON 数组

现在我们来实现数组的解析。我们需要增加一个新的函数 `lept_parse_array`：

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
        if ((ret = lept_parse_value(c, &e)) != LEPT_PARSE_OK)
            break;
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
            ret = LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET;
            break;
        }
    }
    /* Pop and free values on the stack */
    for (int i = 0; i < size; i++)
        lept_free((lept_value*)lept_context_pop(c, sizeof(lept_value)));
    return ret;
}
```

解析数组的流程如下：
1. 判断是否为空数组 `[]`，若是则设置 `size` 为 0，`e` 为 NULL，并返回
2. 循环解析数组中的元素：
   - 解析一个值，并将其压入堆栈
   - 判断是否有逗号 `,`，若有则继续解析下一个元素
   - 若遇到右方括号 `]`，则分配内存存储所有元素，并返回
   - 若解析失败或格式错误，则释放已解析的元素并返回错误

同时，我们还需要在 `lept_parse_value` 函数中增加对数组的处理：

```c
static int lept_parse_value(lept_context* c, lept_value* v) {
    switch (*c->json) {
        case 'n':  return lept_parse_null(c, v);
        case 't':  return lept_parse_true(c, v);
        case 'f':  return lept_parse_false(c, v);
        case '"':  return lept_parse_string(c, v);
        case '[':  return lept_parse_array(c, v);
        case '\0': return LEPT_PARSE_EXPECT_VALUE;
        default:   return lept_parse_number(c, v);
    }
}
```

另外，我们还需要修改 `lept_free` 函数，以释放数组中的元素：

```c
void lept_free(lept_value* v) {
    assert(v != NULL);
    if (v->type == LEPT_STRING)
        free(v->u.s.s);
    else if (v->type == LEPT_ARRAY) {
        for (size_t i = 0; i < v->u.a.size; i++)
            lept_free(&v->u.a.e[i]);
        free(v->u.a.e);
    }
    v->type = LEPT_NULL;
}
```

## 5.5 增加一个错误码

在解析数组时，我们增加了一个新的错误码 `LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET`，表示在解析数组时，期望 `,` 或 `]` 但实际却不是的情况。我们需要在 `leptjson.h` 中定义这个错误码：

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
    LEPT_PARSE_INVALID_UNICODE_SURROGATE,
    LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET
} lept_parse_error;
```

## 5.6 测试数组解析

现在我们来编写测试用例，验证数组解析的功能：

```c
static void test_parse_array() {
    lept_value v;
    
    /* 测试空数组 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_OK, lept_parse(&v, "[ ]"));
    EXPECT_EQ_INT(LEPT_ARRAY, lept_get_type(&v));
    EXPECT_EQ_SIZE_T(0, lept_get_array_size(&v));
    lept_free(&v);
    
    /* 测试包含一个元素的数组 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_OK, lept_parse(&v, "[ null ]"));
    EXPECT_EQ_INT(LEPT_ARRAY, lept_get_type(&v));
    EXPECT_EQ_SIZE_T(1, lept_get_array_size(&v));
    EXPECT_EQ_INT(LEPT_NULL, lept_get_type(lept_get_array_element(&v, 0)));
    lept_free(&v);
    
    /* 测试包含多个元素的数组 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_OK, lept_parse(&v, "[ null , false , true , 123.456 , \"abc\" ]"));
    EXPECT_EQ_INT(LEPT_ARRAY, lept_get_type(&v));
    EXPECT_EQ_SIZE_T(5, lept_get_array_size(&v));
    EXPECT_EQ_INT(LEPT_NULL, lept_get_type(lept_get_array_element(&v, 0)));
    EXPECT_EQ_INT(LEPT_FALSE, lept_get_type(lept_get_array_element(&v, 1)));
    EXPECT_EQ_INT(LEPT_TRUE, lept_get_type(lept_get_array_element(&v, 2)));
    EXPECT_EQ_INT(LEPT_NUMBER, lept_get_type(lept_get_array_element(&v, 3)));
    EXPECT_EQ_DOUBLE(123.456, lept_get_number(lept_get_array_element(&v, 3)));
    EXPECT_EQ_INT(LEPT_STRING, lept_get_type(lept_get_array_element(&v, 4)));
    EXPECT_EQ_STRING("abc", lept_get_string(lept_get_array_element(&v, 4)), lept_get_string_length(lept_get_array_element(&v, 4)));
    lept_free(&v);
    
    /* 测试嵌套数组 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_OK, lept_parse(&v, "[ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]"));
    EXPECT_EQ_INT(LEPT_ARRAY, lept_get_type(&v));
    EXPECT_EQ_SIZE_T(4, lept_get_array_size(&v));
    for (size_t i = 0; i < 4; i++) {
        lept_value* a = lept_get_array_element(&v, i);
        EXPECT_EQ_INT(LEPT_ARRAY, lept_get_type(a));
        EXPECT_EQ_SIZE_T(i, lept_get_array_size(a));
        for (size_t j = 0; j < i; j++) {
            lept_value* e = lept_get_array_element(a, j);
            EXPECT_EQ_INT(LEPT_NUMBER, lept_get_type(e));
            EXPECT_EQ_DOUBLE((double)j, lept_get_number(e));
        }
    }
    lept_free(&v);
}

static void test_parse_error() {
    /* 已有的测试用例 */
    // ...
    
    /* 数组解析错误测试 */
    TEST_ERROR("[", LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET);
    TEST_ERROR("[1", LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET);
    TEST_ERROR("[1,", LEPT_PARSE_EXPECT_VALUE);
    TEST_ERROR("[1,]", LEPT_PARSE_INVALID_VALUE);
    TEST_ERROR("[1 2]", LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET);
}
```

## 5.7 内存管理的问题

在实现数组解析时，我们使用了动态内存分配，这可能会引入内存泄漏的问题。在 `lept_free` 函数中，我们需要释放数组中所有元素分配的内存，避免内存泄漏。

此外，在处理解析错误时，我们需要小心释放已分配的内存。例如，在 `lept_parse_array` 函数中，如果解析过程中出现错误，我们需要释放已解析的元素。

## 5.8 练习

1. 实现一个函数 `lept_set_array`，用于设置一个值为数组类型：

```c
void lept_set_array(lept_value* v, size_t capacity);
```

2. 实现一个函数 `lept_reserve_array`，用于扩展数组的容量：

```c
void lept_reserve_array(lept_value* v, size_t capacity);
```

3. 实现一个函数 `lept_shrink_array`，用于收缩数组的容量：

```c
void lept_shrink_array(lept_value* v);
```

4. 实现一个函数 `lept_clear_array`，用于清空数组中的元素：

```c
void lept_clear_array(lept_value* v);
```

5. 实现一个函数 `lept_push_array_element`，用于在数组尾部添加一个元素：

```c
void lept_push_array_element(lept_value* v, const lept_value* e);
```

6. 实现一个函数 `lept_pop_array_element`，用于删除数组尾部的元素并返回它：

```c
void lept_pop_array_element(lept_value* v);
```

7. 实现一个函数 `lept_insert_array_element`，用于在数组中插入一个元素：

```c
void lept_insert_array_element(lept_value* v, const lept_value* e, size_t index);
```

8. 实现一个函数 `lept_erase_array_element`，用于删除数组中的一个元素：

```c
void lept_erase_array_element(lept_value* v, size_t index, size_t count);
```

## 5.9 下一步

在本章中，我们实现了 JSON 数组的解析。在下一章中，我们将继续实现 JSON 对象的解析，这是 JSON 的另一个重要数据结构。 