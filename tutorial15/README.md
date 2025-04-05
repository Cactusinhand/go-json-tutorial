# 从零开始的 JSON 库教程（十五）：JSON 序列化 (Marshal) 与 Struct Tag 支持

## JSON 序列化简介

JSON 序列化（Marshal）是将程序中的数据结构转换为 JSON 格式字符串的过程。在本章中，我们为 `leptjson` 库添加了将 Go 数据结构序列化为 JSON 字符串的功能。核心是实现了 `Marshal(v interface{}) (string, ErrorCode)` 函数，它能够处理多种 Go 数据类型并支持通过 Struct Tag 自定义序列化行为。

## 主要功能

*   **`Marshal(v interface{}) (string, ErrorCode)`**: 接受任意 Go 数据类型 `v`，尝试将其转换为对应的 JSON 字符串。返回 JSON 字符串和可能的错误码。
*   **支持多种 Go 类型**: 
    *   基本类型: `bool`, `int`, `uint`, `float`, `string`
    *   `nil` (指针或接口)
    *   Slice 和 Array (转换为 JSON 数组)
    *   Map (键必须是 `string` 类型，转换为 JSON 对象)
    *   Struct (转换为 JSON 对象)
*   **Struct Tag 支持**: 
    *   使用 `json:"<name>"` 指定 JSON 对象中的键名。
    *   使用 `json:"-"` 忽略该字段。
    *   使用 `json:",omitempty"` 在字段值为其类型的零值（如 0, false, "", nil slice/map, nil pointer/interface）时忽略该字段。
    *   支持组合，如 `json:"myName,omitempty"`。
*   **循环引用检测**: 在处理指针类型时，能检测并阻止因循环引用导致的无限递归，返回 `MARSHAL_CYCLIC_REFERENCE` 错误。
*   **错误处理**: 对于不支持的类型（如 `complex`, `chan`, `func`，或非字符串键的 map），返回 `MARSHAL_UNSUPPORTED_TYPE` 错误。

## 使用示例

```go
package main

import (
	"fmt"
	leptjson "github.com/Cactusinhand/go-json-tutorial/tutorial15"
)

type Address struct {
	Street string `json:"streetName"`
	City   string `json:"city"`
}

type Person struct {
	Name    string   `json:"name"`
	Age     int      `json:"age,omitempty"`
	Emails  []string `json:"emails"`
	Address *Address `json:"address,omitempty"`
	Extra   string   `json:"-"`           // 忽略此字段
	Notes   string   // 没有 tag，使用字段名 "Notes"
}

func main() {
	// 示例 1: 完整 Person 结构体
	p1 := Person{
		Name:   "Alice",
		Age:    30,
		Emails: []string{"alice@example.com", "alice.work@example.com"},
		Address: &Address{
			Street: "123 Main St",
			City:   "Anytown",
		},
		Extra: "一些额外信息",
		Notes: "重要客户",
	}
	jsonStr1, err1 := leptjson.Marshal(p1)
	if err1 == leptjson.STRINGIFY_OK {
		fmt.Println("Person 1 JSON:", jsonStr1)
		// 输出类似: {"name":"Alice","age":30,"emails":["alice@example.com","alice.work@example.com"],"address":{"streetName":"123 Main St","city":"Anytown"},"Notes":"重要客户"}
	} else {
		fmt.Println("序列化 Person 1 失败:", err1)
	}

	// 示例 2: 使用 omitempty 的 Person 结构体
	p2 := Person{
		Name:   "Bob",
		Emails: []string{}, // 空 slice
		Notes:  "",        // 空字符串 (string 的零值)
		// Age 和 Address 都是零值 (0 和 nil), Extra 被忽略
	}
	jsonStr2, err2 := leptjson.Marshal(p2)
	if err2 == leptjson.STRINGIFY_OK {
		fmt.Println("\nPerson 2 JSON (omitempty):", jsonStr2)
		// 输出: {"name":"Bob","emails":[]}
	} else {
		fmt.Println("序列化 Person 2 失败:", err2)
	}

	// 示例 3: Map
	dataMap := map[string]interface{}{
		"isActive": true,
		"score":    95.5,
		"items":    []int{1, 2, 3},
	}
	jsonStr3, err3 := leptjson.Marshal(dataMap)
	if err3 == leptjson.STRINGIFY_OK {
		fmt.Println("\nMap JSON:", jsonStr3)
		// 输出类似: {"isActive":true,"items":[1,2,3],"score":95.5}
	} else {
		fmt.Println("序列化 Map 失败:", err3)
	}

	// 示例 4: Slice
	sliceData := []interface{}{nil, true, 10, "hello", []int{}}
	jsonStr4, err4 := leptjson.Marshal(sliceData)
	if err4 == leptjson.STRINGIFY_OK {
		fmt.Println("\nSlice JSON:", jsonStr4)
		// 输出: [null,true,10,"hello",[]]
	} else {
		fmt.Println("序列化 Slice 失败:", err4)
	}
}
```

## Struct Tag 详解

Go 语言的 struct tag 是一种元数据特性，允许在结构体字段上附加额外信息。在 JSON 序列化时，可以使用 `json` 标签来控制字段如何被序列化：

1. **重命名字段**: `json:"fieldName"`
   - 将 Go 结构体字段映射到不同的 JSON 属性名

2. **忽略字段**: `json:"-"`
   - 完全忽略该字段，不在 JSON 输出中包含

3. **条件忽略**: `json:",omitempty"`
   - 当字段值为零值时忽略该字段
   - 对于不同类型，零值定义如下：
     - 数值类型: 0
     - 布尔类型: false
     - 字符串: ""
     - 指针、接口: nil
     - 切片、映射: nil 或 长度为 0

4. **组合选项**: `json:"fieldName,omitempty"`
   - 同时重命名字段并在值为零值时忽略

## 本章实现目标

在本章中，我们将实现一个功能完整的 JSON 序列化器，可以将 Go 数据结构转换为 JSON 字符串，具体目标包括：

1. 实现 `Marshal` 函数，支持所有常用的 Go 数据类型
2. 支持 struct tag 功能，允许自定义 JSON 输出
3. 实现循环引用检测，防止无限递归
4. 提供合适的错误处理机制
5. 确保输出符合 JSON 规范

## 实现计划

1. 创建序列化相关的数据结构和函数
2. 实现基本类型（布尔、数值、字符串、nil）的序列化
3. 实现复合类型（数组、切片、映射）的序列化
4. 实现结构体序列化，包括 struct tag 解析和应用
5. 添加循环引用检测
6. 实现错误处理逻辑

## JSON Marshal 与 Unmarshal 的对比

| 特性 | Marshal (序列化) | Unmarshal (反序列化) |
|------|-----------------|-------------------|
| 方向 | Go 数据结构 → JSON 字符串 | JSON 字符串 → Go 数据结构 |
| 复杂度 | 相对简单，直接访问 Go 数据 | 较复杂，需要处理各种 JSON 解析情况 |
| 类型处理 | 从已知类型生成通用格式 | 将通用格式转换为特定类型 |
| Struct Tag | 控制输出的字段名和条件 | 控制如何填充结构体字段 |
| 错误情况 | 较少，主要是类型不支持和循环引用 | 较多，包括格式错误、类型不匹配等 |

## 测试

在 `go-json-tutorial/tutorial15` 目录下运行：

```bash
go test -run TestMarshal
```

## 参考资料

- [Go 官方 JSON 包文档](https://golang.org/pkg/encoding/json/)
- [Go JSON Marshal 实现细节](https://cs.opensource.google/go/go/+/master:src/encoding/json/encode.go)
- [Go Struct Tags 指南](https://medium.com/golangspec/tags-in-golang-3e5db0b8ef3e) 