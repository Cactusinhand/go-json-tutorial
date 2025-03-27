# 从零开始的 JSON 库教程（九）：错误处理增强

## 高级教程开始

从本章开始，我们将进入 JSON 库的高级教程部分，着重于提升库的实用性、性能和安全性。在基础教程（教程一至八）中，我们已经实现了一个功能完整的 JSON 库，包括解析、生成和各种操作功能。现在，我们将进一步完善这个库，使其更加健壮和实用。

## 教程九：增强错误处理

在本教程中，我们将为JSON解析库添加增强的错误处理功能，提供更丰富的错误信息和更强大的恢复机制。这些功能对于构建健壮的应用程序至关重要，尤其是在处理用户输入或外部数据源时。

## 主要内容

1. **详细的错误信息**：
   - 不仅显示错误类型，还提供行号和列号
   - 显示错误上下文（出错位置附近的代码）
   - 提供指向错误位置的指针，直观展示错误位置

2. **错误恢复机制**：
   - 允许从某些错误中恢复并继续解析
   - 可配置的错误恢复策略
   - 在遇到非致命错误时用null替代错误值

3. **解析选项**：
   - 支持JSON注释（单行//和多行/* */）
   - 可配置的嵌套深度限制
   - 错误恢复开关

4. **改进的错误处理接口**：
   - 使用结构化的错误信息
   - 更好的位置跟踪
   - 支持国际化错误消息

## 实现要点

### 1. 增强的错误信息

我们实现了`EnhancedError`结构体，它包含以下信息：

- 错误代码：指示错误的类型
- 错误消息：人类可读的错误描述
- 行号和列号：错误发生的位置
- 上下文：错误发生附近的代码片段
- 指针：指向错误位置的可视化指示器
- 源码输入：原始JSON输入
- 是否可恢复：指示该错误是否可以从中恢复

### 2. 解析上下文改进

- 跟踪行号和列号
- 记录每行的起始位置以便定位错误
- 支持嵌套深度限制和检查
- 增加了对注释解析的支持

### 3. 错误恢复策略

允许从某些非致命错误中恢复：

- 数组中的无效值
- 对象中的无效值
- 格式错误的字符串
- 其他可恢复的语法错误

## 使用示例

基本解析（与之前版本兼容）：

```go
var v Value
err := Parse(&v, jsonText)
if err != PARSE_OK {
    // 处理错误
}
```

使用增强选项：

```go
var v Value
options := ParseOptions{
    RecoverFromErrors: true,  // 启用错误恢复
    AllowComments: true,      // 允许注释
    MaxDepth: 100             // 设置最大嵌套深度
}
err := ParseWithOptions(&v, jsonText, options)
```

## 性能考虑

虽然增强的错误处理提供了更多功能，但也带来了一些性能开销：

- 跟踪位置信息需要额外的计算
- 错误恢复需要额外的状态管理
- 注释解析增加了额外的代码路径

在性能关键的应用中，可以通过配置选项来禁用某些高开销的功能。

## 结论

本教程通过增强错误处理系统，大大提高了JSON解析库的用户友好性和健壮性。这些改进使库更适合在生产环境中使用，特别是在处理不可信数据时。通过提供详细的错误信息和恢复机制，可以帮助开发者更快地定位和解决问题。

# Tutorial 10 - 高级数据结构

本章实现了两个重要的高级数据结构功能：

1. **JSON 指针（JSON Pointer）**：基于 RFC 6901 标准实现，可以使用类似路径的语法访问和操作 JSON 文档中的特定部分。
2. **循环引用检测与处理**：检测和处理 JSON 数据中的循环引用问题，解决了深层复制和序列化时的潜在问题。

## JSON 指针功能

JSON 指针是一种用于指向 JSON 文档特定部分的字符串语法。例如，指针字符串 `/foo/0/bar` 表示访问对象的 "foo" 键，然后访问数组的第一个元素，最后访问该元素的 "bar" 键。

### 主要特性

- 完整支持 RFC 6901 标准
- 支持转义字符：`~0` 表示 `~`，`~1` 表示 `/`
- 提供获取、设置和删除值的操作
- 支持对象属性和数组索引的访问

### 使用示例

```go
// 解析 JSON 数据
doc := &leptjson.Value{}
leptjson.Parse(doc, `{"foo": {"bar": 42}, "baz": [0, 1, {"qux": "hello"}]}`)

// 使用 JSON 指针获取值
value, err := leptjson.GetValueByPointer(doc, "/foo/bar")
if err == nil {
    fmt.Println(leptjson.GetNumber(value)) // 输出: 42
}

// 使用 JSON 指针设置值
newValue := &leptjson.Value{}
leptjson.SetNumber(newValue, 100)
leptjson.SetValueByPointer(doc, "/foo/bar", newValue)

// 使用 JSON 指针删除值
leptjson.RemoveValueByPointer(doc, "/baz/0")

// 构建 JSON 指针字符串
pointerStr, _ := leptjson.BuildJSONPointer("foo", "bar")
// 结果: "/foo/bar"
```

## 循环引用检测与处理

循环引用是指 JSON 结构中存在环形引用关系的情况，这在标准 JSON 中是不允许的，但在内存中的表示可能会出现。

### 主要特性

- 精确检测对象和数组中的循环引用
- 安全复制功能，避免无限递归
- 提供循环引用替换策略
- 支持自定义替换器函数

### 使用示例

```go
// 检测循环引用
if leptjson.HasCycle(doc) {
    fmt.Println("JSON 数据包含循环引用")
}

// 安全复制（会在发现循环引用时返回错误）
dst := &leptjson.Value{}
if err := leptjson.CopySafe(dst, doc); err != nil {
    fmt.Println("无法复制，因为存在循环引用")
}

// 使用默认替换器进行安全复制
dst := &leptjson.Value{}
leptjson.CopySafeWithReplacement(dst, doc)
// 循环引用会被替换为字符串 "循环引用 -> /path/to/cycle"

// 使用自定义替换器
customReplacer := func(path []string) *leptjson.Value {
    v := &leptjson.Value{}
    leptjson.SetObject(v)
    
    cycleType := leptjson.SetObjectValue(v, "type")
    leptjson.SetString(cycleType, "cycle")
    
    cyclePath := leptjson.SetObjectValue(v, "path")
    leptjson.SetString(cyclePath, strings.Join(path, "/"))
    
    return v
}

dst := &leptjson.Value{}
leptjson.CustomCopySafeWithReplacement(dst, doc, customReplacer)
```

## 技术实现

### JSON 指针

- 使用令牌数组（tokens）表示路径段
- 实现了完整的 RFC 6901 转义规则
- 支持相对路径操作

### 循环引用检测

- 使用栈结构跟踪访问路径
- 实现了高效的循环检测算法
- 使用访问过的节点映射优化替换过程

## 使用建议

- 使用 JSON 指针简化复杂 JSON 操作
- 在处理不可信数据时，使用循环引用检测保证安全
- 为关键操作添加错误处理
- 使用自定义替换器创建有意义的循环引用表示

## 后续改进计划

- 支持 JSON Patch（RFC 6902）
- 支持 JSON Merge Patch（RFC 7396）
- 实现 JSON Schema 验证
- 添加更多优化和性能改进