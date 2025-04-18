# 第8章：访问与其他功能

在前面的章节中，我们已经实现了 JSON 的解析和生成功能。在本章中，我们将为我们的 JSON 库添加更多实用的功能，包括访问和修改 JSON 数据的接口，以及一些其他实用功能。

## 8.1 深度复制

首先，我们实现一个深度复制（deep copy）函数，用于创建一个 `lept_value` 的副本：

```c
void lept_copy(lept_value* dst, const lept_value* src) {
    assert(src != NULL && dst != NULL && src != dst);
    switch (src->type) {
        case LEPT_STRING:
            lept_set_string(dst, src->u.s.s, src->u.s.len);
            break;
        case LEPT_ARRAY:
            lept_set_array(dst, src->u.a.size);
            for (size_t i = 0; i < src->u.a.size; i++) {
                lept_value e;
                lept_init(&e);
                lept_copy(&e, &src->u.a.e[i]);
                lept_move(lept_push_array_element(dst), &e);
            }
            break;
        case LEPT_OBJECT:
            lept_set_object(dst, src->u.o.size);
            for (size_t i = 0; i < src->u.o.size; i++) {
                lept_value v;
                lept_init(&v);
                lept_copy(&v, &src->u.o.m[i].v);
                lept_move(lept_set_object_value(dst, src->u.o.m[i].k, src->u.o.m[i].klen), &v);
            }
            break;
        default:
            lept_free(dst);
            memcpy(dst, src, sizeof(lept_value));
            break;
    }
}
```

这个函数根据源值的类型，创建一个新的值，并递归地复制所有子元素。对于简单类型（null、布尔值、数字），我们直接进行内存拷贝；对于复杂类型（字符串、数组、对象），我们需要分配新的内存，并递归地复制内容。

我们还实现了一个移动（move）函数，用于将一个值移动到另一个位置：

```c
void lept_move(lept_value* dst, lept_value* src) {
    assert(dst != NULL && src != NULL && src != dst);
    lept_free(dst);
    memcpy(dst, src, sizeof(lept_value));
    lept_init(src);
}
```

移动操作比复制更高效，因为它避免了内存的分配和拷贝。

## 8.2 相等比较

接下来，我们实现一个相等比较函数，用于判断两个 `lept_value` 是否相等：

```c
int lept_is_equal(const lept_value* lhs, const lept_value* rhs) {
    assert(lhs != NULL && rhs != NULL);
    if (lhs->type != rhs->type)
        return 0;
    
    switch (lhs->type) {
        case LEPT_STRING:
            return lhs->u.s.len == rhs->u.s.len && 
                   memcmp(lhs->u.s.s, rhs->u.s.s, lhs->u.s.len) == 0;
        case LEPT_NUMBER:
            return lhs->u.n == rhs->u.n;
        case LEPT_ARRAY:
            if (lhs->u.a.size != rhs->u.a.size)
                return 0;
            for (size_t i = 0; i < lhs->u.a.size; i++)
                if (!lept_is_equal(&lhs->u.a.e[i], &rhs->u.a.e[i]))
                    return 0;
            return 1;
        case LEPT_OBJECT:
            if (lhs->u.o.size != rhs->u.o.size)
                return 0;
            for (size_t i = 0; i < lhs->u.o.size; i++) {
                size_t rhs_index = lept_find_object_index(rhs, lhs->u.o.m[i].k, lhs->u.o.m[i].klen);
                if (rhs_index == LEPT_KEY_NOT_FOUND)
                    return 0;
                if (!lept_is_equal(&lhs->u.o.m[i].v, &rhs->u.o.m[rhs_index].v))
                    return 0;
            }
            return 1;
        default:
            return 1;  /* null, true, false */
    }
}
```

这个函数首先检查两个值的类型是否相同，然后根据类型进行比较。对于字符串，我们比较长度和内容；对于数字，我们直接比较数值；对于数组，我们比较大小和每个元素；对于对象，我们比较大小和每个键值对。

## 8.3 访问嵌套值

在实际应用中，我们常常需要访问 JSON 中的嵌套值。例如，对于以下 JSON：

```json
{
  "name": "John",
  "age": 30,
  "address": {
    "city": "New York",
    "zip": "10001"
  },
  "skills": ["C", "C++", "JSON"]
}
```

我们可能需要获取 `address.city` 或 `skills[2]` 的值。为此，我们实现一个函数，用于根据路径获取嵌套值：

```c
lept_value* lept_get_value_by_path(lept_value* v, const char* path) {
    lept_value* cur = v;
    const char* p = path;
    size_t i, len;
    
    if (path == NULL || *path == '\0')
        return v;
    
    while (*p != '\0') {
        if (*p == '[') {  /* array */
            if (cur->type != LEPT_ARRAY)
                return NULL;
            
            p++;
            i = 0;
            while (*p >= '0' && *p <= '9')
                i = i * 10 + (*p++ - '0');
            
            if (*p++ != ']')
                return NULL;
            
            if (i >= cur->u.a.size)
                return NULL;
            
            cur = &cur->u.a.e[i];
        }
        else if (*p == '.') {  /* object */
            if (cur->type != LEPT_OBJECT)
                return NULL;
            
            p++;
            len = 0;
            while (p[len] != '\0' && p[len] != '.' && p[len] != '[')
                len++;
            
            for (i = 0; i < cur->u.o.size; i++)
                if (cur->u.o.m[i].klen == len && memcmp(cur->u.o.m[i].k, p, len) == 0)
                    break;
            
            if (i == cur->u.o.size)
                return NULL;
            
            cur = &cur->u.o.m[i].v;
            p += len;
        }
        else
            return NULL;
    }
    
    return cur;
}
```

这个函数从根值开始，根据路径一步步地访问嵌套值。路径中的 `[n]` 表示访问数组的第 n 个元素，`.key` 表示访问对象的 key 成员。例如，`address.city` 表示访问 address 对象的 city 成员；`skills[2]` 表示访问 skills 数组的第 3 个元素（索引从 0 开始）。

## 8.4 查询功能

在某些应用场景中，我们可能需要根据条件查询 JSON 数据。例如，查找满足特定条件的数组元素或对象成员。为此，我们可以实现一些查询函数：

```c
lept_value* lept_find_array_element(lept_value* v, lept_query_callback callback, void* userdata) {
    assert(v != NULL && v->type == LEPT_ARRAY);
    for (size_t i = 0; i < v->u.a.size; i++)
        if (callback(&v->u.a.e[i], userdata))
            return &v->u.a.e[i];
    return NULL;
}

lept_value* lept_find_object_value(lept_value* v, lept_query_callback callback, void* userdata) {
    assert(v != NULL && v->type == LEPT_OBJECT);
    for (size_t i = 0; i < v->u.o.size; i++)
        if (callback(&v->u.o.m[i].v, userdata))
            return &v->u.o.m[i].v;
    return NULL;
}
```

这里的 `lept_query_callback` 是一个回调函数类型，定义如下：

```c
typedef int (*lept_query_callback)(const lept_value* v, void* userdata);
```

回调函数接收一个 `lept_value` 指针和一个用户数据指针，返回一个整数表示是否满足条件。用户可以根据自己的需求实现回调函数，例如：

```c
int is_age_greater_than(const lept_value* v, void* userdata) {
    int target_age = *(int*)userdata;
    return v->type == LEPT_NUMBER && v->u.n > target_age;
}
```

然后使用它来查找年龄大于某个值的元素：

```c
int target_age = 25;
lept_value* found = lept_find_array_element(people_array, is_age_greater_than, &target_age);
```

## 8.5 修改功能

除了访问功能，我们还需要提供修改 JSON 数据的接口。我们已经实现了一些基本的修改函数，如 `lept_set_string`、`lept_set_array` 等。现在我们来实现一些更高级的修改函数：

```c
void lept_swap(lept_value* lhs, lept_value* rhs) {
    assert(lhs != NULL && rhs != NULL);
    if (lhs != rhs) {
        lept_value temp;
        memcpy(&temp, lhs, sizeof(lept_value));
        memcpy(lhs, rhs, sizeof(lept_value));
        memcpy(rhs, &temp, sizeof(lept_value));
    }
}

void lept_set_value_by_path(lept_value* v, const char* path, lept_value* new_v) {
    lept_value* target = lept_get_value_by_path(v, path);
    if (target != NULL)
        lept_move(target, new_v);
}

void lept_remove_value_by_path(lept_value* v, const char* path) {
    const char* p = path;
    const char* q;
    lept_value* parent = NULL;
    size_t i, len;
    
    /* 找到路径的最后一个 '.' 或 '[' */
    q = p;
    while (*p != '\0') {
        if (*p == '.' || *p == '[')
            q = p;
        p++;
    }
    
    if (q == path)  /* 根值 */
        lept_set_null(v);
    else if (*q == '.') {  /* 对象成员 */
        char* parent_path = (char*)malloc(q - path + 1);
        strncpy(parent_path, path, q - path);
        parent_path[q - path] = '\0';
        
        parent = lept_get_value_by_path(v, parent_path);
        free(parent_path);
        
        if (parent != NULL && parent->type == LEPT_OBJECT) {
            q++;  /* 跳过 '.' */
            len = strlen(q);
            
            for (i = 0; i < parent->u.o.size; i++)
                if (parent->u.o.m[i].klen == len && memcmp(parent->u.o.m[i].k, q, len) == 0)
                    break;
            
            if (i < parent->u.o.size)
                lept_remove_object_value(parent, i);
        }
    }
    else if (*q == '[') {  /* 数组元素 */
        char* parent_path = (char*)malloc(q - path + 1);
        strncpy(parent_path, path, q - path);
        parent_path[q - path] = '\0';
        
        parent = lept_get_value_by_path(v, parent_path);
        free(parent_path);
        
        if (parent != NULL && parent->type == LEPT_ARRAY) {
            q++;  /* 跳过 '[' */
            i = 0;
            while (*q >= '0' && *q <= '9')
                i = i * 10 + (*q++ - '0');
            
            if (*q == ']' && i < parent->u.a.size)
                lept_erase_array_element(parent, i, 1);
        }
    }
}
```

这些函数允许我们交换两个值，根据路径设置或删除值。

## 8.6 JSON Patch

[JSON Patch](http://jsonpatch.com/) 是一种描述对 JSON 文档的修改的格式。它由一系列操作组成，每个操作表示一种修改方式，如添加、删除、替换等。例如：

```json
[
  { "op": "add", "path": "/address/country", "value": "USA" },
  { "op": "remove", "path": "/age" },
  { "op": "replace", "path": "/name", "value": "Jane" }
]
```

这个 patch 表示添加 `address.country`，删除 `age`，替换 `name`。

我们可以实现一个函数来应用 JSON Patch：

```c
int lept_patch(lept_value* v, const lept_value* patch) {
    assert(v != NULL && patch != NULL && patch->type == LEPT_ARRAY);
    
    for (size_t i = 0; i < patch->u.a.size; i++) {
        const lept_value* op = &patch->u.a.e[i];
        if (op->type != LEPT_OBJECT)
            return LEPT_PATCH_INVALID_OPERATION;
        
        /* 获取操作类型 */
        const lept_value* op_value = lept_find_object_value((lept_value*)op, "op", 2);
        if (op_value == NULL || op_value->type != LEPT_STRING)
            return LEPT_PATCH_INVALID_OPERATION;
        
        /* 获取路径 */
        const lept_value* path_value = lept_find_object_value((lept_value*)op, "path", 4);
        if (path_value == NULL || path_value->type != LEPT_STRING)
            return LEPT_PATCH_INVALID_PATH;
        
        /* 执行操作 */
        if (memcmp(op_value->u.s.s, "add", 3) == 0 && op_value->u.s.len == 3) {
            const lept_value* value = lept_find_object_value((lept_value*)op, "value", 5);
            if (value == NULL)
                return LEPT_PATCH_MISSING_VALUE;
            
            lept_value new_v;
            lept_init(&new_v);
            lept_copy(&new_v, value);
            lept_set_value_by_path(v, path_value->u.s.s, &new_v);
            lept_free(&new_v);
        }
        else if (memcmp(op_value->u.s.s, "remove", 6) == 0 && op_value->u.s.len == 6) {
            lept_remove_value_by_path(v, path_value->u.s.s);
        }
        else if (memcmp(op_value->u.s.s, "replace", 7) == 0 && op_value->u.s.len == 7) {
            const lept_value* value = lept_find_object_value((lept_value*)op, "value", 5);
            if (value == NULL)
                return LEPT_PATCH_MISSING_VALUE;
            
            lept_value new_v;
            lept_init(&new_v);
            lept_copy(&new_v, value);
            lept_set_value_by_path(v, path_value->u.s.s, &new_v);
            lept_free(&new_v);
        }
        else if (memcmp(op_value->u.s.s, "move", 4) == 0 && op_value->u.s.len == 4) {
            const lept_value* from_value = lept_find_object_value((lept_value*)op, "from", 4);
            if (from_value == NULL || from_value->type != LEPT_STRING)
                return LEPT_PATCH_MISSING_FROM;
            
            lept_value* source = lept_get_value_by_path(v, from_value->u.s.s);
            if (source == NULL)
                return LEPT_PATCH_PATH_NOT_FOUND;
            
            lept_value temp;
            lept_init(&temp);
            lept_copy(&temp, source);
            lept_remove_value_by_path(v, from_value->u.s.s);
            lept_set_value_by_path(v, path_value->u.s.s, &temp);
            lept_free(&temp);
        }
        else if (memcmp(op_value->u.s.s, "copy", 4) == 0 && op_value->u.s.len == 4) {
            const lept_value* from_value = lept_find_object_value((lept_value*)op, "from", 4);
            if (from_value == NULL || from_value->type != LEPT_STRING)
                return LEPT_PATCH_MISSING_FROM;
            
            lept_value* source = lept_get_value_by_path(v, from_value->u.s.s);
            if (source == NULL)
                return LEPT_PATCH_PATH_NOT_FOUND;
            
            lept_value temp;
            lept_init(&temp);
            lept_copy(&temp, source);
            lept_set_value_by_path(v, path_value->u.s.s, &temp);
            lept_free(&temp);
        }
        else
            return LEPT_PATCH_UNSUPPORTED_OPERATION;
    }
    
    return LEPT_PATCH_OK;
}
```

这个函数遍历 patch 中的每个操作，根据操作类型执行相应的修改。

## 8.7 序列化和反序列化

序列化（serialization）和反序列化（deserialization）是将内存中的数据结构转换为字节流以及从字节流恢复数据结构的过程。我们已经实现了 JSON 的序列化（`lept_stringify`）和反序列化（`lept_parse`）。但在某些场景中，我们可能需要将 JSON 数据保存到文件或从文件中加载。为此，我们可以实现一些辅助函数：

```c
int lept_parse_file(lept_value* v, const char* filename) {
    FILE* fp = fopen(filename, "rb");
    if (fp == NULL)
        return LEPT_PARSE_FILE_ERROR;
    
    fseek(fp, 0, SEEK_END);
    long filesize = ftell(fp);
    fseek(fp, 0, SEEK_SET);
    
    char* buffer = (char*)malloc(filesize + 1);
    size_t readsize = fread(buffer, 1, filesize, fp);
    buffer[readsize] = '\0';
    
    int ret = lept_parse(v, buffer);
    
    fclose(fp);
    free(buffer);
    
    return ret;
}

int lept_stringify_file(const lept_value* v, const char* filename) {
    size_t length;
    char* json = lept_stringify(v, &length);
    if (json == NULL)
        return LEPT_STRINGIFY_ERROR;
    
    FILE* fp = fopen(filename, "wb");
    if (fp == NULL) {
        free(json);
        return LEPT_STRINGIFY_FILE_ERROR;
    }
    
    size_t writesize = fwrite(json, 1, length, fp);
    
    fclose(fp);
    free(json);
    
    return writesize == length ? LEPT_STRINGIFY_OK : LEPT_STRINGIFY_FILE_ERROR;
}
```

这些函数允许我们直接从文件中解析 JSON 或将 JSON 数据保存到文件中。

## 8.8 练习

1. 实现一个函数 `lept_deep_clone`，用于创建一个 `lept_value` 的深度克隆。

2. 实现一个函数 `lept_merge`，用于合并两个 JSON 对象。如果两个对象有相同的键，可以选择保留第一个对象的值、保留第二个对象的值，或者递归地合并它们。

3. 改进 `lept_get_value_by_path` 函数，使其支持更复杂的路径表达式，如带引号的键名、数组切片等。

4. 实现一个 JSON 校验器，用于检查一个 JSON 值是否符合指定的模式（schema）。

5. 实现一个函数，用于将 `lept_value` 转换为其他格式，如 XML、YAML 等。

## 8.9 下一步

在本章中，我们为我们的 JSON 库添加了许多实用的功能，使其更加完整和易用。然而，还有许多可以改进的地方，如性能优化、内存管理、错误处理等。

在下一章中，我们将探讨如何增强库的错误处理能力，提供更详细的错误信息，帮助用户更容易地调试和修复问题。 