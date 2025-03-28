// commands.go - 处理命令行命令的工具函数
package jsonutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Cactusinhand/go-json-tutorial/tutorial15/leptjson"
)

// PrintUsage 打印使用说明
func PrintUsage(progName string) {
	fmt.Printf("JSON工具 - Go版 v1.0.0\n")
	fmt.Printf("用法: %s <命令> [参数...]\n\n", progName)
	fmt.Printf("可用命令:\n")
	fmt.Printf("  parse <json文件>        - 解析JSON文件并验证其有效性\n")
	fmt.Printf("  format <json文件>       - 格式化JSON文件并输出\n")
	fmt.Printf("  minify <json文件>       - 最小化JSON文件并输出\n")
	fmt.Printf("  stats <json文件>        - 显示JSON文件的统计信息\n")
	fmt.Printf("  find <json文件> <路径>  - 在JSON中查找指定路径的值\n")
	fmt.Printf("  compare <文件1> <文件2> - 比较两个JSON文件\n")
	fmt.Printf("\n详细帮助请使用: %s help <命令>\n", progName)
}

// DispatchCommand 分发命令到相应的处理函数
func DispatchCommand(progName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("未指定命令")
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "help":
		if len(cmdArgs) == 0 {
			PrintUsage(progName)
			return nil
		}
		return showHelp(cmdArgs[0])
	case "parse":
		return handleParse(cmdArgs)
	case "format":
		return handleFormat(cmdArgs, true)
	case "minify":
		return handleFormat(cmdArgs, false)
	case "stats":
		return handleStats(cmdArgs)
	case "find":
		return handleFind(cmdArgs)
	case "compare":
		return handleCompare(cmdArgs)
	default:
		return fmt.Errorf("未知命令: %s", cmd)
	}
}

// showHelp 显示特定命令的帮助信息
func showHelp(cmd string) error {
	switch cmd {
	case "parse":
		fmt.Println("parse 命令 - 解析JSON文件并验证其有效性")
		fmt.Println("用法: parse <json文件>")
		fmt.Println("描述: 解析指定的JSON文件并验证其语法是否正确。")
		fmt.Println("选项:")
		fmt.Println("  --concurrent  使用并发解析器 (实验性功能)")
	case "format":
		fmt.Println("format 命令 - 格式化JSON文件并输出")
		fmt.Println("用法: format <json文件> [--indent=<缩进字符串>]")
		fmt.Println("描述: 读取JSON文件并以格式化方式输出，默认使用2个空格缩进。")
		fmt.Println("选项:")
		fmt.Println("  --indent=<缩进字符串>  指定缩进字符串 (默认: '  ')")
	case "minify":
		fmt.Println("minify 命令 - 最小化JSON文件并输出")
		fmt.Println("用法: minify <json文件>")
		fmt.Println("描述: 将JSON文件压缩为紧凑格式，移除所有不必要的空白字符。")
	case "stats":
		fmt.Println("stats 命令 - 显示JSON文件的统计信息")
		fmt.Println("用法: stats <json文件>")
		fmt.Println("描述: 分析JSON文件并显示统计信息，如元素数量、值类型分布等。")
	case "find":
		fmt.Println("find 命令 - 在JSON中查找指定路径的值")
		fmt.Println("用法: find <json文件> <路径>")
		fmt.Println("描述: 在JSON文件中查找指定路径的值并显示。")
		fmt.Println("路径格式: 使用点号分隔对象成员，方括号指定数组索引。")
		fmt.Println("示例: user.address.city 或 users[0].name")
	case "compare":
		fmt.Println("compare 命令 - 比较两个JSON文件")
		fmt.Println("用法: compare <文件1> <文件2>")
		fmt.Println("描述: 比较两个JSON文件，检查它们在语义上是否相等。")
	default:
		return fmt.Errorf("未知命令: %s", cmd)
	}
	return nil
}

// handleParse 处理parse命令
func handleParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("parse命令需要一个文件名参数")
	}

	jsonFile := args[0]
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	useConcurrent := false
	for _, arg := range args[1:] {
		if arg == "--concurrent" {
			useConcurrent = true
			break
		}
	}

	if useConcurrent {
		// 注意：这里我们简化了处理，因为ParseConcurrent可能没有实现
		opts := leptjson.ConcurrentParseOptions{}
		_, err := ParseConcurrent(string(data), opts)
		if err != nil {
			return fmt.Errorf("JSON解析失败: %v", err)
		}
	} else {
		// 使用标准解析器
		v := &leptjson.Value{}
		if err := Parse(v, string(data)); err != nil {
			return fmt.Errorf("JSON解析失败: %v", err)
		}
	}

	fmt.Println("JSON语法有效")
	return nil
}

// handleFormat 处理format和minify命令
func handleFormat(args []string, prettyPrint bool) error {
	if len(args) < 1 {
		return fmt.Errorf("format命令需要一个文件名参数")
	}

	jsonFile := args[0]
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	v := &leptjson.Value{}
	if err := Parse(v, string(data)); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	var formatted string
	if prettyPrint {
		// 查找自定义缩进
		indent := "  " // 默认两个空格
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "--indent=") {
				indent = strings.TrimPrefix(arg, "--indent=")
				break
			}
		}
		formatted, err = StringifyIndent(v, indent)
	} else {
		formatted, err = Stringify(v)
	}

	if err != nil {
		return fmt.Errorf("格式化JSON失败: %v", err)
	}

	fmt.Println(formatted)
	return nil
}

// handleStats 处理stats命令
func handleStats(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("stats命令需要一个文件名参数")
	}

	jsonFile := args[0]
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	v := &leptjson.Value{}
	if err := Parse(v, string(data)); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	stats := CalculateStats(v)
	fmt.Println("JSON统计信息:")
	fmt.Printf("总元素数量: %d\n", stats.TotalElements)
	fmt.Printf("最大嵌套深度: %d\n", stats.MaxDepth)
	fmt.Printf("null值数量: %d\n", stats.NullCount)
	fmt.Printf("布尔值数量: %d (true: %d, false: %d)\n",
		stats.BoolCount, stats.TrueCount, stats.FalseCount)
	fmt.Printf("数字数量: %d (总和: %f)\n", stats.NumberCount, stats.NumberSum)
	fmt.Printf("字符串数量: %d (总长度: %d)\n", stats.StringCount, stats.StringLength)
	fmt.Printf("数组数量: %d (元素总数: %d)\n", stats.ArrayCount, stats.ArrayElements)
	fmt.Printf("对象数量: %d (字段总数: %d)\n", stats.ObjectCount, stats.TotalFields)

	return nil
}

// 通过路径查找JSON值
func findByPath(v *leptjson.Value, path string) (*leptjson.Value, error) {
	if v == nil {
		return nil, fmt.Errorf("空JSON值")
	}

	if path == "" {
		return v, nil
	}

	// 如果路径是直接索引数组
	if strings.HasPrefix(path, "[") {
		endBracket := strings.Index(path, "]")
		if endBracket == -1 {
			return nil, fmt.Errorf("无效的数组索引语法")
		}

		indexStr := path[1:endBracket]
		var index int
		_, err := fmt.Sscanf(indexStr, "%d", &index)
		if err != nil {
			return nil, fmt.Errorf("无效的数组索引: %s", indexStr)
		}

		if v.Type != JSON_ARRAY {
			return nil, fmt.Errorf("当前值不是数组")
		}

		if index < 0 || index >= len(v.A) {
			return nil, fmt.Errorf("数组索引越界: %d", index)
		}

		// 递归处理剩余路径
		remainingPath := ""
		if endBracket+1 < len(path) {
			if path[endBracket+1] == '.' {
				remainingPath = path[endBracket+2:]
			} else if path[endBracket+1] == '[' {
				remainingPath = path[endBracket+1:]
			} else {
				return nil, fmt.Errorf("无效的路径语法")
			}
		}

		if remainingPath == "" {
			return v.A[index], nil
		}
		return findByPath(v.A[index], remainingPath)
	}

	// 处理对象成员路径
	dotIndex := strings.Index(path, ".")
	bracketIndex := strings.Index(path, "[")

	var key string
	var remainingPath string

	if dotIndex != -1 && (bracketIndex == -1 || dotIndex < bracketIndex) {
		key = path[:dotIndex]
		remainingPath = path[dotIndex+1:]
	} else if bracketIndex != -1 {
		key = path[:bracketIndex]
		remainingPath = path[bracketIndex:]
	} else {
		key = path
		remainingPath = ""
	}

	if v.Type != JSON_OBJECT {
		return nil, fmt.Errorf("当前值不是对象")
	}

	found, exists := FindObjectValue(v, key)
	if !exists {
		return nil, fmt.Errorf("找不到键: %s", key)
	}

	if remainingPath == "" {
		return found, nil
	}
	return findByPath(found, remainingPath)
}

// handleFind 处理find命令
func handleFind(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("find命令需要文件名和路径两个参数")
	}

	jsonFile := args[0]
	path := args[1]

	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	v := &leptjson.Value{}
	if err := Parse(v, string(data)); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	result, err := findByPath(v, path)
	if err != nil {
		return fmt.Errorf("查找路径失败: %v", err)
	}

	// 格式化并输出结果
	formatted, err := StringifyIndent(result, "  ")
	if err != nil {
		return fmt.Errorf("格式化结果失败: %v", err)
	}

	fmt.Println(formatted)
	return nil
}

// handleCompare 处理compare命令
func handleCompare(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("compare命令需要两个文件名参数")
	}

	jsonFile1 := args[0]
	jsonFile2 := args[1]

	data1, err := ioutil.ReadFile(jsonFile1)
	if err != nil {
		return fmt.Errorf("读取第一个文件失败: %v", err)
	}

	data2, err := ioutil.ReadFile(jsonFile2)
	if err != nil {
		return fmt.Errorf("读取第二个文件失败: %v", err)
	}

	// 使用标准库比较标准化后的JSON
	var json1, json2 interface{}
	if err := json.Unmarshal(data1, &json1); err != nil {
		return fmt.Errorf("解析第一个文件失败: %v", err)
	}
	if err := json.Unmarshal(data2, &json2); err != nil {
		return fmt.Errorf("解析第二个文件失败: %v", err)
	}

	// 使用IsEqual辅助函数比较
	if isEqual(json1, json2) {
		fmt.Println("两个JSON文件在语义上相等")
	} else {
		fmt.Println("两个JSON文件不相等")
	}

	return nil
}

// isEqual 比较两个Go类型是否相等
func isEqual(a, b interface{}) bool {
	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for k, v := range aVal {
			if bv, ok := bVal[k]; !ok || !isEqual(v, bv) {
				return false
			}
		}
		return true
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for i, v := range aVal {
			if !isEqual(v, bVal[i]) {
				return false
			}
		}
		return true
	case string:
		bVal, ok := b.(string)
		return ok && aVal == bVal
	case float64:
		bVal, ok := b.(float64)
		return ok && aVal == bVal
	case bool:
		bVal, ok := b.(bool)
		return ok && aVal == bVal
	case nil:
		return b == nil
	default:
		return a == b
	}
}
