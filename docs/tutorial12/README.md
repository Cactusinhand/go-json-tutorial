# 第12章：JSON Path 实现

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

我们首先定义一些核心数据结构来表示 JSON Path：

```go
// TokenType 表示 JSON Path 令牌的类型
type TokenType int

const (
    ROOT              TokenType = iota // $ - 根节点
    CURRENT                            // @ - 当前节点
    DOT                                // . - 子属性访问
    RECURSIVE_DESCENT                  // .. - 递归下降
    WILDCARD                           // * - 通配符
    BRACKET_START                      // [ - 下标访问开始
    BRACKET_END                        // ] - 下标访问结束
    INDEX                              // 数字索引
    PROPERTY                           // 属性名
    SLICE                              // 切片 [start:end:step]
    UNION                              // 并集 [expr,expr]
    FILTER                             // ?() - 过滤器
)

// Token 表示 JSON Path 中的一个令牌
type Token struct {
    Type  TokenType // 令牌类型
    Value string    // 令牌值
}

// SliceInfo 存储数组切片信息
type SliceInfo struct {
    Start int
    End   int
    Step  int
}

// JSONPath 表示一个解析后的 JSON Path 表达式
type JSONPath struct {
    Path   string  // 原始路径表达式
    Tokens []Token // 令牌列表
}
```

### 路径解析

解析 JSON Path 表达式是实现的第一步。我们需要将字符串形式的表达式转换为结构化的令牌序列：

```go
// 将 JSON Path 表达式解析为令牌列表
func (jp *JSONPath) parse() error {
    // 基本验证
    if len(jp.Path) == 0 {
        return &JSONPathError{Path: jp.Path, Message: "JSON Path 表达式不能为空"}
    }

    // 确保路径以 $ 开头
    if !strings.HasPrefix(jp.Path, "$") {
        return &JSONPathError{Path: jp.Path, Message: "JSON Path 必须以 $ 开始"}
    }

    // 将字符串解析为令牌
    jp.Tokens = []Token{{Type: ROOT, Value: "$"}}

    i := 1 // 跳过 $
    for i < len(jp.Path) {
        switch jp.Path[i] {
        case '.':
            if i+1 < len(jp.Path) && jp.Path[i+1] == '.' {
                // 递归下降 ..
                jp.Tokens = append(jp.Tokens, Token{Type: RECURSIVE_DESCENT, Value: ".."})
                i += 2
            } else {
                // 子属性访问 .
                jp.Tokens = append(jp.Tokens, Token{Type: DOT, Value: "."})
                i++
            }

            // 处理属性名或通配符
            if i < len(jp.Path) {
                if jp.Path[i] == '*' {
                    jp.Tokens = append(jp.Tokens, Token{Type: WILDCARD, Value: "*"})
                    i++
                } else if isValidPropertyNameStart(jp.Path[i]) {
                    propName, newPos := parsePropertyName(jp.Path, i)
                    jp.Tokens = append(jp.Tokens, Token{Type: PROPERTY, Value: propName})
                    i = newPos
                }
            }

        case '[':
            // 处理中括号表达式
            // ...

        default:
            i++ // 跳过不识别的字符
        }
    }

    return nil
}
```

### 求值过程

获取令牌列表后，我们就可以对 JSON 文档进行求值了：

```go
// evaluate 对当前节点应用 JSON Path 表达式的剩余部分
func (jp *JSONPath) evaluate(current *Value, tokenIndex int) ([]*Value, error) {
    // 如果已处理完所有令牌，返回当前节点
    if tokenIndex >= len(jp.Tokens) {
        return []*Value{current}, nil
    }

    token := jp.Tokens[tokenIndex]

    switch token.Type {
    case DOT:
        // 如果下一个令牌是属性名或通配符，解析它
        if tokenIndex+1 < len(jp.Tokens) {
            nextToken := jp.Tokens[tokenIndex+1]
            if nextToken.Type == PROPERTY {
                // 处理属性访问，如 .name
                matches, err := jp.matchProperty(current, tokenIndex+1)
                if err != nil {
                    return nil, err
                }
                
                // 递归处理剩余令牌
                var results []*Value
                for _, match := range matches {
                    subResults, err := jp.evaluate(match, tokenIndex+2)
                    if err != nil {
                        return nil, err
                    }
                    results = append(results, subResults...)
                }
                return results, nil
            } else if nextToken.Type == WILDCARD {
                // 处理通配符访问，如 .*
                matches, err := jp.matchWildcard(current, tokenIndex+1)
                if err != nil {
                    return nil, err
                }
                
                // 递归处理剩余令牌
                var results []*Value
                for _, match := range matches {
                    subResults, err := jp.evaluate(match, tokenIndex+2)
                    if err != nil {
                        return nil, err
                    }
                    results = append(results, subResults...)
                }
                return results, nil
            }
        }
        return nil, &JSONPathError{Path: jp.Path, Message: "在 . 后期望属性名或通配符"}

    case RECURSIVE_DESCENT:
        // 处理递归下降 ..
        return jp.findRecursive(current, tokenIndex)

    case BRACKET_START:
        // 处理中括号表达式 [...]
        // ...
    
    default:
        return nil, &JSONPathError{
            Path:    jp.Path,
            Message: fmt.Sprintf("意外的令牌类型: %v", token.Type),
            Index:   tokenIndex,
        }
    }
}
```

### 处理特殊语法

让我们看看如何实现一些特殊的 JSON Path 功能：

#### 数组切片处理

```go
// 处理数组切片，如 [1:3] 或 [::-1]
func parseSliceParams(sliceStr string) (*SliceInfo, error) {
    parts := strings.Split(sliceStr, ":")
    if len(parts) > 3 {
        return nil, fmt.Errorf("切片表达式格式不正确: %s", sliceStr)
    }
    
    info := &SliceInfo{
        Start: 0,
        End:   -1,
        Step:  1,
    }
    
    // 解析 start
    if parts[0] != "" {
        start, err := strconv.Atoi(parts[0])
        if err != nil {
            return nil, err
        }
        info.Start = start
    }
    
    // 解析 end (如果存在)
    if len(parts) > 1 && parts[1] != "" {
        end, err := strconv.Atoi(parts[1])
        if err != nil {
            return nil, err
        }
        info.End = end
    }
    
    // 解析 step (如果存在)
    if len(parts) > 2 && parts[2] != "" {
        step, err := strconv.Atoi(parts[2])
        if err != nil {
            return nil, err
        }
        if step == 0 {
            return nil, fmt.Errorf("切片步长不能为零")
        }
        info.Step = step
    }
    
    return info, nil
}
```

#### 递归下降搜索

```go
// 递归下降搜索
func (jp *JSONPath) findRecursive(current *Value, tokenIndex int) ([]*Value, error) {
    var result []*Value
    
    // 在当前级别应用下一个表达式
    if tokenIndex+1 < len(jp.Tokens) {
        nextToken := jp.Tokens[tokenIndex+1]
        
        // 先对当前节点尝试剩余的表达式
        if nextToken.Type == PROPERTY || nextToken.Type == WILDCARD {
            matches, err := jp.evaluate(current, tokenIndex+1)
            if err != nil {
                return nil, err
            }
            result = append(result, matches...)
        }
        
        // 然后递归搜索子节点
        if current.Type == OBJECT {
            for i := 0; i < current.Object.Size(); i++ {
                member := current.Object.GetMember(i)
                subMatches, err := jp.findRecursive(&member.Value, tokenIndex)
                if err != nil {
                    return nil, err
                }
                result = append(result, subMatches...)
            }
        } else if current.Type == ARRAY {
            for i := 0; i < current.Array.Size(); i++ {
                element := current.Array.Get(i)
                subMatches, err := jp.findRecursive(element, tokenIndex)
                if err != nil {
                    return nil, err
                }
                result = append(result, subMatches...)
            }
        }
    }
    
    return result, nil
}
```

## 用法示例

下面是如何使用我们的 JSON Path 实现的示例：

```go
package main

import (
    "fmt"
    "leptjson"
)

func main() {
    // 解析 JSON 数据
    jsonData := `{
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
    }`
    
    var doc leptjson.Value
    err := leptjson.Parse(&doc, jsonData)
    if err != nil {
        fmt.Printf("解析错误: %v\n", err)
        return
    }
    
    // 使用 JSON Path 查询 - 所有书籍的作者
    results, err := leptjson.QueryString(&doc, "$.store.book[*].author")
    if err != nil {
        fmt.Printf("查询错误: %v\n", err)
        return
    }
    
    fmt.Println("书籍作者:")
    for i, result := range results {
        author := leptjson.GetString(result)
        fmt.Printf("  %d: %s\n", i+1, author)
    }
    
    // 使用递归下降查找所有价格
    results, err = leptjson.QueryString(&doc, "$..price")
    if err != nil {
        fmt.Printf("查询错误: %v\n", err)
        return
    }
    
    fmt.Println("\n所有价格:")
    for i, result := range results {
        price := leptjson.GetNumber(result)
        fmt.Printf("  %d: %.2f\n", i+1, price)
    }
    
    // 使用过滤器查找便宜的书籍
    results, err = leptjson.QueryString(&doc, "$..book[?(@.price < 10)].title")
    if err != nil {
        fmt.Printf("查询错误: %v\n", err)
        return
    }
    
    fmt.Println("\n价格低于 10 的书籍:")
    for i, result := range results {
        title := leptjson.GetString(result)
        fmt.Printf("  %d: %s\n", i+1, title)
    }
}
```

这个程序的输出应该类似于：

```
书籍作者:
  1: Nigel Rees
  2: Evelyn Waugh

所有价格:
  1: 8.95
  2: 12.99
  3: 19.95

价格低于 10 的书籍:
  1: Sayings of the Century
```

## 完整实现

由于完整的 JSON Path 实现相当复杂，我们只在这里展示了核心部分。完整的实现包括：

1. 支持带引号和不带引号的属性名
2. 支持负索引和切片语法
3. 支持通配符 `*`
4. 支持递归下降 `..`
5. 支持过滤表达式 `?()` (有限支持)
6. 支持并集操作 `,`

完整的代码可以在项目的源文件中找到。

## 练习

1. 扩展实现过滤器表达式，支持更多的运算符，如大于等于、小于等于、不等于等。

2. 添加对正则表达式的支持，例如 `$..book[?(@.author =~ /.*Rees/)]`。

3. 实现脚本表达式支持，例如 `$..book[(@.length-1)]`。

4. 添加对自定义函数的支持，例如 `$..book[?(@.price == min($..price))]`。

5. 优化递归下降搜索的性能，特别是在大型 JSON 文档上。

## 下一步

在本章中，我们实现了 JSON Path 查询功能，使我们的 JSON 库能够方便地从复杂的 JSON 文档中提取数据。在下一章中，我们将探讨 JSON Patch 操作，它允许我们以标准化的方式修改 JSON 文档。