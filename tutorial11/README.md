# 从零开始的 JSON 库教程（十一）：JSON Schema 验证

## JSON Schema 简介

JSON Schema 是一种基于 JSON 格式定义的规范，用于验证、描述和标注 JSON 文档。它提供了一种描述 JSON 数据结构的标准方法，可以用来：

- 验证数据符合预期的格式和内容
- 提供清晰、人类可读的文档
- 自动生成表单或配置界面
- 支持数据自动完成功能
- 验证服务器和客户端之间的数据交换

本章实现了 JSON Schema Draft 7 规范的核心部分，作为我们 JSON 库的扩展功能。

## 实现概述

我们的 JSON Schema 验证器支持以下功能：

1. **基本类型验证**：验证 JSON 数据的类型（字符串、数字、整数、布尔值、空值、对象和数组）
2. **数值约束**：支持最小值、最大值、独占范围和倍数验证
3. **字符串约束**：支持最小/最大长度、正则表达式模式和常见格式（如电子邮件和日期时间）
4. **数组约束**：支持最小/最大元素数量、元素唯一性和元素类型验证
5. **对象约束**：支持最小/最大属性数、必需属性、属性模式和附加属性控制
6. **逻辑组合**：支持 allOf（所有模式都匹配）、anyOf（至少一个模式匹配）、oneOf（恰好一个模式匹配）和 not（模式不匹配）

## 关键结构

### SchemaValidationError

表示验证过程中发现的错误，包含错误路径和消息：

```go
type SchemaValidationError struct {
    Path    string // 导致错误的 JSON 路径
    Message string // 错误描述
}
```

### SchemaValidationResult

保存验证的最终结果，包括验证是否通过和错误列表：

```go
type SchemaValidationResult struct {
    Valid  bool                    // 是否验证通过
    Errors []SchemaValidationError // 验证错误列表
}
```

### JSONSchema

表示一个 JSON Schema 对象，提供验证方法：

```go
type JSONSchema struct {
    Schema *Value // 存储 JSON Schema 的 Value 对象
}
```

## 使用示例

创建一个 JSON Schema 并验证数据：

```go
package main

import (
    "fmt"
    "github.com/Cactusinhand/go-json-tutorial/tutorial11"
)

func main() {
    // 定义 Schema
    schemaJSON := `{
        "type": "object",
        "required": ["name", "age"],
        "properties": {
            "name": {
                "type": "string",
                "minLength": 2
            },
            "age": {
                "type": "integer",
                "minimum": 18
            },
            "email": {
                "type": "string",
                "format": "email"
            }
        }
    }`
    
    // 解析 Schema
    schema, err := leptjson.NewJSONSchema(schemaJSON)
    if err != nil {
        fmt.Printf("Schema 创建失败: %v\n", err)
        return
    }
    
    // 要验证的有效数据
    validJSON := `{
        "name": "张三",
        "age": 25,
        "email": "zhangsan@example.com"
    }`
    validValue := &leptjson.Value{}
    if err := leptjson.Parse(validValue, validJSON); err != leptjson.PARSE_OK {
        fmt.Printf("数据解析失败: %v\n", err)
        return
    }
    
    // 验证有效数据
    result := schema.Validate(validValue)
    if result.Valid {
        fmt.Println("数据验证通过！")
    } else {
        fmt.Println("数据验证失败:")
        for _, err := range result.Errors {
            fmt.Printf("- %s\n", err.Error())
        }
    }
    
    // 要验证的无效数据
    invalidJSON := `{
        "name": "李",
        "age": 16
    }`
    invalidValue := &leptjson.Value{}
    if err := leptjson.Parse(invalidValue, invalidJSON); err != leptjson.PARSE_OK {
        fmt.Printf("数据解析失败: %v\n", err)
        return
    }
    
    // 验证无效数据
    result = schema.Validate(invalidValue)
    if !result.Valid {
        fmt.Println("\n无效数据验证（预期失败）:")
        for _, err := range result.Errors {
            fmt.Printf("- %s\n", err.Error())
        }
    }
}
```

输出结果：

```
数据验证通过！

无效数据验证（预期失败）:
- 位于 'name': 字符串长度 1 小于最小长度 2
- 位于 'age': 值 16 小于最小值 18
```

## 支持的 JSON Schema 关键字

### 通用关键字

- **type**: 指定值的类型，可以是单一类型或类型数组
- **enum**: 限制值为指定的枚举列表中的一个
- **const**: 要求值必须等于指定的常量

### 数值关键字

- **minimum**: 指定最小值（包含）
- **exclusiveMinimum**: 指定独占最小值（不包含）
- **maximum**: 指定最大值（包含）
- **exclusiveMaximum**: 指定独占最大值（不包含）
- **multipleOf**: 要求值是指定数的倍数

### 字符串关键字

- **minLength**: 指定最小字符串长度
- **maxLength**: 指定最大字符串长度
- **pattern**: 指定字符串必须匹配的正则表达式
- **format**: 指定字符串格式（如电子邮件、日期时间等）

### 数组关键字

- **minItems**: 指定最小数组长度
- **maxItems**: 指定最大数组长度
- **uniqueItems**: 要求数组元素必须唯一
- **items**: 指定数组元素的模式（对象或数组）
- **contains**: 要求数组至少包含一个匹配指定模式的元素

### 对象关键字

- **minProperties**: 指定最小属性数量
- **maxProperties**: 指定最大属性数量
- **required**: 指定必需的属性名数组
- **properties**: 定义对象的属性及其模式
- **patternProperties**: 使用正则表达式匹配属性名并验证其值
- **additionalProperties**: 控制未在 properties 中定义的属性
- **propertyNames**: 验证所有属性名
- **dependencies**: 指定当某个属性存在时，其他属性也必须存在

### 逻辑关键字

- **allOf**: 要求值同时满足所有指定的模式
- **anyOf**: 要求值至少满足一个指定的模式
- **oneOf**: 要求值恰好满足一个指定的模式
- **not**: 要求值不满足指定的模式

## 限制与未来改进

当前实现有以下限制：

1. 不支持 Schema 引用（$ref）和模式重用
2. 不支持基于 URI 的模式标识符（$id）
3. 不支持默认值（default）关键字的处理
4. 条件验证（if/then/else）尚未实现
5. 某些格式验证的实现是简化的

未来改进方向：

1. 支持 Schema 引用和 $ref 解析
2. 实现条件验证（if/then/else）
3. 支持注释关键字（title, description, examples等）
4. 增加更多格式验证（date, time, hostname等）
5. 支持自定义格式验证器
6. 完善错误信息，提供更友好的错误提示

## 参考资料

- [JSON Schema 官方网站](https://json-schema.org/)
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/)
- [JSON Schema Draft 7 规范](https://datatracker.ietf.org/doc/html/draft-handrews-json-schema-validation-01)
