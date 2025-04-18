# 第10章：JSON 指针

在前面的章节中，我们已经实现了一个功能完整的 JSON 库，包括解析、生成、访问和修改 JSON 数据的功能。在本章中，我们将实现 JSON 指针，这是一种标准化的方法，用于在 JSON 文档中定位特定的值。

## 10.1 JSON 指针简介

[JSON 指针](https://tools.ietf.org/html/rfc6901)（JSON Pointer）是一个在 [RFC 6901](https://tools.ietf.org/html/rfc6901) 中定义的标准，用于在 JSON 文档中定位特定的值。它提供了一种简单的语法来表示 JSON 文档中的位置。

JSON 指针的语法很简单：它由一个可选的 `#` 字符（用于 URI 片段标识符）开始，后面跟着一系列由 `/` 分隔的引用令牌组成。每个令牌指定一个对象成员或数组元素。例如：

- `""` 指向整个 JSON 文档
- `"/foo"` 指向对象的 `foo` 成员
- `"/foo/0"` 指向 `foo` 数组的第一个元素
- `"/foo/0/bar"` 指向 `foo` 数组第一个元素的 `bar` 成员
- `"/~1"` 指向 `~` 成员（`~1` 是 `~` 的转义形式）
- `"/~0"` 指向 `/` 成员（`~0` 是 `/` 的转义形式）

在本章中，我们将实现 JSON 指针的解析和应用功能，使用户能够通过 JSON 指针定位和操作 JSON 文档中的值。

## 10.2 JSON 指针的解析

首先，我们需要解析 JSON 指针字符串，将其转换为一系列令牌。我们定义一个函数 `lept_parse_pointer` 来完成这个任务：

```c
typedef struct {
    char** tokens;
    size_t count;
} lept_pointer;

int lept_parse_pointer(const char* str, lept_pointer* pointer) {
    size_t i, j, token_len, count = 0;
    const char* p = str;
    char** tokens;
    
    /* 计算令牌数量 */
    if (*p == '#')
        p++;
    if (*p == '/')
        count++;
    for (i = 0; p[i]; i++)
        if (p[i] == '/')
            count++;
    
    /* 分配令牌数组 */
    tokens = (char**)malloc(count * sizeof(char*));
    if (tokens == NULL)
        return LEPT_POINTER_OUT_OF_MEMORY;
    
    /* 解析令牌 */
    p = str;
    if (*p == '#')
        p++;
    
    j = 0;
    while (*p) {
        if (*p == '/') {
            p++;
            i = 0;
            while (p[i] && p[i] != '/')
                i++;
            token_len = i;
            
            tokens[j] = (char*)malloc(token_len + 1);
            if (tokens[j] == NULL) {
                /* 释放已分配的令牌 */
                while (j > 0)
                    free(tokens[--j]);
                free(tokens);
                return LEPT_POINTER_OUT_OF_MEMORY;
            }
            
            /* 处理转义字符 */
            for (i = 0; i < token_len; i++) {
                if (p[i] == '~' && i + 1 < token_len) {
                    if (p[i + 1] == '0')
                        tokens[j][i] = '/';
                    else if (p[i + 1] == '1')
                        tokens[j][i] = '~';
                    else
                        tokens[j][i] = p[i];
                    i++;
                }
                else
                    tokens[j][i] = p[i];
            }
            tokens[j][token_len] = '\0';
            
            j++;
            p += token_len;
        }
        else
            p++;
    }
    
    pointer->tokens = tokens;
    pointer->count = count;
    
    return LEPT_POINTER_OK;
}
```

这个函数解析 JSON 指针字符串，将其分解为一系列令牌，并处理转义字符。它返回一个 `lept_pointer` 结构，其中包含令牌数组和令牌数量。

## 10.3 通过 JSON 指针获取值

有了 JSON 指针的解析函数，我们可以实现一个函数，用于通过 JSON 指针获取 JSON 值：

```c
lept_value* lept_get_by_pointer(lept_value* v, const lept_pointer* pointer) {
    lept_value* target = v;
    size_t i;
    
    for (i = 0; i < pointer->count; i++) {
        const char* token = pointer->tokens[i];
        
        if (target->type == LEPT_OBJECT) {
            /* 在对象中查找 */
            size_t j;
            size_t len = strlen(token);
            int found = 0;
            
            for (j = 0; j < target->u.o.size; j++) {
                if (target->u.o.m[j].klen == len && memcmp(target->u.o.m[j].k, token, len) == 0) {
                    target = &target->u.o.m[j].v;
                    found = 1;
                    break;
                }
            }
            
            if (!found)
                return NULL;
        }
        else if (target->type == LEPT_ARRAY) {
            /* 在数组中查找 */
            size_t index = 0;
            size_t j;
            
            /* 将令牌转换为索引 */
            if (token[0] == '-' && token[1] == '\0') {
                index = target->u.a.size;
            }
            else {
                for (j = 0; token[j]; j++) {
                    if (token[j] < '0' || token[j] > '9')
                        return NULL;
                    index = index * 10 + (token[j] - '0');
                }
            }
            
            if (index >= target->u.a.size)
                return NULL;
            
            target = &target->u.a.e[index];
        }
        else {
            /* 不是对象或数组，无法继续解析 */
            return NULL;
        }
    }
    
    return target;
}
```

这个函数遍历 JSON 指针中的每个令牌，根据令牌类型和目标值类型进行相应的操作，最终返回指向目标值的指针。

## 10.4 通过 JSON 指针设置值

除了获取值外，我们还需要一个函数，用于通过 JSON 指针设置 JSON 值：

```c
int lept_set_by_pointer(lept_value* v, const lept_pointer* pointer, const lept_value* value) {
    lept_value* target = v;
    size_t i;
    
    /* 空指针直接替换整个值 */
    if (pointer->count == 0) {
        lept_free(v);
        lept_copy(v, value);
        return LEPT_POINTER_OK;
    }
    
    /* 处理前面的令牌，直到最后一个 */
    for (i = 0; i < pointer->count - 1; i++) {
        const char* token = pointer->tokens[i];
        
        if (target->type == LEPT_OBJECT) {
            /* 在对象中查找 */
            size_t j;
            size_t len = strlen(token);
            int found = 0;
            
            for (j = 0; j < target->u.o.size; j++) {
                if (target->u.o.m[j].klen == len && memcmp(target->u.o.m[j].k, token, len) == 0) {
                    target = &target->u.o.m[j].v;
                    found = 1;
                    break;
                }
            }
            
            if (!found) {
                /* 如果成员不存在，创建一个 */
                lept_value new_value;
                lept_init(&new_value);
                
                /* 确定新值的类型（对象或数组，根据下一个令牌） */
                if (i + 1 < pointer->count) {
                    const char* next_token = pointer->tokens[i + 1];
                    if (next_token[0] >= '0' && next_token[0] <= '9')
                        lept_set_array(&new_value, 0);
                    else
                        lept_set_object(&new_value, 0);
                }
                else {
                    /* 最后一个令牌，将其设置为 null */
                    lept_set_null(&new_value);
                }
                
                /* 添加新成员 */
                lept_set_object_value(target, token, len, &new_value);
                lept_free(&new_value);
                
                /* 获取新添加的成员的值 */
                target = lept_get_object_value(target, lept_get_object_size(target) - 1);
            }
        }
        else if (target->type == LEPT_ARRAY) {
            /* 在数组中查找 */
            size_t index = 0;
            size_t j;
            
            /* 将令牌转换为索引 */
            if (token[0] == '-' && token[1] == '\0') {
                index = target->u.a.size;
            }
            else {
                for (j = 0; token[j]; j++) {
                    if (token[j] < '0' || token[j] > '9')
                        return LEPT_POINTER_INVALID_TOKEN;
                    index = index * 10 + (token[j] - '0');
                }
            }
            
            if (index > target->u.a.size)
                return LEPT_POINTER_ARRAY_INDEX_OUT_OF_RANGE;
            
            if (index == target->u.a.size) {
                /* 如果索引等于数组大小，添加一个新元素 */
                lept_value new_value;
                lept_init(&new_value);
                
                /* 确定新值的类型（对象或数组，根据下一个令牌） */
                if (i + 1 < pointer->count) {
                    const char* next_token = pointer->tokens[i + 1];
                    if (next_token[0] >= '0' && next_token[0] <= '9')
                        lept_set_array(&new_value, 0);
                    else
                        lept_set_object(&new_value, 0);
                }
                else {
                    /* 最后一个令牌，将其设置为 null */
                    lept_set_null(&new_value);
                }
                
                /* 添加新元素 */
                lept_push_array_element(target, &new_value);
                lept_free(&new_value);
                
                /* 获取新添加的元素 */
                target = &target->u.a.e[index];
            }
            else {
                target = &target->u.a.e[index];
            }
        }
        else {
            /* 不是对象或数组，无法继续解析 */
            return LEPT_POINTER_NOT_FOUND;
        }
    }
    
    /* 处理最后一个令牌 */
    const char* token = pointer->tokens[pointer->count - 1];
    
    if (target->type == LEPT_OBJECT) {
        size_t len = strlen(token);
        size_t j;
        int found = 0;
        
        /* 查找成员 */
        for (j = 0; j < target->u.o.size; j++) {
            if (target->u.o.m[j].klen == len && memcmp(target->u.o.m[j].k, token, len) == 0) {
                lept_value* member_value = &target->u.o.m[j].v;
                lept_free(member_value);
                lept_copy(member_value, value);
                found = 1;
                break;
            }
        }
        
        if (!found) {
            /* 如果成员不存在，添加一个新成员 */
            lept_set_object_value(target, token, len, value);
        }
    }
    else if (target->type == LEPT_ARRAY) {
        size_t index = 0;
        size_t j;
        
        /* 将令牌转换为索引 */
        if (token[0] == '-' && token[1] == '\0') {
            index = target->u.a.size;
        }
        else {
            for (j = 0; token[j]; j++) {
                if (token[j] < '0' || token[j] > '9')
                    return LEPT_POINTER_INVALID_TOKEN;
                index = index * 10 + (token[j] - '0');
            }
        }
        
        if (index > target->u.a.size)
            return LEPT_POINTER_ARRAY_INDEX_OUT_OF_RANGE;
        
        if (index == target->u.a.size) {
            /* 如果索引等于数组大小，添加一个新元素 */
            lept_push_array_element(target, value);
        }
        else {
            /* 否则，替换现有元素 */
            lept_value* element = &target->u.a.e[index];
            lept_free(element);
            lept_copy(element, value);
        }
    }
    else {
        /* 不是对象或数组，无法设置值 */
        return LEPT_POINTER_NOT_FOUND;
    }
    
    return LEPT_POINTER_OK;
}
```

这个函数通过 JSON 指针设置 JSON 值。它首先导航到指针所指向的位置，然后根据目标位置的类型（对象或数组）进行相应的操作。如果路径中的某些部分不存在，它会创建必要的对象或数组。

## 10.5 通过 JSON 指针删除值

接下来，我们实现一个函数，用于通过 JSON 指针删除 JSON 值：

```c
int lept_remove_by_pointer(lept_value* v, const lept_pointer* pointer) {
    lept_value* target = v;
    size_t i;
    
    /* 空指针无法删除整个文档 */
    if (pointer->count == 0)
        return LEPT_POINTER_INVALID;
    
    /* 处理前面的令牌，直到最后一个 */
    for (i = 0; i < pointer->count - 1; i++) {
        const char* token = pointer->tokens[i];
        
        if (target->type == LEPT_OBJECT) {
            /* 在对象中查找 */
            size_t j;
            size_t len = strlen(token);
            int found = 0;
            
            for (j = 0; j < target->u.o.size; j++) {
                if (target->u.o.m[j].klen == len && memcmp(target->u.o.m[j].k, token, len) == 0) {
                    target = &target->u.o.m[j].v;
                    found = 1;
                    break;
                }
            }
            
            if (!found)
                return LEPT_POINTER_NOT_FOUND;
        }
        else if (target->type == LEPT_ARRAY) {
            /* 在数组中查找 */
            size_t index = 0;
            size_t j;
            
            /* 将令牌转换为索引 */
            for (j = 0; token[j]; j++) {
                if (token[j] < '0' || token[j] > '9')
                    return LEPT_POINTER_INVALID_TOKEN;
                index = index * 10 + (token[j] - '0');
            }
            
            if (index >= target->u.a.size)
                return LEPT_POINTER_NOT_FOUND;
            
            target = &target->u.a.e[index];
        }
        else {
            /* 不是对象或数组，无法继续解析 */
            return LEPT_POINTER_NOT_FOUND;
        }
    }
    
    /* 处理最后一个令牌 */
    const char* token = pointer->tokens[pointer->count - 1];
    
    if (target->type == LEPT_OBJECT) {
        size_t len = strlen(token);
        size_t j;
        
        /* 查找并删除成员 */
        for (j = 0; j < target->u.o.size; j++) {
            if (target->u.o.m[j].klen == len && memcmp(target->u.o.m[j].k, token, len) == 0) {
                lept_remove_object_value(target, j);
                return LEPT_POINTER_OK;
            }
        }
        
        return LEPT_POINTER_NOT_FOUND;
    }
    else if (target->type == LEPT_ARRAY) {
        size_t index = 0;
        size_t j;
        
        /* 将令牌转换为索引 */
        for (j = 0; token[j]; j++) {
            if (token[j] < '0' || token[j] > '9')
                return LEPT_POINTER_INVALID_TOKEN;
            index = index * 10 + (token[j] - '0');
        }
        
        if (index >= target->u.a.size)
            return LEPT_POINTER_NOT_FOUND;
        
        /* 删除数组元素 */
        lept_erase_array_element(target, index, 1);
        return LEPT_POINTER_OK;
    }
    else {
        /* 不是对象或数组，无法删除值 */
        return LEPT_POINTER_NOT_FOUND;
    }
}
```

这个函数通过 JSON 指针删除 JSON 值。它首先导航到指针所指向的位置，然后根据目标位置的类型（对象或数组）删除相应的成员或元素。

## 10.6 释放 JSON 指针

当我们不再需要 JSON 指针时，我们需要释放它占用的内存：

```c
void lept_free_pointer(lept_pointer* pointer) {
    size_t i;
    
    for (i = 0; i < pointer->count; i++)
        free(pointer->tokens[i]);
    
    free(pointer->tokens);
    
    pointer->tokens = NULL;
    pointer->count = 0;
}
```

## 10.7 封装 JSON 指针操作

为了方便使用，我们可以封装一些函数，直接使用 JSON 指针字符串操作 JSON 值：

```c
lept_value* lept_get_by_pointer_str(lept_value* v, const char* pointer_str) {
    lept_pointer pointer;
    lept_value* result;
    
    if (lept_parse_pointer(pointer_str, &pointer) != LEPT_POINTER_OK)
        return NULL;
    
    result = lept_get_by_pointer(v, &pointer);
    
    lept_free_pointer(&pointer);
    
    return result;
}

int lept_set_by_pointer_str(lept_value* v, const char* pointer_str, const lept_value* value) {
    lept_pointer pointer;
    int result;
    
    if (lept_parse_pointer(pointer_str, &pointer) != LEPT_POINTER_OK)
        return LEPT_POINTER_INVALID;
    
    result = lept_set_by_pointer(v, &pointer, value);
    
    lept_free_pointer(&pointer);
    
    return result;
}

int lept_remove_by_pointer_str(lept_value* v, const char* pointer_str) {
    lept_pointer pointer;
    int result;
    
    if (lept_parse_pointer(pointer_str, &pointer) != LEPT_POINTER_OK)
        return LEPT_POINTER_INVALID;
    
    result = lept_remove_by_pointer(v, &pointer);
    
    lept_free_pointer(&pointer);
    
    return result;
}
```

这些函数直接接受 JSON 指针字符串，解析它，执行相应的操作，然后释放 JSON 指针。

## 10.8 测试 JSON 指针

为了测试我们的 JSON 指针实现，我们可以编写一些测试用例：

```c
static void test_pointer() {
    lept_value v;
    lept_value* result;
    const char* json = "{\"foo\":[\"bar\",\"baz\"],\"\":0,\"a/b\":1,\"c%d\":2,\"e^f\":3,\"g|h\":4,\"i\\\\j\":5,\"k\\\"l\":6,\" \":7,\"m~n\":8}";
    
    /* 解析 JSON */
    lept_init(&v);
    lept_parse(&v, json);
    
    /* 测试获取值 */
    result = lept_get_by_pointer_str(&v, "");
    EXPECT_EQ_INT(LEPT_OBJECT, result->type);
    
    result = lept_get_by_pointer_str(&v, "/foo");
    EXPECT_EQ_INT(LEPT_ARRAY, result->type);
    EXPECT_EQ_SIZE_T(2, result->u.a.size);
    
    result = lept_get_by_pointer_str(&v, "/foo/0");
    EXPECT_EQ_INT(LEPT_STRING, result->type);
    EXPECT_EQ_STRING("bar", result->u.s.s, result->u.s.len);
    
    result = lept_get_by_pointer_str(&v, "/foo/1");
    EXPECT_EQ_INT(LEPT_STRING, result->type);
    EXPECT_EQ_STRING("baz", result->u.s.s, result->u.s.len);
    
    result = lept_get_by_pointer_str(&v, "/");
    EXPECT_EQ_INT(LEPT_NUMBER, result->type);
    EXPECT_EQ_DOUBLE(0.0, result->u.n);
    
    result = lept_get_by_pointer_str(&v, "/a~1b");
    EXPECT_EQ_INT(LEPT_NUMBER, result->type);
    EXPECT_EQ_DOUBLE(1.0, result->u.n);
    
    result = lept_get_by_pointer_str(&v, "/m~0n");
    EXPECT_EQ_INT(LEPT_NUMBER, result->type);
    EXPECT_EQ_DOUBLE(8.0, result->u.n);
    
    /* 测试设置值 */
    lept_value new_value;
    lept_init(&new_value);
    lept_set_string(&new_value, "qux", 3);
    
    EXPECT_EQ_INT(LEPT_POINTER_OK, lept_set_by_pointer_str(&v, "/foo/1", &new_value));
    result = lept_get_by_pointer_str(&v, "/foo/1");
    EXPECT_EQ_INT(LEPT_STRING, result->type);
    EXPECT_EQ_STRING("qux", result->u.s.s, result->u.s.len);
    
    lept_set_number(&new_value, 42.0);
    EXPECT_EQ_INT(LEPT_POINTER_OK, lept_set_by_pointer_str(&v, "/new", &new_value));
    result = lept_get_by_pointer_str(&v, "/new");
    EXPECT_EQ_INT(LEPT_NUMBER, result->type);
    EXPECT_EQ_DOUBLE(42.0, result->u.n);
    
    /* 测试删除值 */
    EXPECT_EQ_INT(LEPT_POINTER_OK, lept_remove_by_pointer_str(&v, "/foo/0"));
    result = lept_get_by_pointer_str(&v, "/foo/0");
    EXPECT_EQ_INT(LEPT_STRING, result->type);
    EXPECT_EQ_STRING("qux", result->u.s.s, result->u.s.len);
    
    EXPECT_EQ_INT(LEPT_POINTER_OK, lept_remove_by_pointer_str(&v, "/new"));
    result = lept_get_by_pointer_str(&v, "/new");
    EXPECT_NULL(result);
    
    lept_free(&new_value);
    lept_free(&v);
}
```

这些测试用例验证了我们的 JSON 指针实现的各种功能，包括获取、设置和删除值。

## 10.9 JSON 指针的应用

JSON 指针在许多场景中都非常有用，例如：

1. **API 设计**：在 REST API 中，可以使用 JSON 指针来指定要操作的资源的特定部分。

2. **数据验证**：在验证 JSON 数据时，可以使用 JSON 指针来指示验证错误的位置。

3. **数据修改**：在需要修改 JSON 数据的特定部分时，JSON 指针提供了一种标准化的方法来定位和修改数据。

4. **数据查询**：在查询 JSON 数据时，JSON 指针可以用于指定要检索的数据的路径。

## 10.10 练习

1. 实现一个函数 `lept_create_pointer`，用于创建 JSON 指针。

2. 实现一个函数 `lept_compare_pointer`，用于比较两个 JSON 指针。

3. 实现一个函数 `lept_escape_pointer_token`，用于转义 JSON 指针令牌。

4. 实现一个函数 `lept_unescape_pointer_token`，用于解除 JSON 指针令牌的转义。

5. 实现一个函数 `lept_pointer_to_string`，用于将 JSON 指针转换为字符串。

## 10.11 下一步

在本章中，我们实现了 JSON 指针，使用户能够通过标准化的方法在 JSON 文档中定位和操作特定的值。JSON 指针是 JSON 相关标准中的一个重要组成部分，它为 JSON 数据的操作提供了更大的灵活性。

在下一章中，我们将探讨另一个重要的 JSON 相关标准：JSON Schema。JSON Schema 是一种用于描述和验证 JSON 数据结构的规范，它可以帮助我们确保 JSON 数据符合预期的格式和约束。