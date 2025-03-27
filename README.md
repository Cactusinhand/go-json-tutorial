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

## 当前进度

当前已经完成了所有教程的实现，包括：

* null、true、false等字面值
* 数字（包括整数、浮点数、科学计数法）
* 字符串（支持转义序列和Unicode）
* 数组（支持嵌套）
* 对象（支持嵌套）
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

## 安装与使用

要使用这个库，你需要Go语言环境（推荐Go 1.16或更高版本）。克隆该仓库后，可以直接导入使用：

```go
import "github.com/Cactusinhand/go-json-tutorial/tutorial08"
// 或使用增强错误处理版本
import "github.com/Cactusinhand/go-json-tutorial/tutorial09"

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
}
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

## 测试

每个章节都包含完整的测试用例，测试覆盖了各种正常和边缘情况。运行测试：

```bash
go test ./tutorial08
```

## 注意事项

* 处理Unicode时，需要特别注意UTF-16代理对的处理
* 数字解析需要考虑各种边缘情况，如前导零、溢出等
* 字符串解析需要正确处理所有转义序列
* 数组和对象解析需要处理嵌套和边界情况
* 在使用动态数组和对象函数时，注意内存管理和资源释放

## 参考资料

* [JSON官方规范](https://www.json.org/)
* [原始C语言教程](https://github.com/miloyip/json-tutorial)
* [RFC 7159: The JavaScript Object Notation (JSON) Data Interchange Format](https://tools.ietf.org/html/rfc7159)
* [RFC 3629: UTF-8, a transformation format of ISO 10646](https://tools.ietf.org/html/rfc3629)
```