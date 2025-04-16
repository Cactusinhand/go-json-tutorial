# 从零开始的 JSON 库教程 (Go语言版)

* 基于 Milo Yip 的 C 语言 JSON 教程改编
* 2025年

这是一个使用Go语言实现的JSON库教程，基于[《从零开始的 JSON 库教程》](https://github.com/miloyip/json-tutorial)改编。本教程保持原教程的章节结构和渐进式开发方法，但根据Go语言的特性进行了适当调整。

## 对象与目标

教程对象：学习过基本Go语言编程的同学。

通过这个教程，同学可以了解如何从零开始写一个Go语言版本的JSON库，其特性如下：

* 符合标准的JSON解析器和生成器
* 使用Go语言的递归下降解析器（recursive descent parser）
* 跨平台（如Windows/Linux/macOS）
* 支持UTF-8 JSON文本
* 使用Go语言内置类型存储JSON数据
* 解析器和生成器的代码简洁高效
* 提供丰富的访问和修改API

除了围绕JSON作为例子，希望能在教程中讲述一些课题：

* 测试驱动开发（test driven development, TDD）
* Go语言编程风格
* 数据结构
* API设计
* 错误处理
* Unicode
* 浮点数
* Go模块、单元测试等工具和实践

## 学习价值与收获

本教程不仅是一个JSON库的实现指南，更是一次全面的编程能力提升之旅：

### 计算机科学基础
* **词法分析与语法分析**：通过实现递归下降解析器，实际应用编译原理的核心知识
* **数据结构应用**：深入理解树形结构的构建与操作
* **状态机思想**：在解析字符串和数字时应用有限状态机的概念
* **递归与栈**：通过处理嵌套JSON结构，掌握递归机制的实际应用

### Go语言特性
* **类型系统**：深入理解Go的结构体、接口和类型断言
* **错误处理机制**：掌握Go独特的错误处理模式
* **切片和映射**：通过实现数组和对象处理，熟练使用这些核心数据结构
* **字符串处理**：学习Go的Unicode支持和字符串操作

### 软件工程实践
* **测试驱动开发**：通过编写测试用例验证实现的正确性
* **API设计**：学习如何设计清晰、一致且易用的接口
* **性能优化**：在保证正确性的前提下提高解析和生成效率
* **错误处理策略**：实现健壮的错误处理和报告机制

完成本教程后，你不仅拥有了一个可用的JSON库，更重要的是获得了从零构建软件的信心和能力，以及对Go语言的深入理解，这些技能可以应用到各种实际项目开发中。

## 项目结构

教程按照章节组织，每个章节对应一个子目录：

1. tutorial01: 基础结构、解析null/true/false
2. tutorial02: 解析数字
3. tutorial03: 解析字符串
4. tutorial04: Unicode支持
5. tutorial05: 解析数组
6. tutorial06: 解析对象
7. tutorial07: 生成器
8. tutorial08: 访问与其他功能
9. tutorial09: 增强的错误处理
10. tutorial10: JSON指针实现
11. tutorial11: JSON Schema验证
12. tutorial12: JSON Path实现
13. tutorial13: JSON Patch实现
14. tutorial14: JSON Merge Patch实现
15. tutorial15: JSON序列化与Struct Tag支持
16. tutorial16: 命令行工具实现
17. tutorial17: 安全性增强

## 当前进度

当前已经完成了所有教程的实现，包括：

* 基础功能:
  * null、true、false等字面值解析
  * 数字解析（包括整数、浮点数、科学计数法）
  * 字符串解析（支持转义序列和Unicode）
  * 数组解析（支持嵌套）
  * 对象解析（支持嵌套）
  * JSON生成器 - 将JSON值转换为文本

* 高级功能:
  * 对象键值查询
  * JSON值比较
  * 深度复制、移动和交换
  * 动态数组和对象操作
  * 高效内存管理

* 增强错误处理:
  * 详细的错误信息与位置
  * 错误恢复能力
  * 自定义解析选项
  * 支持JSON注释

* 高级数据结构:
  * JSON指针实现
  * 支持循环引用检测

* JSON Schema验证:
  * 支持类型验证
  * 数值约束（最小值、最大值、倍数等）
  * 字符串约束（长度、模式、格式等）
  * 数组约束（元素数量、唯一性等）
  * 对象约束（属性数量、必需属性等）
  * 逻辑组合（allOf、anyOf、oneOf、not）

* JSON Path实现:
  * 支持属性访问和数组索引
  * 支持递归下降和通配符
  * 支持数组切片操作
  * 提供便捷的查询API

* JSON Patch实现:
  * 支持RFC 6902中定义的所有操作（add、remove、replace、move、copy、test）
  * 可以从两个JSON文档生成Patch
  * 优雅处理复杂的JSON文档修改

* JSON Merge Patch实现:
  * 支持RFC 7396中定义的合并补丁操作
  * 提供更简单直观的JSON文档修改方式
  * 完全兼容HTTP PATCH操作

* JSON序列化与Struct Tag支持:
  * 将Go数据结构转换为JSON文本
  * 基本类型序列化（布尔、数值、字符串、nil）
  * 复合类型序列化（数组、切片、映射、结构体）
  * 支持通过struct tag自定义序列化行为
  * 字段重命名、忽略和条件忽略(omitempty)
  * 循环引用检测防止无限递归

* 命令行工具实现:
  * JSON解析和格式验证功能
  * 格式化和最小化功能
  * JSON统计分析功能
  * 基于JSONPath的查询功能
  * JSON文档比较功能
  * Schema验证、指针操作和补丁应用
  * 用户友好的命令行界面和帮助文档

* 安全性增强:
  * 最大输入大小限制（防止内存溢出攻击）
  * 嵌套深度限制（防止栈溢出攻击）
  * 字符串长度限制（防止资源耗尽）
  * 数组元素数量限制（防止资源耗尽）
  * 对象成员数量限制（防止资源耗尽） 
  * 数值范围限制（防止精度攻击）
  * 可自定义安全策略

## 安装与使用

要使用这个库，你需要Go语言环境（推荐Go 1.16或更高版本）。克隆该仓库后，可以直接导入使用：

```go
import "github.com/Cactusinhand/go-json-tutorial/tutorial08"
// 或使用增强错误处理版本
import "github.com/Cactusinhand/go-json-tutorial/tutorial09"
// 或使用JSON指针功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial10"
// 或使用JSON Schema验证功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial11"
// 或使用JSON Path查询功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial12"
// 或使用JSON Patch功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial13"
// 或使用JSON Merge Patch功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial14"
// 或使用JSON序列化与Struct Tag支持
import "github.com/Cactusinhand/go-json-tutorial/tutorial15"
// 或使用命令行工具功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial16"
// 或使用安全性增强功能
import "github.com/Cactusinhand/go-json-tutorial/tutorial17"

func main() {
    // 基本解析JSON
    v := leptjson.Value{}
    if err := leptjson.Parse(&v, `{"name": "John", "age": 30}`); err != leptjson.PARSE_OK {
        // 处理错误
    }
    
    // 使用增强错误处理
    options := leptjson.ParseOptions{
        RecoverFromErrors: true,  // 启用错误恢复
        AllowComments: true,      // 允许JSON注释
        MaxDepth: 1000            // 设置最大嵌套深度
    }
    if err := leptjson.ParseWithOptions(&v, `{"name": "John", /* 这是注释 */ "age": 30}`, options); err != leptjson.PARSE_OK {
        // 错误处理但仍能继续解析
        fmt.Println("解析遇到错误但已恢复:", err)
    }
    
    // 访问解析后的数据
    name := leptjson.GetString(leptjson.GetObjectValueByKey(&v, "name"))
    age := leptjson.GetNumber(leptjson.GetObjectValueByKey(&v, "age"))
    
    // 使用FindObjectKey快速访问
    if value, found := leptjson.FindObjectKey(&v, "name"); found {
        name = leptjson.GetString(value)
    }
    
    // 修改JSON数据
    newPerson := &leptjson.Value{}
    leptjson.SetObject(newPerson)
    nameValue := leptjson.SetObjectValue(newPerson, "name")
    leptjson.SetString(nameValue, "Alice")
    
    // 动态数组操作
    arr := &leptjson.Value{}
    leptjson.SetArray(arr, 0)
    elem := leptjson.PushBackArrayElement(arr)
    leptjson.SetNumber(elem, 42)
    
    // 生成JSON字符串
    jsonStr, _ := leptjson.Stringify(newPerson)
    
    // 使用JSON指针
    pointer, _ := leptjson.NewJSONPointer("/users/0/name")
    if value, err := leptjson.Resolve(&v, pointer); err == nil {
        userName := leptjson.GetString(value)
        fmt.Println("用户名：", userName)
    }
    
    // 使用JSON Schema验证
    schemaJSON := `{
        "type": "object",
        "required": ["name", "age"],
        "properties": {
            "name": {"type": "string", "minLength": 2},
            "age": {"type": "integer", "minimum": 18}
        }
    }`
    schema, _ := leptjson.NewJSONSchema(schemaJSON)
    result := schema.Validate(&v)
    if !result.Valid {
        fmt.Println("数据不符合Schema:", result.Errors)
    }
    
    // 使用JSON Path查询
    // 查询所有书籍的作者
    doc := &leptjson.Value{} // 假设这是一个包含书籍信息的JSON
    leptjson.Parse(doc, `{
        "store": {
            "book": [
                {"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century"},
                {"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour"}
            ]
        }
    }`)
    
    // 方法1：使用JSONPath对象
    path, _ := leptjson.NewJSONPath("$.store.book[*].author")
    authors, _ := path.Query(doc)
    for _, author := range authors {
        fmt.Println("作者:", leptjson.GetString(author))
    }
    
    // 方法2：使用便捷函数
    allPrices, _ := leptjson.QueryString(doc, "$..price")
    fmt.Println("找到", len(allPrices), "个价格")
    
    // 获取单个结果
    firstBook, _ := leptjson.QueryOneString(doc, "$.store.book[0]")
    if firstBook != nil {
        bookTitle := leptjson.GetString(leptjson.GetObjectValueByKey(firstBook, "title"))
        fmt.Println("第一本书:", bookTitle)
    }
    
    // 使用JSON Patch
    // 创建一个JSON Patch
    patchJSON := `[
        {"op": "add", "path": "/user/name", "value": "John Doe"},
        {"op": "remove", "path": "/old_field"},
        {"op": "replace", "path": "/status", "value": "active"}
    ]`
    
    // 方法1：解析并应用Patch
    document := map[string]interface{}{
        "status": "pending",
        "old_field": "to be removed"
    }
    
    patch, _ := tutorial13.NewJSONPatch(patchJSON)
    resultDoc, _ := patch.Apply(document)
    
    // 方法2：生成两个文档之间的差异Patch
    source := map[string]interface{}{"a": 1, "b": 2}
    target := map[string]interface{}{"a": 1, "c": 3}
    
    diffPatch, _ := tutorial13.CreatePatch(source, target)
    patchStr, _ := diffPatch.String()
    fmt.Println("生成的Patch:", patchStr)
    // 输出: [{"op":"remove","path":"/b"},{"op":"add","path":"/c","value":3}]
    
    // 使用JSON Merge Patch
    // 创建一个Merge Patch
    mergePatchJSON := `{
        "title": "更新的标题",
        "author": {"email": null},
        "tags": ["news", "updated"],
        "content": "新内容"
    }`
    
    originalDoc := map[string]interface{}{
        "title": "原标题",
        "author": {
            "name": "作者名",
            "email": "author@example.com"
        },
        "tags": ["original"],
        "published": true
    }
    
    // 方法1：解析并应用Merge Patch
    var mergePatchObj interface{}
    tutorial14.Unmarshal([]byte(mergePatchJSON), &mergePatchObj)
    
    patch, _ := tutorial14.NewJSONMergePatch(mergePatchObj)
    updatedDoc, _ := patch.Apply(originalDoc)
    
    // 方法2：从两个文档生成Merge Patch
    sourceMerge := map[string]interface{}{
        "a": 1,
        "b": {"c": 3}
    }
    targetMerge := map[string]interface{}{
        "a": 1,
        "b": {"d": 4}
    }
    
    mergePatch, _ := tutorial14.CreateMergePatch(sourceMerge, targetMerge)
    mergePatchStr, _ := mergePatch.String()
    fmt.Println("生成的Merge Patch:", mergePatchStr)
    // 输出: {"b":{"c":null,"d":4}}
}

接下来展示JSON序列化与Struct Tag支持的使用示例：

```go
import "github.com/Cactusinhand/go-json-tutorial/tutorial15"

// 定义Go结构体
type Address struct {
    Street string `json:"streetName"`
    City   string `json:"city"`
}

type Person struct {
    Name    string   `json:"name"`              // 重命名字段
    Age     int      `json:"age,omitempty"`     // 零值时忽略
    Emails  []string `json:"emails"`            // 数组字段
    Address *Address `json:"address,omitempty"` // 嵌套结构体
    Extra   string   `json:"-"`                 // 忽略此字段
    Notes   string   // 没有tag，使用字段名
}

func marshalExample() {
    // 创建结构体实例
    person := Person{
        Name: "张三",
        Age:  30,
        Emails: []string{"zhangsan@example.com", "zs@work.com"},
        Address: &Address{
            Street: "中关村大街1号",
            City:   "北京",
        },
        Extra: "不会被序列化的内容",
        Notes: "客户备注",
    }
    
    // 将结构体序列化为JSON
    jsonStr, err := tutorial15.Marshal(person)
    if err == tutorial15.STRINGIFY_OK {
        fmt.Println(jsonStr)
        // 输出: {"name":"张三","age":30,"emails":["zhangsan@example.com","zs@work.com"],"address":{"streetName":"中关村大街1号","city":"北京"},"Notes":"客户备注"}
    } else {
        fmt.Println("序列化失败:", err)
    }
    
    // omitempty示例
    emptyPerson := Person{
        Name:   "李四",
        Emails: []string{},  // 非nil但为空的切片
        // Age和Address为零值，会被omitempty忽略
    }
    
    jsonStr, _ = tutorial15.Marshal(emptyPerson)
    fmt.Println(jsonStr)
    // 输出: {"name":"李四","emails":[],"Notes":""}
    
    // 复合类型示例
    data := map[string]interface{}{
        "numbers": []int{1, 2, 3},
        "active": true,
        "details": map[string]string{
            "department": "IT",
            "location": "北京"
        }
    }
    
    jsonStr, _ = tutorial15.Marshal(data)
    fmt.Println(jsonStr)
    // 输出类似: {"active":true,"details":{"department":"IT","location":"北京"},"numbers":[1,2,3]}
}

命令行工具的使用示例：

```bash
# 安装命令行工具
$ go install github.com/Cactusinhand/go-json-tutorial/tutorial16/main

# 验证JSON文件
$ leptjson parse data.json

# 格式化JSON文件（使用4个空格缩进）
$ leptjson format --indent=4 data.json formatted.json

# 最小化JSON文件（移除所有空白字符）
$ leptjson minify formatted.json data.min.json

# 查看JSON统计信息
$ leptjson stats data.json

# 使用JSONPath查询数据
$ leptjson path data.json "$.store.book[*].author"

# 比较两个JSON文件
$ leptjson compare original.json updated.json

# 使用JSON Schema验证文件
$ leptjson validate schema.json data.json

# 使用JSON Pointer获取特定值
$ leptjson pointer data.json "/users/0/name"

# 应用JSON Patch
$ leptjson patch patch.json data.json result.json
```

## 关键API说明

### 解析和生成
* `Parse(&v, json)`: 解析JSON文本到Value结构
* `ParseWithOptions(&v, json, options)`: 使用自定义选项解析JSON文本
* `Stringify(&v)`: 将Value结构转换为JSON文本

### 类型和值访问
* `GetType(&v)`: 获取JSON值的类型
* `GetBoolean(&v)`, `SetBoolean(&v, b)`: 获取/设置布尔值
* `GetNumber(&v)`, `SetNumber(&v, n)`: 获取/设置数字值
* `GetString(&v)`, `SetString(&v, s)`: 获取/设置字符串值

### 数组操作
* `GetArraySize(&v)`: 获取数组大小
* `GetArrayElement(&v, index)`: 获取数组元素
* `SetArray(&v, capacity)`: 设置为数组类型
* `PushBackArrayElement(&v)`: 添加数组元素
* `PopBackArrayElement(&v)`: 移除最后一个元素
* `InsertArrayElement(&v, index)`: 插入元素
* `EraseArrayElement(&v, index, count)`: 删除元素
* `ClearArray(&v)`: 清空数组

### 对象操作
* `GetObjectSize(&v)`: 获取对象大小
* `GetObjectKey(&v, index)`: 获取对象键
* `GetObjectValue(&v, index)`: 获取对象值
* `GetObjectValueByKey(&v, key)`: 根据键获取对象值
* `FindObjectKey(&v, key)`: 查找键并返回值
* `SetObject(&v)`: 设置为对象类型
* `SetObjectValue(&v, key)`: 设置对象键值
* `RemoveObjectValue(&v, index)`: 移除成员
* `ClearObject(&v)`: 清空对象

### JSON指针
* `NewJSONPointer(ptr)`: 创建新的JSON指针
* `Resolve(&v, pointer)`: 解析JSON指针并返回引用的值
* `Contains(&v, pointer)`: 检查指针是否能解析到值
* `Add(&v, pointer, value)`: 添加或替换指针引用的值
* `Remove(&v, pointer)`: 移除指针引用的值
* `GetTokens(pointer)`: 获取指针的令牌数组

### JSON Schema验证
* `NewJSONSchema(schemaJSON)`: 创建新的JSON Schema
* `Validate(&v)`: 验证值是否符合Schema
* `SchemaValidationResult`: 包含验证结果和错误信息
* `SchemaValidationError`: 表示具体的验证错误，包含路径和消息

### JSON Path查询
* `NewJSONPath(path)`: 创建新的JSON Path查询对象
* `Query(doc)`: 使用路径表达式查询JSON文档
* `QueryOne(doc)`: 返回第一个匹配的值
* `QueryString(doc, path)`: 便捷函数，直接使用路径字符串查询
* `QueryOneString(doc, path)`: 便捷函数，返回第一个匹配值

### 内存和资源管理
* `Copy(dst, src)`: 深度复制JSON值
* `Move(dst, src)`: 移动JSON值
* `Swap(lhs, rhs)`: 交换两个JSON值
* `Free(&v)`: 释放资源

### 比较和其他
* `Equal(lhs, rhs)`: 比较两个JSON值是否相等

### 增强错误处理
* `ParseOptions`: 定义解析选项的结构体，包含错误恢复、注释支持和最大深度等选项
* `EnhancedError`: 提供详细的错误信息，包括行号、列号、上下文和错误位置指示
* `GetErrorMessage(code)`: 获取特定错误码的错误描述信息

### 安全选项
* `EnabledSecurity`: 是否启用安全检查（默认开启）
* `MaxTotalSize`: 限制JSON输入的总字节数
* `MaxStringLength`: 限制字符串值的最大长度
* `MaxArraySize`: 限制数组允许的最大元素数量
* `MaxObjectSize`: 限制对象允许的最大成员数量 
* `MaxDepth`: 限制嵌套结构的最大深度
* `MaxNumberValue`: 限制数值的最大值
* `MinNumberValue`: 限制数值的最小值

### 高级API与性能优化
* `NewStreamParser()`: 创建流式解析器
* `ParseStream(reader)`: 从IO读取器流式解析JSON
* `NewValuePool(size)`: 创建值对象池，减少内存分配
* `ParseConcurrent(jsonTexts, goroutines)`: 并发解析多个JSON文本
* `LazyParse(json)`: 惰性解析JSON，只在实际访问时解析
* `SetPoolSize(size)`: 设置内部缓冲池大小
* `EnableZeroCopy(enabled)`: 启用/禁用零拷贝模式

### Go与JSON深度集成
* `Marshal(v interface{})`: 将Go值转换为JSON文本
* `Unmarshal(data []byte, v interface{})`: 将JSON文本解析为Go值
* `GetValue[T any](json, path)`: 泛型函数，直接获取指定类型的值
* `RegisterTypeHandler(handler)`: 注册自定义类型处理器
* `SetFieldNamingStrategy(strategy)`: 设置字段命名策略
* `MarshalWithOptions(v, options)`: 使用自定义选项进行转换
* `UnmarshalWithOptions(data, v, options)`: 使用自定义选项进行解析

## 测试

每个章节都包含完整的测试用例，测试覆盖了各种正常和边缘情况。运行测试：

```bash
# 测试基本功能
go test ./tutorial08

# 测试安全特性
go test ./tutorial17

# 运行性能测试，比较安全检查开启与关闭的性能差异
go test -bench=BenchmarkParseSecurity ./tutorial17
```

## 注意事项

* 处理Unicode时，需要特别注意UTF-16代理对的处理
* 数字解析需要考虑各种边缘情况，如前导零、溢出等
* 字符串解析需要正确处理所有转义序列
* 数组和对象解析需要处理嵌套和边界情况
* 在使用动态数组和对象函数时，注意内存管理和资源释放
* JSON Schema验证仅实现了部分Draft 7规范，用于学习目的
* JSON Path实现不支持过滤表达式和脚本表达式
* 开启安全检查会对解析性能产生一定影响，在性能敏感场景可以考虑关闭部分检查
* 默认的安全限制较为宽松，在实际应用中应根据具体需求调整各项安全参数
* 在解析不受信任的JSON数据时，强烈建议启用所有安全检查

## 参考资料

* [JSON官方规范](https://www.json.org/)
* [原始C语言教程](https://github.com/miloyip/json-tutorial)
* [RFC 7159: The JavaScript Object Notation (JSON) Data Interchange Format](https://tools.ietf.org/html/rfc7159)
* [RFC 3629: UTF-8, a transformation format of ISO 10646](https://tools.ietf.org/html/rfc3629)
* [RFC 6901: JavaScript Object Notation (JSON) Pointer](https://tools.ietf.org/html/rfc6901)
* [JSON Schema Draft 7](https://json-schema.org/specification-links.html#draft-7)
* [Stefan Goessner的JSONPath文章](https://goessner.net/articles/JsonPath/)
* [OWASP JSON安全备忘录](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Security_Cheat_Sheet.html)
* [OWASP拒绝服务攻击防护指南](https://owasp.org/www-community/attacks/Denial_of_Service)