package main

import (
	"fmt"
	"strings"

	"github.com/Cactusinhand/go-json-tutorial/tutorial09/leptjson"
)

func main() {
	// 示例1：基本错误处理
	fmt.Println("示例1：基本错误处理")
	fmt.Println("-------------------")

	jsonWithError := `{"name": "Bob", "age": invalid, "city": "Shanghai"}`
	var v leptjson.Value

	// 使用默认选项（不恢复错误）
	err := leptjson.Parse(&v, jsonWithError)
	fmt.Printf("解析结果：%v\n", err)
	fmt.Println()

	// 示例2：使用增强错误处理
	fmt.Println("示例2：使用增强错误处理（错误恢复）")
	fmt.Println("----------------------------------")

	options := leptjson.ParseOptions{
		RecoverFromErrors: true,
		AllowComments:     false,
		MaxDepth:          1000,
	}

	err = leptjson.ParseWithOptions(&v, jsonWithError, options)
	fmt.Printf("解析结果：%v（有错误但仍然继续解析）\n", err)
	fmt.Printf("解析后的值：%s\n", v.String())
	fmt.Println()

	// 示例3：解析带注释的JSON
	fmt.Println("示例3：解析带注释的JSON")
	fmt.Println("----------------------")

	jsonWithComments := `{
		// 这是用户信息
		"name": "Alice", /* 用户名 */
		"age": 28, // 年龄
		"address": {
			/* 详细地址信息 */
			"city": "Beijing",
			"postcode": "100000"
		}
	}`

	options.AllowComments = true
	err = leptjson.ParseWithOptions(&v, jsonWithComments, options)

	if err == leptjson.PARSE_OK {
		fmt.Println("成功解析带注释的JSON")
		fmt.Printf("解析后的值：%s\n", v.String())
	} else {
		fmt.Printf("解析失败：%v\n", err)
	}
	fmt.Println()

	// 示例4：嵌套深度限制
	fmt.Println("示例4：嵌套深度限制")
	fmt.Println("------------------")

	// 创建一个深度为10的嵌套JSON
	nestedJSON := ""
	for i := 0; i < 10; i++ {
		nestedJSON += "{"
		if i < 9 {
			nestedJSON += "\"nested\": "
		} else {
			nestedJSON += "\"value\": 42"
		}
	}
	nestedJSON += strings.Repeat("}", 10)

	// 设置最大深度为5
	options.MaxDepth = 5
	err = leptjson.ParseWithOptions(&v, nestedJSON, options)

	fmt.Printf("设置最大深度为5，解析深度为10的JSON结果：%v\n", err)

	// 增加最大深度到15
	options.MaxDepth = 15
	err = leptjson.ParseWithOptions(&v, nestedJSON, options)

	if err == leptjson.PARSE_OK {
		fmt.Println("设置最大深度为15，成功解析深度为10的JSON")
	} else {
		fmt.Printf("解析失败：%v\n", err)
	}
}
