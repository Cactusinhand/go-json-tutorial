# 从零开始的 JSON 库教程（十二）：JSON Path 实现

## JSON Path 简介

JSON Path 是一种用于从 JSON 文档中提取数据的查询语言，类似于 XML 的 XPath。它提供了一种简洁的语法来指定 JSON 结构中的位置，便于从复杂的嵌套 JSON 数据中查找、提取和操作数据。

JSON Path 的核心思想是通过路径表达式来定位 JSON 文档中的元素。Stefan Goessner 在 2007 年首次提出了这个概念，尽管目前尚未有官方标准，但已被广泛采用。

## JSON Path 语法

JSON Path 使用以下基本语法元素：

1. **$** - 根对象/元素
2. **@** - 当前对象/元素
3. **.** - 子元素操作符
4. **[]** - 下标操作符
5. **..** - 递归下降
6. **\*** - 通配符，表示所有对象/元素
7. **?()** - 过滤表达式
8. **()** - 脚本表达式
9. **,** - 并集操作

### 示例

假设有如下 JSON 数据：

```json
{
  "store": {
    "book": [
      {
        "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      {
        "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  }
}
```

常见的 JSON Path 表达式示例：

| JSON Path | 描述 |
|-----------|------|
| `$.store.book[*].author` | 所有书籍的作者 |
| `$..author` | 所有作者，无论在哪个层级 |
| `$.store.*` | store 对象的所有成员 |
| `$.store..price` | store 下所有价格 |
| `$..book[2]` | 第三本书 |
| `$..book[-1:]` | 最后一本书 |
| `$..book[0,1]` | 前两本书 |
| `$..book[:2]` | 前两本书（使用切片语法） |
| `$..book[?(@.price<10)]` | 所有价格小于 10 的书 |
| `$..book[?(@.category=="fiction")]` | 所有分类为 "fiction" 的书 |

## 本章实现功能

在本章中，我们实现了一个 JSON Path 解析器和求值器，支持以下功能：

1. **路径解析**：将 JSON Path 表达式解析为令牌序列
2. **属性访问**：支持点表示法 `.property` 和括号表示法 `['property']`
3. **数组索引**：支持通过索引访问数组元素，包括负索引
4. **数组切片**：支持类似 Python 的切片语法 `[start:end:step]`
5. **通配符**：支持 `*` 通配符，匹配所有属性或数组元素
6. **递归下降**：支持 `..` 操作符，在任意深度查找匹配的元素
7. **多种查询方法**：提供单值查询和多值查询功能

## 实现细节

### 主要数据结构

1. **JSONPath**: 表示解析后的 JSON Path 表达式
   ```go
   type JSONPath struct {
       Path   string  // 原始路径表达式
       Tokens []Token // 解析后的令牌列表
   }
   ```

2. **Token**: 表示 JSON Path 中的一个令牌
   ```go
   type Token struct {
       Type  TokenType // 令牌类型
       Value string    // 令牌值
   }
   ```

3. **TokenType**: 令牌类型枚举
   ```go
   type TokenType int
   
   const (
       ROOT TokenType = iota
       CURRENT
       DOT
       // ... 其他类型
   )
   ```

4. **SliceInfo**: 用于数组切片操作
   ```go
   type SliceInfo struct {
       Start int
       End   int
       Step  int
   }
   ```

### 解析过程

JSON Path 解析分为以下步骤：

1. 验证路径表达式的基本有效性
2. 从左到右遍历路径表达式，识别并创建令牌
3. 处理特殊情况，如属性名、数组索引、切片等
4. 生成令牌序列，表示路径表达式的结构

### 求值过程

对于给定的 JSON 文档和 JSON Path 表达式，求值过程是：

1. 从根节点开始，按照令牌序列依次求值
2. 对于每种令牌类型，应用相应的操作（属性访问、数组索引等）
3. 递归处理复杂操作，如通配符和递归下降
4. 收集满足条件的所有值，返回结果集

## 使用示例

### 基本用法

```go
// 创建 JSON 文档
doc := &Value{}
// ... 填充文档数据 ...

// 方法 1: 使用 NewJSONPath 创建 JSONPath 对象
path, err := NewJSONPath("$.store.book[0].title")
if err != nil {
    // 处理错误
}
results, err := path.Query(doc)
if err != nil {
    // 处理错误
}

// 方法 2: 使用便捷函数
results, err := QueryString(doc, "$.store.book[0].title")
if err != nil {
    // 处理错误
}

// 获取单个结果
result, err := QueryOneString(doc, "$.store.book[0].title")
if err != nil {
    // 处理错误
}
title := GetString(result) // 使用相应的获取器获取具体值
```

### 复杂查询示例

```go
// 获取所有书籍作者
authors, _ := QueryString(doc, "$.store.book[*].author")

// 找出所有价格（无论位置）
prices, _ := QueryString(doc, "$..price")

// 使用数组切片获取第一本到第二本书
firstTwoBooks, _ := QueryString(doc, "$.store.book[0:2]")

// 反向获取所有书籍
reversedBooks, _ := QueryString(doc, "$.store.book[::-1]")
```

## 支持的特性和限制

### 已支持的特性

- 基本路径导航（$, ., []）
- 属性访问（.property 和 ['property']）
- 数组索引访问，包括负索引
- 数组切片操作，支持起始、结束和步长
- 通配符（*）匹配
- 递归下降（..）操作
- 友好的错误消息

### 当前限制

- 不支持过滤表达式 ?()
- 不支持脚本表达式 ()
- 不支持联合操作符 [expr1,expr2,expr3]
- 不支持当前节点引用 @

### 未来扩展计划

- 实现过滤表达式支持
- 添加并集操作支持
- 优化性能，特别是对于大型 JSON 文档
- 添加更多错误恢复机制

## 参考资料

- [Stefan Goessner 的原始 JSON Path 文章](https://goessner.net/articles/JsonPath/)
- [JSONPath 在线求值器](https://jsonpath.com/)
- [JSON Path 表达式](https://support.smartbear.com/alertsite/docs/monitors/api/endpoint/jsonpath.html)
- [Jayway JsonPath (Java 实现)](https://github.com/json-path/JsonPath) 