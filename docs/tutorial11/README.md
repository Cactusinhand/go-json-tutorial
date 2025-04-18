# 第11章：JSON Schema

在前面的章节中，我们已经实现了一个功能完整的 JSON 库，包括解析、生成、访问、修改等功能。然而，在实际应用中，我们常常需要验证 JSON 数据是否符合特定的结构和约束。本章将介绍 JSON Schema，这是一种用于描述 JSON 数据结构和验证 JSON 数据的标准。

## 11.1 JSON Schema 简介

[JSON Schema](https://json-schema.org/) 是一种用于定义 JSON 数据格式的声明式语言。它本身也是用 JSON 格式表示的，用于描述 JSON 数据的结构、类型、约束等。JSON Schema 可以用于：

1. **数据验证**：检查 JSON 数据是否符合预期的格式和约束。
2. **文档生成**：根据 Schema 自动生成 API 文档。
3. **代码生成**：根据 Schema 自动生成数据模型代码。
4. **客户端验证**：在客户端验证用户输入，减少服务器的负担。

JSON Schema 有自己的 MIME 类型：`application/schema+json`。

## 11.2 JSON Schema 的基本结构

一个简单的 JSON Schema 示例：

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "description": "A person's information",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "The person's name"
    },
    "age": {
      "type": "integer",
      "minimum": 0,
      "description": "The person's age"
    },
    "email": {
      "type": "string",
      "format": "email",
      "description": "The person's email address"
    }
  },
  "required": ["name", "age"]
}
```

上面的 Schema 描述了一个"人"的 JSON 对象，它应该包含名字、年龄和可选的电子邮件。其中，名字必须是字符串，年龄必须是非负整数，电子邮件（如果提供）必须是有效的电子邮件格式。

## 11.3 实现 JSON Schema 验证器

在我们的 JSON 库中，我们将实现一个 JSON Schema 验证器，用于验证 JSON 数据是否符合指定的 Schema。首先，我们定义一个结构体，表示验证结果：

```c
typedef struct {
    int valid;            /* 是否有效 */
    const char* message;  /* 错误消息 */
    const char* path;     /* 错误路径 */
} lept_schema_result;
```

然后，我们定义验证函数的接口：

```c
int lept_validate_schema(const lept_value* schema, const lept_value* instance, lept_schema_result* result);
```

这个函数接收一个 Schema（`schema`）、一个实例（`instance`）和一个用于存储结果的指针（`result`）。它返回一个整数，表示验证是否成功。

## 11.4 验证基本类型

首先，我们实现对基本类型的验证。JSON Schema 使用 `type` 关键字来指定值的类型，它可以是字符串或字符串数组。例如：

```json
{ "type": "string" }
{ "type": ["string", "number"] }
```

我们实现一个函数，用于验证值的类型：

```c
static int validate_type(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    const lept_value* type = lept_find_object_value((lept_value*)schema, "type", 4);
    if (type == NULL)
        return 1;  /* 如果没有指定类型，默认为有效 */
    
    if (type->type == LEPT_STRING) {
        /* 单一类型 */
        const char* type_str = type->u.s.s;
        size_t type_len = type->u.s.len;
        
        if (strncmp(type_str, "null", type_len) == 0 && instance->type == LEPT_NULL)
            return 1;
        else if (strncmp(type_str, "boolean", type_len) == 0 && (instance->type == LEPT_TRUE || instance->type == LEPT_FALSE))
            return 1;
        else if (strncmp(type_str, "integer", type_len) == 0 && instance->type == LEPT_NUMBER && (double)(long)instance->u.n == instance->u.n)
            return 1;
        else if (strncmp(type_str, "number", type_len) == 0 && instance->type == LEPT_NUMBER)
            return 1;
        else if (strncmp(type_str, "string", type_len) == 0 && instance->type == LEPT_STRING)
            return 1;
        else if (strncmp(type_str, "array", type_len) == 0 && instance->type == LEPT_ARRAY)
            return 1;
        else if (strncmp(type_str, "object", type_len) == 0 && instance->type == LEPT_OBJECT)
            return 1;
    }
    else if (type->type == LEPT_ARRAY) {
        /* 多个类型中的一个 */
        for (size_t i = 0; i < type->u.a.size; i++) {
            const lept_value* t = &type->u.a.e[i];
            if (t->type == LEPT_STRING) {
                const char* type_str = t->u.s.s;
                size_t type_len = t->u.s.len;
                
                if (strncmp(type_str, "null", type_len) == 0 && instance->type == LEPT_NULL)
                    return 1;
                else if (strncmp(type_str, "boolean", type_len) == 0 && (instance->type == LEPT_TRUE || instance->type == LEPT_FALSE))
                    return 1;
                else if (strncmp(type_str, "integer", type_len) == 0 && instance->type == LEPT_NUMBER && (double)(long)instance->u.n == instance->u.n)
                    return 1;
                else if (strncmp(type_str, "number", type_len) == 0 && instance->type == LEPT_NUMBER)
                    return 1;
                else if (strncmp(type_str, "string", type_len) == 0 && instance->type == LEPT_STRING)
                    return 1;
                else if (strncmp(type_str, "array", type_len) == 0 && instance->type == LEPT_ARRAY)
                    return 1;
                else if (strncmp(type_str, "object", type_len) == 0 && instance->type == LEPT_OBJECT)
                    return 1;
            }
        }
    }
    
    result->valid = 0;
    result->message = "Type mismatch";
    return 0;
}
```

## 11.5 验证数值约束

JSON Schema 提供了多种关键字来约束数值，例如 `minimum`、`maximum`、`exclusiveMinimum`、`exclusiveMaximum`、`multipleOf` 等。

```c
static int validate_number(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    if (instance->type != LEPT_NUMBER)
        return 1;  /* 非数值类型不验证数值约束 */
    
    double value = instance->u.n;
    const lept_value* minimum = lept_find_object_value((lept_value*)schema, "minimum", 7);
    const lept_value* maximum = lept_find_object_value((lept_value*)schema, "maximum", 7);
    const lept_value* exclusive_minimum = lept_find_object_value((lept_value*)schema, "exclusiveMinimum", 16);
    const lept_value* exclusive_maximum = lept_find_object_value((lept_value*)schema, "exclusiveMaximum", 16);
    const lept_value* multiple_of = lept_find_object_value((lept_value*)schema, "multipleOf", 10);
    
    /* 检查最小值 */
    if (minimum != NULL && minimum->type == LEPT_NUMBER) {
        if (value < minimum->u.n) {
            result->valid = 0;
            result->message = "Value less than minimum";
            return 0;
        }
    }
    
    /* 检查最大值 */
    if (maximum != NULL && maximum->type == LEPT_NUMBER) {
        if (value > maximum->u.n) {
            result->valid = 0;
            result->message = "Value greater than maximum";
            return 0;
        }
    }
    
    /* 检查排他性最小值 */
    if (exclusive_minimum != NULL && exclusive_minimum->type == LEPT_NUMBER) {
        if (value <= exclusive_minimum->u.n) {
            result->valid = 0;
            result->message = "Value less than or equal to exclusiveMinimum";
            return 0;
        }
    }
    
    /* 检查排他性最大值 */
    if (exclusive_maximum != NULL && exclusive_maximum->type == LEPT_NUMBER) {
        if (value >= exclusive_maximum->u.n) {
            result->valid = 0;
            result->message = "Value greater than or equal to exclusiveMaximum";
            return 0;
        }
    }
    
    /* 检查倍数 */
    if (multiple_of != NULL && multiple_of->type == LEPT_NUMBER) {
        double n = multiple_of->u.n;
        if (n <= 0) {
            result->valid = 0;
            result->message = "multipleOf must be greater than 0";
            return 0;
        }
        double remainder = fmod(value, n);
        if (remainder > 0.0 && remainder < n) {
            result->valid = 0;
            result->message = "Value is not a multiple of multipleOf";
            return 0;
        }
    }
    
    return 1;
}
```

## 11.6 验证字符串约束

JSON Schema 提供了多种关键字来约束字符串，例如 `minLength`、`maxLength`、`pattern`、`format` 等。

```c
static int validate_string(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    if (instance->type != LEPT_STRING)
        return 1;  /* 非字符串类型不验证字符串约束 */
    
    size_t length = instance->u.s.len;
    const lept_value* min_length = lept_find_object_value((lept_value*)schema, "minLength", 9);
    const lept_value* max_length = lept_find_object_value((lept_value*)schema, "maxLength", 9);
    const lept_value* pattern = lept_find_object_value((lept_value*)schema, "pattern", 7);
    
    /* 检查最小长度 */
    if (min_length != NULL && min_length->type == LEPT_NUMBER) {
        if (length < (size_t)min_length->u.n) {
            result->valid = 0;
            result->message = "String length less than minLength";
            return 0;
        }
    }
    
    /* 检查最大长度 */
    if (max_length != NULL && max_length->type == LEPT_NUMBER) {
        if (length > (size_t)max_length->u.n) {
            result->valid = 0;
            result->message = "String length greater than maxLength";
            return 0;
        }
    }
    
    /* 检查模式 */
    if (pattern != NULL && pattern->type == LEPT_STRING) {
        /* 这里需要使用正则表达式库来验证模式，例如 PCRE */
        /* 由于正则表达式库的整合超出了本教程的范围，这里简化处理 */
        if (pattern->u.s.len > 0) {
            result->valid = 0;
            result->message = "Pattern validation not implemented";
            return 0;
        }
    }
    
    return 1;
}
```

## 11.7 验证数组约束

JSON Schema 提供了多种关键字来约束数组，例如 `items`、`minItems`、`maxItems`、`uniqueItems` 等。

```c
static int validate_array(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    if (instance->type != LEPT_ARRAY)
        return 1;  /* 非数组类型不验证数组约束 */
    
    size_t size = instance->u.a.size;
    const lept_value* min_items = lept_find_object_value((lept_value*)schema, "minItems", 8);
    const lept_value* max_items = lept_find_object_value((lept_value*)schema, "maxItems", 8);
    const lept_value* unique_items = lept_find_object_value((lept_value*)schema, "uniqueItems", 11);
    const lept_value* items = lept_find_object_value((lept_value*)schema, "items", 5);
    
    /* 检查最小项数 */
    if (min_items != NULL && min_items->type == LEPT_NUMBER) {
        if (size < (size_t)min_items->u.n) {
            result->valid = 0;
            result->message = "Array size less than minItems";
            return 0;
        }
    }
    
    /* 检查最大项数 */
    if (max_items != NULL && max_items->type == LEPT_NUMBER) {
        if (size > (size_t)max_items->u.n) {
            result->valid = 0;
            result->message = "Array size greater than maxItems";
            return 0;
        }
    }
    
    /* 检查唯一性 */
    if (unique_items != NULL && (unique_items->type == LEPT_TRUE || (unique_items->type == LEPT_NUMBER && unique_items->u.n != 0))) {
        for (size_t i = 0; i < size; i++) {
            for (size_t j = i + 1; j < size; j++) {
                if (lept_is_equal(&instance->u.a.e[i], &instance->u.a.e[j])) {
                    result->valid = 0;
                    result->message = "Array items are not unique";
                    return 0;
                }
            }
        }
    }
    
    /* 验证数组元素 */
    if (items != NULL) {
        if (items->type == LEPT_OBJECT) {
            /* 每个元素都使用相同的 Schema */
            for (size_t i = 0; i < size; i++) {
                char path_buffer[32];
                sprintf(path_buffer, "[%zu]", i);
                result->path = path_buffer;
                
                if (!lept_validate_schema(items, &instance->u.a.e[i], result))
                    return 0;
            }
        }
        else if (items->type == LEPT_ARRAY) {
            /* 每个元素使用对应的 Schema */
            size_t schema_size = items->u.a.size;
            for (size_t i = 0; i < size && i < schema_size; i++) {
                char path_buffer[32];
                sprintf(path_buffer, "[%zu]", i);
                result->path = path_buffer;
                
                if (!lept_validate_schema(&items->u.a.e[i], &instance->u.a.e[i], result))
                    return 0;
            }
        }
    }
    
    return 1;
}
```

## 11.8 验证对象约束

JSON Schema 提供了多种关键字来约束对象，例如 `properties`、`required`、`minProperties`、`maxProperties`、`dependencies` 等。

```c
static int validate_object(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    if (instance->type != LEPT_OBJECT)
        return 1;  /* 非对象类型不验证对象约束 */
    
    size_t size = instance->u.o.size;
    const lept_value* min_properties = lept_find_object_value((lept_value*)schema, "minProperties", 13);
    const lept_value* max_properties = lept_find_object_value((lept_value*)schema, "maxProperties", 13);
    const lept_value* required = lept_find_object_value((lept_value*)schema, "required", 8);
    const lept_value* properties = lept_find_object_value((lept_value*)schema, "properties", 10);
    
    /* 检查最小属性数 */
    if (min_properties != NULL && min_properties->type == LEPT_NUMBER) {
        if (size < (size_t)min_properties->u.n) {
            result->valid = 0;
            result->message = "Object has fewer properties than minProperties";
            return 0;
        }
    }
    
    /* 检查最大属性数 */
    if (max_properties != NULL && max_properties->type == LEPT_NUMBER) {
        if (size > (size_t)max_properties->u.n) {
            result->valid = 0;
            result->message = "Object has more properties than maxProperties";
            return 0;
        }
    }
    
    /* 检查必需属性 */
    if (required != NULL && required->type == LEPT_ARRAY) {
        for (size_t i = 0; i < required->u.a.size; i++) {
            const lept_value* req = &required->u.a.e[i];
            if (req->type == LEPT_STRING) {
                size_t index = lept_find_object_index(instance, req->u.s.s, req->u.s.len);
                if (index == LEPT_KEY_NOT_FOUND) {
                    result->valid = 0;
                    result->message = "Required property missing";
                    result->path = req->u.s.s;
                    return 0;
                }
            }
        }
    }
    
    /* 验证属性 */
    if (properties != NULL && properties->type == LEPT_OBJECT) {
        for (size_t i = 0; i < instance->u.o.size; i++) {
            const char* key = instance->u.o.m[i].k;
            size_t key_len = instance->u.o.m[i].klen;
            const lept_value* value = &instance->u.o.m[i].v;
            
            size_t schema_index = lept_find_object_index(properties, key, key_len);
            if (schema_index != LEPT_KEY_NOT_FOUND) {
                const lept_value* property_schema = &properties->u.o.m[schema_index].v;
                result->path = key;
                
                if (!lept_validate_schema(property_schema, value, result))
                    return 0;
            }
        }
    }
    
    return 1;
}
```

## 11.9 验证条件约束

JSON Schema 还提供了一些条件约束关键字，如 `allOf`、`anyOf`、`oneOf`、`not` 等，用于组合多个 Schema 或对 Schema 进行逻辑操作。

```c
static int validate_logic(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    const lept_value* all_of = lept_find_object_value((lept_value*)schema, "allOf", 5);
    const lept_value* any_of = lept_find_object_value((lept_value*)schema, "anyOf", 5);
    const lept_value* one_of = lept_find_object_value((lept_value*)schema, "oneOf", 5);
    const lept_value* not_schema = lept_find_object_value((lept_value*)schema, "not", 3);
    
    /* allOf：所有 Schema 都必须匹配 */
    if (all_of != NULL && all_of->type == LEPT_ARRAY) {
        for (size_t i = 0; i < all_of->u.a.size; i++) {
            if (!lept_validate_schema(&all_of->u.a.e[i], instance, result)) {
                return 0;
            }
        }
    }
    
    /* anyOf：至少一个 Schema 必须匹配 */
    if (any_of != NULL && any_of->type == LEPT_ARRAY) {
        int valid = 0;
        for (size_t i = 0; i < any_of->u.a.size; i++) {
            lept_schema_result temp_result = { 1, NULL, NULL };
            if (lept_validate_schema(&any_of->u.a.e[i], instance, &temp_result)) {
                valid = 1;
                break;
            }
        }
        if (!valid) {
            result->valid = 0;
            result->message = "Instance does not match any schema in anyOf";
            return 0;
        }
    }
    
    /* oneOf：恰好一个 Schema 必须匹配 */
    if (one_of != NULL && one_of->type == LEPT_ARRAY) {
        int valid_count = 0;
        for (size_t i = 0; i < one_of->u.a.size; i++) {
            lept_schema_result temp_result = { 1, NULL, NULL };
            if (lept_validate_schema(&one_of->u.a.e[i], instance, &temp_result)) {
                valid_count++;
            }
        }
        if (valid_count != 1) {
            result->valid = 0;
            result->message = valid_count == 0 
                ? "Instance does not match any schema in oneOf" 
                : "Instance matches more than one schema in oneOf";
            return 0;
        }
    }
    
    /* not：Schema 必须不匹配 */
    if (not_schema != NULL) {
        lept_schema_result temp_result = { 1, NULL, NULL };
        if (lept_validate_schema(not_schema, instance, &temp_result)) {
            result->valid = 0;
            result->message = "Instance matches schema in not";
            return 0;
        }
    }
    
    return 1;
}
```

## 11.10 整合验证函数

最后，我们整合所有的验证函数，实现 `lept_validate_schema`：

```c
int lept_validate_schema(const lept_value* schema, const lept_value* instance, lept_schema_result* result) {
    assert(schema != NULL && instance != NULL && result != NULL);
    
    result->valid = 1;
    result->message = NULL;
    result->path = NULL;
    
    if (!validate_type(schema, instance, result))
        return 0;
    
    if (!validate_number(schema, instance, result))
        return 0;
    
    if (!validate_string(schema, instance, result))
        return 0;
    
    if (!validate_array(schema, instance, result))
        return 0;
    
    if (!validate_object(schema, instance, result))
        return 0;
    
    if (!validate_logic(schema, instance, result))
        return 0;
    
    return 1;
}
```

## 11.11 使用 JSON Schema 验证器

以下是一个使用 JSON Schema 验证器的例子：

```c
void test_schema_validation() {
    lept_value schema, instance;
    lept_schema_result result;
    
    /* 解析 Schema */
    lept_init(&schema);
    lept_parse(&schema, "{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"age\":{\"type\":\"integer\",\"minimum\":0}},\"required\":[\"name\",\"age\"]}");
    
    /* 解析实例 */
    lept_init(&instance);
    lept_parse(&instance, "{\"name\":\"John\",\"age\":30}");
    
    /* 验证实例 */
    if (lept_validate_schema(&schema, &instance, &result)) {
        printf("Validation passed.\n");
    }
    else {
        printf("Validation failed: %s at %s\n", result.message, result.path ? result.path : "root");
    }
    
    /* 试验一个无效实例 */
    lept_free(&instance);
    lept_init(&instance);
    lept_parse(&instance, "{\"name\":\"John\",\"age\":-1}");
    
    if (lept_validate_schema(&schema, &instance, &result)) {
        printf("Validation passed.\n");
    }
    else {
        printf("Validation failed: %s at %s\n", result.message, result.path ? result.path : "root");
    }
    
    lept_free(&schema);
    lept_free(&instance);
}
```

## 11.12 练习

1. 实现一个函数 `lept_schema_generate_error`，用于生成详细的错误消息，包括 Schema 的路径、实例的路径、错误类型等。

2. 增强 `validate_string` 函数，支持 `format` 关键字，验证字符串是否符合特定的格式，如 email、uri、date-time 等。

3. 实现对 `additionalProperties` 和 `patternProperties` 的支持，控制对象中未在 `properties` 中定义的属性。

4. 实现对 `enum` 关键字的支持，验证值是否在枚举列表中。

5. 实现对 `const` 关键字的支持，验证值是否等于指定的常量。

## 11.13 下一步

在本章中，我们实现了一个简单的 JSON Schema 验证器，可以用于验证 JSON 数据是否符合指定的 Schema。JSON Schema 是一个强大的工具，可以帮助我们保证数据的质量和一致性。在实际应用中、我们可以进一步扩展它，支持更多的关键字和功能。

在下一章中，我们将探讨 JSON 库的高级功能，如序列化性能优化、内存优化等，使我们的 JSON 库更加高效和实用。 