# 第13章：JSON Patch 操作

JSON Patch 是一种用于对 JSON 文档进行部分更新的格式和语法。它允许我们以结构化的方式描述对 JSON 文档的修改，而不需要传输整个文档。本章我们将学习 JSON Patch 的规范，并实现一个 JSON Patch 处理器。

## 13.1 JSON Patch 规范简介

JSON Patch 由 [RFC 6902](https://tools.ietf.org/html/rfc6902) 定义，是一个用于描述 JSON 文档修改的 JSON 文档。一个 JSON Patch 文档是一个由操作对象组成的数组，每个操作对象代表对目标 JSON 文档执行的一个操作。

每个操作对象包含以下属性：
- `op`：操作类型，包括 `add`、`remove`、`replace`、`move`、`copy` 和 `test`
- `path`：操作的目标路径，使用 [JSON Pointer](https://tools.ietf.org/html/rfc6901) 格式
- `value`：（某些操作需要）要添加、替换或测试的值
- `from`：（某些操作需要）源路径

例如，以下 JSON Patch 文档描述了对一个对象的多个修改：

```json
[
  { "op": "add", "path": "/name", "value": "John" },
  { "op": "remove", "path": "/age" },
  { "op": "replace", "path": "/address/city", "value": "New York" },
  { "op": "move", "path": "/contact/phone", "from": "/phone" },
  { "op": "copy", "path": "/oldAddress", "from": "/address" },
  { "op": "test", "path": "/active", "value": true }
]
```

## 13.2 JSON Patch 基本操作

让我们详细了解每种操作类型：

### 13.2.1 add 操作

`add` 操作向目标路径添加新值：
- 如果路径指向对象中不存在的成员，会创建它
- 如果路径指向数组索引，在该位置插入新值
- 如果路径是 "-"，在数组末尾添加新值

```json
{ "op": "add", "path": "/biscuits/1", "value": { "name": "Ginger Nut" } }
```

### 13.2.2 remove 操作

`remove` 操作从目标路径移除值：

```json
{ "op": "remove", "path": "/biscuits/0" }
```

### 13.2.3 replace 操作

`replace` 操作用新值替换目标路径上的旧值：

```json
{ "op": "replace", "path": "/biscuits/0/name", "value": "Chocolate Digestive" }
```

### 13.2.4 move 操作

`move` 操作从一个路径移动值到另一个路径：

```json
{ "op": "move", "from": "/biscuits", "path": "/cookies" }
```

### 13.2.5 copy 操作

`copy` 操作从一个路径复制值到另一个路径：

```json
{ "op": "copy", "from": "/biscuits/0", "path": "/best_biscuit" }
```

### 13.2.6 test 操作

`test` 操作测试路径上的值是否与指定值相等，如果不相等则终止整个 Patch 操作：

```json
{ "op": "test", "path": "/best_biscuit/name", "value": "Choco Leibniz" }
```

## 13.3 实现 JSON Patch 处理器

现在我们来实现一个 JSON Patch 处理器。我们需要实现以下功能：
1. 解析 JSON Patch 文档
2. 应用 Patch 操作到目标文档
3. 处理可能出现的错误

### 13.3.1 基本结构

首先定义一些结构和类型：

```c
typedef enum {
    LEPT_PATCH_ADD,
    LEPT_PATCH_REMOVE,
    LEPT_PATCH_REPLACE,
    LEPT_PATCH_MOVE,
    LEPT_PATCH_COPY,
    LEPT_PATCH_TEST
} lept_patch_operation_type;

typedef struct {
    lept_patch_operation_type op;
    char* path;
    char* from;  /* 仅用于 move 和 copy 操作 */
    lept_value* value;  /* 仅用于 add, replace 和 test 操作 */
} lept_patch_operation;

typedef struct {
    lept_patch_operation* operations;
    size_t size;
} lept_patch;

/* 错误码 */
enum {
    LEPT_PATCH_OK = 0,
    LEPT_PATCH_INVALID_OPERATION,
    LEPT_PATCH_INVALID_PATH,
    LEPT_PATCH_PATH_NOT_FOUND,
    LEPT_PATCH_TEST_FAILED,
    /* ... 其他错误码 ... */
};
```

### 13.3.2 解析 JSON Patch 文档

然后我们需要解析 JSON Patch 文档，将其转换为 `lept_patch` 结构：

```c
int lept_parse_patch(lept_patch* patch, const lept_value* json) {
    size_t i;
    assert(patch != NULL);
    assert(json != NULL && json->type == LEPT_ARRAY);
    
    /* 分配内存 */
    patch->size = json->u.a.size;
    patch->operations = (lept_patch_operation*)malloc(patch->size * sizeof(lept_patch_operation));
    if (patch->operations == NULL)
        return LEPT_PATCH_NO_MEMORY;
    
    /* 初始化 */
    memset(patch->operations, 0, patch->size * sizeof(lept_patch_operation));
    
    /* 解析每个操作 */
    for (i = 0; i < patch->size; i++) {
        lept_patch_operation* op = &patch->operations[i];
        const lept_value* op_obj = &json->u.a.e[i];
        
        /* 确保操作是一个对象 */
        if (op_obj->type != LEPT_OBJECT)
            return LEPT_PATCH_INVALID_OPERATION;
        
        /* 解析操作类型 */
        const lept_value* op_type = lept_find_object_value(op_obj, "op");
        if (op_type == NULL || op_type->type != LEPT_STRING)
            return LEPT_PATCH_INVALID_OPERATION;
        
        /* 设置操作类型 */
        if (strcmp(op_type->u.s.s, "add") == 0)
            op->op = LEPT_PATCH_ADD;
        else if (strcmp(op_type->u.s.s, "remove") == 0)
            op->op = LEPT_PATCH_REMOVE;
        else if (strcmp(op_type->u.s.s, "replace") == 0)
            op->op = LEPT_PATCH_REPLACE;
        else if (strcmp(op_type->u.s.s, "move") == 0)
            op->op = LEPT_PATCH_MOVE;
        else if (strcmp(op_type->u.s.s, "copy") == 0)
            op->op = LEPT_PATCH_COPY;
        else if (strcmp(op_type->u.s.s, "test") == 0)
            op->op = LEPT_PATCH_TEST;
        else
            return LEPT_PATCH_INVALID_OPERATION;
        
        /* 解析路径 */
        const lept_value* path = lept_find_object_value(op_obj, "path");
        if (path == NULL || path->type != LEPT_STRING)
            return LEPT_PATCH_INVALID_PATH;
        op->path = strdup(path->u.s.s);
        
        /* 根据操作类型解析其他属性 */
        switch (op->op) {
            case LEPT_PATCH_ADD:
            case LEPT_PATCH_REPLACE:
            case LEPT_PATCH_TEST: {
                const lept_value* value = lept_find_object_value(op_obj, "value");
                if (value == NULL)
                    return LEPT_PATCH_INVALID_OPERATION;
                op->value = (lept_value*)malloc(sizeof(lept_value));
                lept_init(op->value);
                lept_copy(op->value, value);
                break;
            }
            case LEPT_PATCH_MOVE:
            case LEPT_PATCH_COPY: {
                const lept_value* from = lept_find_object_value(op_obj, "from");
                if (from == NULL || from->type != LEPT_STRING)
                    return LEPT_PATCH_INVALID_OPERATION;
                op->from = strdup(from->u.s.s);
                break;
            }
            default:
                break;
        }
    }
    
    return LEPT_PATCH_OK;
}
```

### 13.3.3 应用 Patch 操作

接下来，我们实现函数来应用 Patch 操作到目标文档：

```c
int lept_apply_patch(lept_value* target, const lept_patch* patch) {
    size_t i;
    int ret;
    assert(target != NULL);
    assert(patch != NULL);
    
    /* 应用每个操作 */
    for (i = 0; i < patch->size; i++) {
        const lept_patch_operation* op = &patch->operations[i];
        
        switch (op->op) {
            case LEPT_PATCH_ADD:
                ret = lept_patch_add(target, op->path, op->value);
                break;
            case LEPT_PATCH_REMOVE:
                ret = lept_patch_remove(target, op->path);
                break;
            case LEPT_PATCH_REPLACE:
                ret = lept_patch_replace(target, op->path, op->value);
                break;
            case LEPT_PATCH_MOVE:
                ret = lept_patch_move(target, op->path, op->from);
                break;
            case LEPT_PATCH_COPY:
                ret = lept_patch_copy(target, op->path, op->from);
                break;
            case LEPT_PATCH_TEST:
                ret = lept_patch_test(target, op->path, op->value);
                break;
            default:
                ret = LEPT_PATCH_INVALID_OPERATION;
                break;
        }
        
        if (ret != LEPT_PATCH_OK)
            return ret;
    }
    
    return LEPT_PATCH_OK;
}
```

现在，我们需要实现每种操作的处理函数。

### 13.3.4 实现各种操作

以下是每种操作的实现示例：

```c
/* 添加操作 */
int lept_patch_add(lept_value* target, const char* path, const lept_value* value) {
    /* 解析 JSON 指针 */
    lept_json_pointer pointer;
    int ret = lept_parse_json_pointer(&pointer, path);
    if (ret != LEPT_POINTER_OK)
        return LEPT_PATCH_INVALID_PATH;
    
    /* 找到父节点 */
    lept_value* parent;
    char* last_token;
    ret = lept_json_pointer_resolve_parent(target, &pointer, &parent, &last_token);
    
    if (ret != LEPT_POINTER_OK) {
        lept_free_json_pointer(&pointer);
        return LEPT_PATCH_PATH_NOT_FOUND;
    }
    
    /* 根据父节点类型处理 */
    if (parent->type == LEPT_OBJECT) {
        /* 添加/替换对象成员 */
        lept_value v;
        lept_init(&v);
        lept_copy(&v, value);
        lept_set_object_value(parent, last_token, &v);
    } 
    else if (parent->type == LEPT_ARRAY) {
        /* 向数组添加元素 */
        size_t index;
        
        if (strcmp(last_token, "-") == 0) {
            /* 添加到数组末尾 */
            index = parent->u.a.size;
        } 
        else {
            /* 添加到指定位置 */
            char* end;
            index = strtol(last_token, &end, 10);
            if (*end != '\0' || index > parent->u.a.size) {
                lept_free_json_pointer(&pointer);
                return LEPT_PATCH_INVALID_PATH;
            }
        }
        
        /* 插入元素 */
        lept_reserve_array(parent, parent->u.a.size + 1);
        if (index < parent->u.a.size) {
            /* 移动后面的元素 */
            memmove(&parent->u.a.e[index + 1], &parent->u.a.e[index], 
                   (parent->u.a.size - index) * sizeof(lept_value));
        }
        
        lept_init(&parent->u.a.e[index]);
        lept_copy(&parent->u.a.e[index], value);
        parent->u.a.size++;
    } 
    else {
        lept_free_json_pointer(&pointer);
        return LEPT_PATCH_INVALID_PATH;
    }
    
    lept_free_json_pointer(&pointer);
    return LEPT_PATCH_OK;
}

/* 其他操作的实现类似... */
```

以上只是 `add` 操作的实现示例，其他操作如 `remove`、`replace`、`move`、`copy` 和 `test` 的实现逻辑类似，但各有不同。

### 13.3.5 资源清理

最后，我们需要实现资源清理函数：

```c
void lept_free_patch(lept_patch* patch) {
    size_t i;
    assert(patch != NULL);
    
    for (i = 0; i < patch->size; i++) {
        lept_patch_operation* op = &patch->operations[i];
        
        /* 释放路径和源路径 */
        if (op->path) free(op->path);
        if (op->from) free(op->from);
        
        /* 释放值 */
        if (op->value) {
            lept_free(op->value);
            free(op->value);
        }
    }
    
    free(patch->operations);
    patch->operations = NULL;
    patch->size = 0;
}
```

## 13.4 使用 JSON Patch

以下是一个使用我们实现的 JSON Patch 功能的示例：

```c
void example_json_patch() {
    /* 原始文档 */
    lept_value doc;
    lept_parse(&doc, "{\"foo\":\"bar\",\"baz\":\"qux\"}");
    
    /* Patch 文档 */
    lept_value patch_doc;
    lept_parse(&patch_doc, "[{\"op\":\"replace\",\"path\":\"/foo\",\"value\":\"boo\"},{\"op\":\"add\",\"path\":\"/hello\",\"value\":\"world\"},{\"op\":\"remove\",\"path\":\"/baz\"}]");
    
    /* 创建 Patch */
    lept_patch patch;
    int ret = lept_parse_patch(&patch, &patch_doc);
    if (ret != LEPT_PATCH_OK) {
        printf("Failed to parse patch: %d\n", ret);
        return;
    }
    
    /* 应用 Patch */
    ret = lept_apply_patch(&doc, &patch);
    if (ret != LEPT_PATCH_OK) {
        printf("Failed to apply patch: %d\n", ret);
        lept_free_patch(&patch);
        return;
    }
    
    /* 输出结果 */
    char* json = lept_stringify(&doc, NULL);
    printf("Patched document: %s\n", json);  /* 输出: {"foo":"boo","hello":"world"} */
    
    /* 清理资源 */
    free(json);
    lept_free_patch(&patch);
    lept_free(&doc);
    lept_free(&patch_doc);
}
```

## 13.5 批量处理 JSON 文档

JSON Patch 特别适合用于批量处理服务器上的 JSON 文档，以及客户端和服务器之间的数据同步。

### 13.5.1 HTTP 与 JSON Patch

JSON Patch 已被定义为 HTTP PATCH 方法的标准媒体类型 `application/json-patch+json`。当使用 HTTP PATCH 请求时，客户端可以发送一个 JSON Patch 文档来修改服务器上的资源，而不必发送整个资源。

以下是一个使用 HTTP PATCH 的示例：

```
PATCH /my-document HTTP/1.1
Host: example.org
Content-Type: application/json-patch+json

[
  { "op": "replace", "path": "/title", "value": "New Title" },
  { "op": "add", "path": "/author", "value": "John Doe" }
]
```

### 13.5.2 处理冲突

在并发环境中，可能会出现多个客户端尝试修改同一个文档的情况，这可能导致冲突。我们可以使用 `test` 操作来检测这些冲突：

```json
[
  { "op": "test", "path": "/version", "value": 1 },
  { "op": "replace", "path": "/title", "value": "New Title" },
  { "op": "replace", "path": "/version", "value": 2 }
]
```

如果 `test` 操作失败（即文档的版本号不是1），那么整个 Patch 操作将失败，从而避免了覆盖其他客户端的更改。

## 13.6 练习

1. 完整实现所有 JSON Patch 操作（`add`、`remove`、`replace`、`move`、`copy` 和 `test`）。

2. 添加对复杂 JSON 指针的支持，包括转义字符（例如 `/path/with~1slash/and~0tilde`）。

3. 实现 JSON Merge Patch（RFC 7396）作为 JSON Patch 的替代方案。

4. 实现一个功能，用于生成两个 JSON 文档之间的 JSON Patch 文档。

5. 为 JSON Patch 实现添加详细的错误报告，包括操作索引和具体失败原因。

## 13.7 下一步

本章我们学习了 JSON Patch 操作，它是对 JSON 文档进行部分更新的强大工具。在下一章中，我们将探讨另一种 JSON 文档更新方式：JSON Merge Patch，这是一种更简单但功能有限的替代方案。