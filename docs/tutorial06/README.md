# 第6章：对象解析

在上一章中，我们实现了 JSON 数组的解析。本章我们将实现 JSON 对象的解析，这是 JSON 中最重要的数据结构。

## 6.1 JSON 对象

JSON 对象是一个无序的键值对集合，使用花括号 `{}` 表示，键值对之间用逗号 `,` 分隔。键是字符串，值可以是任意 JSON 数据类型，包括数字、字符串、布尔值、null、数组和对象。

例如：
```json
{
  "name": "张三",
  "age": 30,
  "isStudent": false,
  "courses": ["数学", "物理", "计算机科学"],
  "address": {
    "city": "北京",
    "zip": "100000"
  }
}
```

一个对象可以为空，表示为 `{}`。

## 6.2 修改数据结构

首先，我们需要在 `lept_value` 结构体中添加对象类型的支持。为了存储对象的键值对，我们定义一个新的结构体 `lept_member`，修改 `leptjson.h` 如下：

```c
typedef struct lept_member lept_member;
typedef struct lept_value lept_value;

struct lept_member {
    char* k;        /* member key string */
    size_t klen;    /* key string length */
    lept_value v;   /* member value */
};

struct lept_value {
    union {
        struct { lept_member* m; size_t size; }o; /* object */
        struct { lept_value* e; size_t size; }a;  /* array */
        struct { char* s; size_t len; }s;         /* string */
        double n;                                 /* number */
    } u;
    lept_type type;
};
```

上面的代码中，我们添加了一个结构体 `o` 用于表示对象，其中：
- `m` 是指向成员数组的指针
- `size` 是成员的个数

## 6.3 访问对象的接口

在实现对象解析之前，我们先定义一些用于访问对象的函数：

```c
size_t lept_get_object_size(const lept_value* v);
const char* lept_get_object_key(const lept_value* v, size_t index);
size_t lept_get_object_key_length(const lept_value* v, size_t index);
lept_value* lept_get_object_value(const lept_value* v, size_t index);
```

这些函数的实现也很简单：

```c
size_t lept_get_object_size(const lept_value* v) {
    assert(v != NULL && v->type == LEPT_OBJECT);
    return v->u.o.size;
}

const char* lept_get_object_key(const lept_value* v, size_t index) {
    assert(v != NULL && v->type == LEPT_OBJECT && index < v->u.o.size);
    return v->u.o.m[index].k;
}

size_t lept_get_object_key_length(const lept_value* v, size_t index) {
    assert(v != NULL && v->type == LEPT_OBJECT && index < v->u.o.size);
    return v->u.o.m[index].klen;
}

lept_value* lept_get_object_value(const lept_value* v, size_t index) {
    assert(v != NULL && v->type == LEPT_OBJECT && index < v->u.o.size);
    return &v->u.o.m[index].v;
}
```

## 6.4 解析 JSON 对象

现在我们来实现对象的解析。我们需要增加一个新的函数 `lept_parse_object`：

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
            ret = LEPT_PARSE_MISS_COLON;
            break;
        }
        c->json++;
        lept_parse_whitespace(c);
        /* parse value */
        if ((ret = lept_parse_value(c, &m.v)) != LEPT_PARSE_OK)
            break;
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
            ret = LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET;
            break;
        }
    }
    /* Pop and free members on the stack */
    free(m.k);
    for (int i = 0; i < size; i++) {
        lept_member* m = (lept_member*)lept_context_pop(c, sizeof(lept_member));
        free(m->k);
        lept_free(&m->v);
    }
    return ret;
}
```

解析对象的流程与解析数组类似：
1. 判断是否为空对象 `{}`，若是则设置 `size` 为 0，`m` 为 NULL，并返回
2. 循环解析对象中的键值对：
   - 解析键（必须是字符串）
   - 解析冒号 `:`
   - 解析值
   - 将键值对压入堆栈
   - 判断是否有逗号 `,`，若有则继续解析下一个键值对
   - 若遇到右花括号 `}`，则分配内存存储所有键值对，并返回
   - 若解析失败或格式错误，则释放已解析的键值对并返回错误

同时，我们还需要在 `lept_parse_value` 函数中增加对对象的处理：

```c
static int lept_parse_value(lept_context* c, lept_value* v) {
    switch (*c->json) {
        case 'n':  return lept_parse_null(c, v);
        case 't':  return lept_parse_true(c, v);
        case 'f':  return lept_parse_false(c, v);
        case '"':  return lept_parse_string(c, v);
        case '[':  return lept_parse_array(c, v);
        case '{':  return lept_parse_object(c, v);
        case '\0': return LEPT_PARSE_EXPECT_VALUE;
        default:   return lept_parse_number(c, v);
    }
}
```

另外，我们还需要修改 `lept_free` 函数，以释放对象中的成员：

```c
void lept_free(lept_value* v) {
    assert(v != NULL);
    switch (v->type) {
        case LEPT_STRING:
            free(v->u.s.s);
            break;
        case LEPT_ARRAY:
            for (size_t i = 0; i < v->u.a.size; i++)
                lept_free(&v->u.a.e[i]);
            free(v->u.a.e);
            break;
        case LEPT_OBJECT:
            for (size_t i = 0; i < v->u.o.size; i++) {
                free(v->u.o.m[i].k);
                lept_free(&v->u.o.m[i].v);
            }
            free(v->u.o.m);
            break;
        default: break;
    }
    v->type = LEPT_NULL;
}
```

## 6.5 增加错误码

在解析对象时，我们增加了几个新的错误码，表示在解析对象时可能遇到的错误。我们需要在 `leptjson.h` 中定义这些错误码：

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
    LEPT_PARSE_MISS_COMMA_OR_SQUARE_BRACKET,
    LEPT_PARSE_MISS_KEY,
    LEPT_PARSE_MISS_COLON,
    LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET
} lept_parse_error;
```

## 6.6 重构字符串解析

在解析对象的键时，我们需要解析字符串，并将结果存储在临时变量中。为了避免代码重复，我们可以将字符串解析的功能分离出来，定义一个新的函数 `lept_parse_string_raw`：

```c
static int lept_parse_string_raw(lept_context* c, char** str, size_t* len) {
    size_t head = c->top;
    unsigned u, u2;
    const char* p;
    EXPECT(c, '\"');
    p = c->json;
    for (;;) {
        char ch = *p++;
        switch (ch) {
            case '\"':
                *len = c->top - head;
                *str = lept_context_pop(c, *len);
                c->json = p;
                return LEPT_PARSE_OK;
            case '\\':
                switch (*p++) {
                    case '\"': PUTC(c, '\"'); break;
                    case '\\': PUTC(c, '\\'); break;
                    case '/':  PUTC(c, '/');  break;
                    case 'b':  PUTC(c, '\b'); break;
                    case 'f':  PUTC(c, '\f'); break;
                    case 'n':  PUTC(c, '\n'); break;
                    case 'r':  PUTC(c, '\r'); break;
                    case 't':  PUTC(c, '\t'); break;
                    case 'u':
                        if (!(p = lept_parse_hex4(p, &u)))
                            STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_HEX);
                        if (u >= 0xD800 && u <= 0xDBFF) { /* surrogate pair */
                            if (*p++ != '\\')
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_SURROGATE);
                            if (*p++ != 'u')
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_SURROGATE);
                            if (!(p = lept_parse_hex4(p, &u2)))
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_HEX);
                            if (u2 < 0xDC00 || u2 > 0xDFFF)
                                STRING_ERROR(LEPT_PARSE_INVALID_UNICODE_SURROGATE);
                            u = (((u - 0xD800) << 10) | (u2 - 0xDC00)) + 0x10000;
                        }
                        lept_encode_utf8(c, u);
                        break;
                    default:
                        STRING_ERROR(LEPT_PARSE_INVALID_STRING_ESCAPE);
                }
                break;
            case '\0':
                STRING_ERROR(LEPT_PARSE_MISS_QUOTATION_MARK);
            default:
                if ((unsigned char)ch < 0x20)
                    STRING_ERROR(LEPT_PARSE_INVALID_STRING_CHAR);
                PUTC(c, ch);
        }
    }
}

static int lept_parse_string(lept_context* c, lept_value* v) {
    int ret;
    char* s;
    size_t len;
    if ((ret = lept_parse_string_raw(c, &s, &len)) == LEPT_PARSE_OK)
        lept_set_string(v, s, len);
    return ret;
}
```

这样，我们就可以在解析对象键和字符串值时复用同一个函数。

## 6.7 测试对象解析

现在我们来编写测试用例，验证对象解析的功能：

```c
static void test_parse_object() {
    lept_value v;
    
    /* 测试空对象 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_OK, lept_parse(&v, " { } "));
    EXPECT_EQ_INT(LEPT_OBJECT, lept_get_type(&v));
    EXPECT_EQ_SIZE_T(0, lept_get_object_size(&v));
    lept_free(&v);
    
    /* 测试简单对象 */
    lept_init(&v);
    EXPECT_EQ_INT(LEPT_PARSE_OK, lept_parse(&v, 
        " { "
        "\"n\" : null , "
        "\"f\" : false , "
        "\"t\" : true , "
        "\"i\" : 123 , "
        "\"s\" : \"abc\", "
        "\"a\" : [ 1, 2, 3 ],"
        "\"o\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }"
        " } "
    ));
    EXPECT_EQ_INT(LEPT_OBJECT, lept_get_type(&v));
    EXPECT_EQ_SIZE_T(7, lept_get_object_size(&v));
    EXPECT_EQ_STRING("n", lept_get_object_key(&v, 0), lept_get_object_key_length(&v, 0));
    EXPECT_EQ_INT(LEPT_NULL, lept_get_type(lept_get_object_value(&v, 0)));
    EXPECT_EQ_STRING("f", lept_get_object_key(&v, 1), lept_get_object_key_length(&v, 1));
    EXPECT_EQ_INT(LEPT_FALSE, lept_get_type(lept_get_object_value(&v, 1)));
    EXPECT_EQ_STRING("t", lept_get_object_key(&v, 2), lept_get_object_key_length(&v, 2));
    EXPECT_EQ_INT(LEPT_TRUE, lept_get_type(lept_get_object_value(&v, 2)));
    EXPECT_EQ_STRING("i", lept_get_object_key(&v, 3), lept_get_object_key_length(&v, 3));
    EXPECT_EQ_INT(LEPT_NUMBER, lept_get_type(lept_get_object_value(&v, 3)));
    EXPECT_EQ_DOUBLE(123.0, lept_get_number(lept_get_object_value(&v, 3)));
    EXPECT_EQ_STRING("s", lept_get_object_key(&v, 4), lept_get_object_key_length(&v, 4));
    EXPECT_EQ_INT(LEPT_STRING, lept_get_type(lept_get_object_value(&v, 4)));
    EXPECT_EQ_STRING("abc", lept_get_string(lept_get_object_value(&v, 4)), lept_get_string_length(lept_get_object_value(&v, 4)));
    EXPECT_EQ_STRING("a", lept_get_object_key(&v, 5), lept_get_object_key_length(&v, 5));
    EXPECT_EQ_INT(LEPT_ARRAY, lept_get_type(lept_get_object_value(&v, 5)));
    EXPECT_EQ_SIZE_T(3, lept_get_array_size(lept_get_object_value(&v, 5)));
    for (int i = 0; i < 3; i++) {
        lept_value* e = lept_get_array_element(lept_get_object_value(&v, 5), i);
        EXPECT_EQ_INT(LEPT_NUMBER, lept_get_type(e));
        EXPECT_EQ_DOUBLE(i + 1.0, lept_get_number(e));
    }
    EXPECT_EQ_STRING("o", lept_get_object_key(&v, 6), lept_get_object_key_length(&v, 6));
    {
        lept_value* o = lept_get_object_value(&v, 6);
        EXPECT_EQ_INT(LEPT_OBJECT, lept_get_type(o));
        EXPECT_EQ_SIZE_T(3, lept_get_object_size(o));
        for (int i = 0; i < 3; i++) {
            lept_value* ov = lept_get_object_value(o, i);
            EXPECT_EQ_TRUE('1' + i == lept_get_object_key(o, i)[0]);
            EXPECT_EQ_SIZE_T(1, lept_get_object_key_length(o, i));
            EXPECT_EQ_INT(LEPT_NUMBER, lept_get_type(ov));
            EXPECT_EQ_DOUBLE(i + 1.0, lept_get_number(ov));
        }
    }
    lept_free(&v);
}

static void test_parse_error() {
    /* 已有的测试用例 */
    // ...
    
    /* 对象解析错误测试 */
    TEST_ERROR("{", LEPT_PARSE_MISS_KEY);
    TEST_ERROR("{]", LEPT_PARSE_MISS_KEY);
    TEST_ERROR("{1", LEPT_PARSE_MISS_KEY);
    TEST_ERROR("{\"a\"", LEPT_PARSE_MISS_COLON);
    TEST_ERROR("{\"a\":", LEPT_PARSE_EXPECT_VALUE);
    TEST_ERROR("{\"a\":1", LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET);
    TEST_ERROR("{\"a\":1]", LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET);
    TEST_ERROR("{\"a\":1 \"b\"", LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET);
    TEST_ERROR("{\"a\":{}", LEPT_PARSE_MISS_COMMA_OR_CURLY_BRACKET);
}
```

## 6.8 关于对象键查找的优化

在实际应用中，我们经常需要根据键名查找对象中的值。但目前我们的实现只提供了通过索引访问的接口，如果要查找特定键名的值，需要遍历所有键：

```c
lept_value* lept_find_object_value(const lept_value* v, const char* key, size_t klen) {
    size_t i;
    assert(v != NULL && v->type == LEPT_OBJECT && key != NULL);
    for (i = 0; i < v->u.o.size; i++)
        if (v->u.o.m[i].klen == klen && memcmp(v->u.o.m[i].k, key, klen) == 0)
            return &v->u.o.m[i].v;
    return NULL;
}
```

这种线性查找的时间复杂度为 O(n)，对于大型对象来说效率较低。在实际应用中，我们可以考虑使用哈希表等数据结构来加速查找。

## 6.9 练习

1. 实现一个函数 `lept_find_object_index`，用于查找对象中指定键名的索引：

```c
size_t lept_find_object_index(const lept_value* v, const char* key, size_t klen);
```

2. 实现一个函数 `lept_find_object_value`，用于查找对象中指定键名的值：

```c
lept_value* lept_find_object_value(const lept_value* v, const char* key, size_t klen);
```

3. 实现一个函数 `lept_set_object`，用于设置一个值为对象类型：

```c
void lept_set_object(lept_value* v, size_t capacity);
```

4. 实现一个函数 `lept_reserve_object`，用于扩展对象的容量：

```c
void lept_reserve_object(lept_value* v, size_t capacity);
```

5. 实现一个函数 `lept_shrink_object`，用于收缩对象的容量：

```c
void lept_shrink_object(lept_value* v);
```

6. 实现一个函数 `lept_clear_object`，用于清空对象中的成员：

```c
void lept_clear_object(lept_value* v);
```

7. 实现一个函数 `lept_set_object_value`，用于设置对象中指定键名的值：

```c
void lept_set_object_value(lept_value* v, const char* key, size_t klen, const lept_value* value);
```

8. 实现一个函数 `lept_remove_object_value`，用于删除对象中指定键名的值：

```c
void lept_remove_object_value(lept_value* v, size_t index);
```

## 6.10 下一步

在本章中，我们实现了 JSON 对象的解析。现在，我们的 JSON 解析器已经可以支持所有 JSON 数据类型（null、布尔值、数字、字符串、数组和对象）的解析了。在下一章中，我们将实现 JSON 生成器，将 `lept_value` 转换为 JSON 文本。 