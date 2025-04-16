package leptjson

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// 命令行工具的版本号
const Version = "1.0.0"

// 统计信息结构
type JSONStats struct {
	TotalSize    int64  // 文件总大小（字节）
	ObjectCount  int    // 对象数量
	ArrayCount   int    // 数组数量
	StringCount  int    // 字符串数量
	NumberCount  int    // 数字数量
	BooleanCount int    // 布尔值数量
	NullCount    int    // null值数量
	MaxDepth     int    // 最大嵌套深度
	KeyCount     int    // 键的总数
	MaxKeyLength int    // 最长键的长度
	LongestKey   string // 最长的键
}

// 计算JSON的统计信息
func calculateStats(v *Value) JSONStats {
	stats := JSONStats{}
	calculateStatsRecursive(v, &stats, 0)
	return stats
}

// 递归计算JSON统计信息
func calculateStatsRecursive(v *Value, stats *JSONStats, depth int) {
	if v == nil {
		return
	}

	// 更新最大深度
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	// 根据类型更新计数
	switch v.Type {
	case NULL:
		stats.NullCount++
	case TRUE, FALSE:
		stats.BooleanCount++
	case NUMBER:
		stats.NumberCount++
	case STRING:
		stats.StringCount++
	case ARRAY:
		stats.ArrayCount++
		for _, elem := range v.A {
			calculateStatsRecursive(elem, stats, depth+1)
		}
	case OBJECT:
		stats.ObjectCount++
		for _, member := range v.O {
			stats.KeyCount++

			// 更新最长键信息
			if len(member.K) > stats.MaxKeyLength {
				stats.MaxKeyLength = len(member.K)
				stats.LongestKey = member.K
			}

			calculateStatsRecursive(member.V, stats, depth+1)
		}
	}
}

// 格式化输出带有缩进的JSON
func formatJSON(v *Value, indent string) (string, error) {
	// 创建格式化选项
	var result strings.Builder
	formatJSONRecursive(&result, v, 0, indent)
	return result.String(), nil
}

// 递归格式化JSON
func formatJSONRecursive(out *strings.Builder, v *Value, level int, indent string) {
	if v == nil {
		out.WriteString("null")
		return
	}

	switch v.Type {
	case NULL:
		out.WriteString("null")
	case TRUE:
		out.WriteString("true")
	case FALSE:
		out.WriteString("false")
	case NUMBER:
		out.WriteString(fmt.Sprintf("%g", v.N))
	case STRING:
		out.WriteString(formatJSONString(v.S))
	case ARRAY:
		if len(v.A) == 0 {
			out.WriteString("[]")
			return
		}

		out.WriteString("[\n")
		for i, elem := range v.A {
			out.WriteString(strings.Repeat(indent, level+1))
			formatJSONRecursive(out, elem, level+1, indent)
			if i < len(v.A)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		out.WriteString(strings.Repeat(indent, level))
		out.WriteString("]")
	case OBJECT:
		if len(v.O) == 0 {
			out.WriteString("{}")
			return
		}

		out.WriteString("{\n")
		for i, member := range v.O {
			out.WriteString(strings.Repeat(indent, level+1))
			out.WriteString(formatJSONString(member.K))
			out.WriteString(": ")
			formatJSONRecursive(out, member.V, level+1, indent)
			if i < len(v.O)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		out.WriteString(strings.Repeat(indent, level))
		out.WriteString("}")
	}
}

// 格式化JSON字符串（添加引号和转义）
func formatJSONString(s string) string {
	result := strings.Builder{}
	result.WriteString("\"")

	for _, c := range s {
		switch c {
		case '"':
			result.WriteString("\\\"")
		case '\\':
			result.WriteString("\\\\")
		case '\b':
			result.WriteString("\\b")
		case '\f':
			result.WriteString("\\f")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			if c < 32 {
				result.WriteString(fmt.Sprintf("\\u%04x", c))
			} else {
				result.WriteRune(c)
			}
		}
	}

	result.WriteString("\"")
	return result.String()
}

// 最小化JSON（去除所有多余空格）
func minifyJSON(v *Value) (string, error) {
	jsonStr, errCode := Stringify(v)
	if errCode != STRINGIFY_OK {
		return "", fmt.Errorf("最小化JSON失败: %s", errCode)
	}
	return jsonStr, nil
}

// 比较两个JSON文档，返回差异
func compareJSON(v1, v2 *Value) []string {
	differences := []string{}
	compareJSONRecursive(v1, v2, "$", &differences)
	return differences
}

// 递归比较JSON文档
func compareJSONRecursive(v1, v2 *Value, path string, differences *[]string) {
	// 检查类型是否相同
	if v1 == nil || v2 == nil {
		if (v1 == nil) != (v2 == nil) {
			*differences = append(*differences, fmt.Sprintf("路径 %s: 一个为null，另一个不是", path))
		}
		return
	}

	if v1.Type != v2.Type {
		*differences = append(*differences, fmt.Sprintf("路径 %s: 类型不匹配 (%s vs %s)", path, getValueTypeName(v1.Type), getValueTypeName(v2.Type)))
		return
	}

	// 根据类型进一步比较
	switch v1.Type {
	case NULL:
		// null相等，不需要进一步比较
	case TRUE, FALSE:
		isV1True := v1.Type == TRUE
		isV2True := v2.Type == TRUE
		if isV1True != isV2True {
			*differences = append(*differences, fmt.Sprintf("路径 %s: 布尔值不同 (%t vs %t)", path, isV1True, isV2True))
		}
	case NUMBER:
		if v1.N != v2.N {
			*differences = append(*differences, fmt.Sprintf("路径 %s: 数字不同 (%g vs %g)", path, v1.N, v2.N))
		}
	case STRING:
		if v1.S != v2.S {
			if len(v1.S) > 50 || len(v2.S) > 50 {
				*differences = append(*differences, fmt.Sprintf("路径 %s: 字符串不同 (长度: %d vs %d)", path, len(v1.S), len(v2.S)))
			} else {
				*differences = append(*differences, fmt.Sprintf("路径 %s: 字符串不同 (\"%s\" vs \"%s\")", path, v1.S, v2.S))
			}
		}
	case ARRAY:
		// 检查数组长度
		if len(v1.A) != len(v2.A) {
			*differences = append(*differences, fmt.Sprintf("路径 %s: 数组长度不同 (%d vs %d)", path, len(v1.A), len(v2.A)))
		}

		// 比较数组元素
		minLen := len(v1.A)
		if len(v2.A) < minLen {
			minLen = len(v2.A)
		}

		for i := 0; i < minLen; i++ {
			compareJSONRecursive(v1.A[i], v2.A[i], fmt.Sprintf("%s[%d]", path, i), differences)
		}
	case OBJECT:
		// 创建v2的键映射，用于快速查找
		v2Keys := make(map[string]*Value)
		for _, member := range v2.O {
			v2Keys[member.K] = member.V
		}

		// 检查v1中的每个键
		for _, member := range v1.O {
			v2Value, exists := v2Keys[member.K]
			if !exists {
				*differences = append(*differences, fmt.Sprintf("路径 %s: 第一个JSON有键 '%s'，但第二个没有", path, member.K))
				continue
			}

			// 递归比较值
			compareJSONRecursive(member.V, v2Value, fmt.Sprintf("%s.%s", path, member.K), differences)

			// 从v2Keys中删除已比较的键
			delete(v2Keys, member.K)
		}

		// 检查v2中的剩余键（v1中不存在的键）
		for k := range v2Keys {
			*differences = append(*differences, fmt.Sprintf("路径 %s: 第二个JSON有键 '%s'，但第一个没有", path, k))
		}
	}
}

// 获取类型的字符串表示
func getValueTypeName(t ValueType) string {
	switch t {
	case NULL:
		return "null"
	case TRUE, FALSE:
		return "boolean"
	case NUMBER:
		return "number"
	case STRING:
		return "string"
	case ARRAY:
		return "array"
	case OBJECT:
		return "object"
	default:
		return "unknown"
	}
}

// RunCLI 运行CLI，处理命令行参数和子命令
func RunCLI() {
	// 检查参数数量
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// 检查是否使用简单的帮助选项
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		return
	}

	// 检查是否使用版本选项
	if os.Args[1] == "--version" {
		fmt.Printf("LeptJSON CLI 版本 %s\n", Version)
		return
	}

	// 创建主命令行解析器
	mainCmd := flag.NewFlagSet("leptjson", flag.ExitOnError)
	verbose := mainCmd.Bool("verbose", false, "显示详细输出")
	verboseShort := mainCmd.Bool("v", false, "显示详细输出")
	help := mainCmd.Bool("help", false, "显示帮助信息")
	helpShort := mainCmd.Bool("h", false, "显示帮助信息")
	version := mainCmd.Bool("version", false, "显示版本信息")

	// 解析全局选项
	mainCmd.Parse(os.Args[1:])

	// 处理全局标志
	if *version {
		fmt.Printf("LeptJSON CLI 版本 %s\n", Version)
		return
	}

	if *help || *helpShort {
		printUsage()
		return
	}

	// 获取子命令
	args := mainCmd.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	// 使用-v或--verbose都可以开启详细模式
	verboseMode := *verbose || *verboseShort

	subCommand := args[0]
	subArgs := args[1:]

	// 检查子命令是否是帮助请求
	if len(subArgs) > 0 && (subArgs[0] == "-h" || subArgs[0] == "--help") {
		printSubcommandHelp(subCommand)
		return
	}

	// 根据子命令执行对应的操作
	switch subCommand {
	case "parse":
		runParse(subArgs, verboseMode)
	case "format":
		runFormat(subArgs, verboseMode)
	case "minify":
		runMinify(subArgs, verboseMode)
	case "stats":
		runStats(subArgs, verboseMode)
	case "find":
		runFind(subArgs, verboseMode)
	case "path":
		runPath(subArgs, verboseMode)
	case "compare":
		runCompare(subArgs, verboseMode)
	case "validate":
		runValidate(subArgs, verboseMode)
	case "pointer":
		runPointer(subArgs, verboseMode)
	case "patch":
		runPatch(subArgs, verboseMode)
	case "merge-patch":
		runMergePatch(subArgs, verboseMode)
	default:
		fmt.Printf("未知的命令: %s\n", subCommand)
		printUsage()
	}
}

// 打印子命令的帮助信息
func printSubcommandHelp(command string) {
	switch command {
	case "parse":
		fmt.Println("leptjson parse - 解析并验证JSON文件")
		fmt.Println("\n用法: leptjson parse FILE")
		fmt.Println("\n参数:")
		fmt.Println("  FILE    要解析的JSON文件路径")

	case "format":
		fmt.Println("leptjson format - 格式化JSON文件")
		fmt.Println("\n用法: leptjson format [选项] FILE [OUTPUT]")
		fmt.Println("\n选项:")
		fmt.Println("  --indent=N    设置缩进空格数（默认为2）")
		fmt.Println("\n参数:")
		fmt.Println("  FILE          要格式化的JSON文件路径")
		fmt.Println("  OUTPUT        输出文件路径（可选，默认为FILE.formatted.json）")

	case "minify":
		fmt.Println("leptjson minify - 最小化JSON文件")
		fmt.Println("\n用法: leptjson minify FILE [OUTPUT]")
		fmt.Println("\n参数:")
		fmt.Println("  FILE          要最小化的JSON文件路径")
		fmt.Println("  OUTPUT        输出文件路径（可选，默认为FILE.min.json）")

	case "stats":
		fmt.Println("leptjson stats - 显示JSON统计信息")
		fmt.Println("\n用法: leptjson stats [选项] FILE")
		fmt.Println("\n选项:")
		fmt.Println("  --json        以JSON格式输出统计信息")
		fmt.Println("\n参数:")
		fmt.Println("  FILE          要分析的JSON文件路径")

	case "find":
		fmt.Println("leptjson find - 在JSON中查找特定路径的值")
		fmt.Println("\n用法: leptjson find [选项] FILE JSONPATH")
		fmt.Println("\n选项:")
		fmt.Println("  --output=FORMAT    设置输出格式，可选值: compact, pretty, raw（默认为compact）")
		fmt.Println("\n参数:")
		fmt.Println("  FILE          要搜索的JSON文件路径")
		fmt.Println("  JSONPATH      JSONPath表达式，如$.store.book[0].title")

	case "compare":
		fmt.Println("leptjson compare - 比较两个JSON文件")
		fmt.Println("\n用法: leptjson compare [选项] FILE1 FILE2")
		fmt.Println("\n选项:")
		fmt.Println("  --json        以JSON格式输出差异")
		fmt.Println("\n参数:")
		fmt.Println("  FILE1         第一个JSON文件路径")
		fmt.Println("  FILE2         第二个JSON文件路径")

	case "validate":
		fmt.Println("leptjson validate - 使用JSON Schema验证JSON文件")
		fmt.Println("\n用法: leptjson validate [选项] SCHEMA FILE")
		fmt.Println("\n选项:")
		fmt.Println("  --format=FORMAT    设置输出格式，可选值: text, json（默认为text）")
		fmt.Println("\n参数:")
		fmt.Println("  SCHEMA             JSON Schema文件路径")
		fmt.Println("  FILE               要验证的JSON文件路径")
		fmt.Println("\n说明:")
		fmt.Println("  该命令使用JSON Schema验证JSON文件的结构和内容。")
		fmt.Println("  验证失败时会显示详细的错误信息。")
		fmt.Println("  支持Draft-07版本的JSON Schema规范的主要功能。")

	case "pointer":
		fmt.Println("leptjson pointer - 使用JSON Pointer操作JSON文件")
		fmt.Println("\n用法: leptjson pointer [选项] FILE POINTER")
		fmt.Println("\n选项:")
		fmt.Println("  --operation=OP    操作类型，可选值:")
		fmt.Println("                      - get: 获取值（默认）")
		fmt.Println("                      - add: 添加或替换值")
		fmt.Println("                      - remove: 删除值")
		fmt.Println("                      - replace: 替换值")
		fmt.Println("  --value=JSON      用于add和replace操作的JSON值")
		fmt.Println("  --output=FILE     保存修改后的JSON到指定文件")
		fmt.Println("\n参数:")
		fmt.Println("  FILE              要操作的JSON文件路径")
		fmt.Println("  POINTER           JSON Pointer路径，如/users/0/name")
		fmt.Println("\n说明:")
		fmt.Println("  该命令实现了RFC 6901中定义的JSON Pointer，用于在JSON文档中定位和操作值。")
		fmt.Println("  JSON Pointer以/开头，使用/分隔路径片段，如/foo/0/bar引用{\"foo\":[{\"bar\":42}]}中的42。")
		fmt.Println("  ~0表示~，~1表示/。")

	case "patch":
		fmt.Println("leptjson patch - 使用JSON Patch修改JSON文件")
		fmt.Println("\n用法: leptjson patch [选项] PATCH FILE [OUTPUT]")
		fmt.Println("\n选项:")
		fmt.Println("  --in-place         直接修改原文件，不创建新文件")
		fmt.Println("  --test             仅测试补丁，不实际修改文件")
		fmt.Println("\n参数:")
		fmt.Println("  PATCH              包含JSON Patch操作的文件")
		fmt.Println("  FILE               要修改的JSON文件")
		fmt.Println("  OUTPUT             输出文件路径（可选）")
		fmt.Println("\n说明:")
		fmt.Println("  该命令实现了RFC 6902中定义的JSON Patch，用于修改JSON文档。")
		fmt.Println("  JSON Patch是一组操作指令，如add、remove、replace、move、copy和test。")
		fmt.Println("  每个操作都有一个'op'字段指定操作类型，以及一个'path'字段指定操作位置。")

	case "merge-patch":
		fmt.Println("leptjson merge-patch - 使用JSON Merge Patch合并JSON文件")
		fmt.Println("\n用法: leptjson merge-patch [选项] PATCH FILE [OUTPUT]")
		fmt.Println("\n选项:")
		fmt.Println("  --in-place         直接修改原文件，不创建新文件")
		fmt.Println("\n参数:")
		fmt.Println("  PATCH              包含Merge Patch操作的JSON文件")
		fmt.Println("  FILE               要修改的目标JSON文件")
		fmt.Println("  OUTPUT             输出文件路径（可选，默认为FILE.merged.json）")
		fmt.Println("\n说明:")
		fmt.Println("  该命令实现了RFC 7396中定义的JSON Merge Patch，用于简化JSON文档的合并。")
		fmt.Println("  与JSON Patch不同，JSON Merge Patch本身就是一个JSON对象，结构与目标文档类似。")
		fmt.Println("  合并规则:")
		fmt.Println("    - 如果补丁中的值为null，则从目标中删除该字段")
		fmt.Println("    - 如果补丁中包含非null值，则替换目标中的相应值")
		fmt.Println("    - 如果两边都是对象，则递归合并")
		fmt.Println("    - 如果补丁中的值是数组，则完全替换目标中的数组")

	case "path":
		fmt.Println("leptjson path - 使用JSONPath查询JSON文件")
		fmt.Println("\n用法: leptjson path [选项] FILE JSONPATH")
		fmt.Println("\n选项:")
		fmt.Println("  --output=FORMAT    设置输出格式，可选值: compact, pretty, raw, table（默认为pretty）")
		fmt.Println("  --all              显示所有匹配的结果（默认只显示前10个）")
		fmt.Println("  --csv=FILE         将结果输出为CSV文件")
		fmt.Println("  --no-path          不在输出中显示路径信息")
		fmt.Println("\n参数:")
		fmt.Println("  FILE               要查询的JSON文件路径")
		fmt.Println("  JSONPATH           JSONPath表达式，如$.store.book[*].author")
		fmt.Println("\n说明:")
		fmt.Println("  该命令使用JSONPath表达式从JSON文件中提取数据。")
		fmt.Println("  JSONPath是一种用于从JSON文档中选择和提取数据的查询语言。")
		fmt.Println("\n支持的JSONPath语法:")
		fmt.Println("  $                  根对象或数组")
		fmt.Println("  .property          子属性")
		fmt.Println("  ['property']       子属性（带引号）")
		fmt.Println("  [index]            数组索引")
		fmt.Println("  [start:end:step]   数组切片")
		fmt.Println("  *                  通配符，匹配所有属性或元素")
		fmt.Println("  ..property         递归下降，匹配任意深度的属性")
		fmt.Println("  [?(@.prop > 10)]   过滤表达式")
		fmt.Println("  [?(@.prop)]        存在性检查")
		fmt.Println("  [?(@.name == 'x')] 相等性检查")
		fmt.Println("  ['a','b']          多属性选择")

	default:
		fmt.Printf("未知的命令: %s\n", command)
		printUsage()
	}
}

// 打印用法信息
func printUsage() {
	fmt.Println("用法: leptjson [选项] 命令 [参数]")
	fmt.Println("\n全局选项:")
	fmt.Println("  --help, -h      显示帮助信息")
	fmt.Println("  --verbose, -v   显示详细输出")
	fmt.Println("  --version       显示版本信息")

	fmt.Println("\n可用命令:")
	fmt.Println("  parse           解析并验证JSON文件")
	fmt.Println("  format          格式化JSON文件")
	fmt.Println("  minify          最小化JSON文件")
	fmt.Println("  stats           显示JSON统计信息")
	fmt.Println("  find            在JSON中查找特定路径的值（简化版JSONPath）")
	fmt.Println("  path            使用完整JSONPath语法查询JSON数据")
	fmt.Println("  compare         比较两个JSON文件")
	fmt.Println("  validate        使用JSON Schema验证JSON文件")
	fmt.Println("  pointer         使用JSON Pointer操作JSON文件")
	fmt.Println("  patch           使用JSON Patch修改JSON文件")
	fmt.Println("  merge-patch     使用JSON Merge Patch合并JSON文件")

	fmt.Println("\n命令详情:")

	// parse命令
	fmt.Println("\n  parse FILE")
	fmt.Println("    解析并验证JSON文件的格式")
	fmt.Println("    参数:")
	fmt.Println("      FILE        要解析的JSON文件路径")

	// format命令
	fmt.Println("\n  format [选项] FILE [OUTPUT]")
	fmt.Println("    格式化JSON文件，增加缩进和换行")
	fmt.Println("    选项:")
	fmt.Println("      --indent=N  设置缩进空格数（默认为2）")
	fmt.Println("    参数:")
	fmt.Println("      FILE        要格式化的JSON文件路径")
	fmt.Println("      OUTPUT      输出文件路径（可选，默认为FILE.formatted.json）")

	// minify命令
	fmt.Println("\n  minify FILE [OUTPUT]")
	fmt.Println("    最小化JSON文件，移除所有不必要的空白字符")
	fmt.Println("    参数:")
	fmt.Println("      FILE        要最小化的JSON文件路径")
	fmt.Println("      OUTPUT      输出文件路径（可选，默认为FILE.min.json）")

	// stats命令
	fmt.Println("\n  stats [选项] FILE")
	fmt.Println("    分析JSON文件并显示统计信息")
	fmt.Println("    选项:")
	fmt.Println("      --json      以JSON格式输出统计信息")
	fmt.Println("    参数:")
	fmt.Println("      FILE        要分析的JSON文件路径")

	// find命令
	fmt.Println("\n  find [选项] FILE JSONPATH")
	fmt.Println("    使用JSONPath表达式在JSON文件中查找值")
	fmt.Println("    选项:")
	fmt.Println("      --output=FORMAT  设置输出格式，可选值: compact, pretty, raw（默认为compact）")
	fmt.Println("    参数:")
	fmt.Println("      FILE        要搜索的JSON文件路径")
	fmt.Println("      JSONPATH    JSONPath表达式，如$.store.book[0].title")

	// compare命令
	fmt.Println("\n  compare [选项] FILE1 FILE2")
	fmt.Println("    比较两个JSON文件并显示差异")
	fmt.Println("    选项:")
	fmt.Println("      --json      以JSON格式输出差异")
	fmt.Println("    参数:")
	fmt.Println("      FILE1       第一个JSON文件路径")
	fmt.Println("      FILE2       第二个JSON文件路径")

	// validate命令
	fmt.Println("\n  validate [选项] SCHEMA FILE")
	fmt.Println("    使用JSON Schema验证JSON文件")
	fmt.Println("    选项:")
	fmt.Println("      --format=FORMAT  设置输出格式，可选值: text, json（默认为text）")
	fmt.Println("    参数:")
	fmt.Println("      SCHEMA       JSON Schema文件路径")
	fmt.Println("      FILE         要验证的JSON文件路径")

	// pointer命令
	fmt.Println("\n  pointer [选项] FILE POINTER")
	fmt.Println("    使用JSON Pointer (RFC 6901)操作JSON文件")
	fmt.Println("    选项:")
	fmt.Println("      --operation=OP  操作类型：get(默认),add,remove,replace")
	fmt.Println("      --value=JSON    用于add和replace操作的JSON值")
	fmt.Println("      --output=FILE   保存修改后的JSON文件路径（默认覆盖原文件）")
	fmt.Println("    参数:")
	fmt.Println("      FILE         要操作的JSON文件路径")
	fmt.Println("      POINTER      JSON Pointer路径，如/users/0/name")

	// patch命令
	fmt.Println("\n  patch [选项] PATCH FILE [OUTPUT]")
	fmt.Println("    使用JSON Patch (RFC 6902)修改JSON文件")
	fmt.Println("    选项:")
	fmt.Println("      --in-place       直接修改原文件，不创建新文件")
	fmt.Println("      --test           仅测试补丁，不实际修改文件")
	fmt.Println("    参数:")
	fmt.Println("      PATCH        包含JSON Patch操作的文件")
	fmt.Println("      FILE         要修改的JSON文件")
	fmt.Println("      OUTPUT       输出文件路径（可选）")

	// merge-patch命令
	fmt.Println("\n  merge-patch [选项] PATCH FILE [OUTPUT]")
	fmt.Println("    使用JSON Merge Patch (RFC 7396)合并JSON文件")
	fmt.Println("    选项:")
	fmt.Println("      --in-place       直接修改原文件，不创建新文件")
	fmt.Println("    参数:")
	fmt.Println("      PATCH        包含Merge Patch操作的JSON文件")
	fmt.Println("      FILE         要修改的目标JSON文件")
	fmt.Println("      OUTPUT       输出文件路径（可选，默认为FILE.merged.json）")

	// path命令
	fmt.Println("\n  path [选项] FILE JSONPATH")
	fmt.Println("    使用完整的JSONPath语法查询JSON文件")
	fmt.Println("    选项:")
	fmt.Println("      --output=FORMAT  设置输出格式: compact, pretty, raw, table")
	fmt.Println("      --all            显示所有匹配结果(默认仅显示前10个)")
	fmt.Println("      --csv=FILE       将结果保存为CSV文件")
	fmt.Println("      --no-path        不在输出中显示路径信息")
	fmt.Println("    参数:")
	fmt.Println("      FILE           要查询的JSON文件路径")
	fmt.Println("      JSONPATH       JSONPath表达式，如$..book[?(@.price<10)]")

	fmt.Println("\n示例:")
	fmt.Println("  leptjson parse data.json")
	fmt.Println("  leptjson format --indent=2 data.json pretty.json")
	fmt.Println("  leptjson minify large.json small.json")
	fmt.Println("  leptjson stats --json data.json")
	fmt.Println("  leptjson find --output=pretty data.json \"$.store.book[0].title\"")
	fmt.Println("  leptjson path --output=table data.json \"$..book[?(@.price < 10)]\"")
	fmt.Println("  leptjson compare original.json updated.json")
	fmt.Println("  leptjson validate --format=json schema.json data.json")
	fmt.Println("  leptjson pointer data.json \"/users/0/name\"")
	fmt.Println("  leptjson pointer --operation=replace --value=\"John\" data.json \"/users/0/name\"")
	fmt.Println("  leptjson patch patch.json data.json result.json")
	fmt.Println("  leptjson merge-patch merge.json data.json result.json")

}

// 从文件加载JSON
func loadJSON(filename string, verbose bool) (*Value, error) {
	if verbose {
		fmt.Printf("正在读取文件: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析JSON
	var v Value
	parseErr := Parse(&v, string(data))
	if parseErr != PARSE_OK {
		return nil, fmt.Errorf("解析JSON失败: %s", parseErr)
	}

	return &v, nil
}

// 保存JSON到文件
func saveJSON(filename string, content string, verbose bool) error {
	if verbose {
		fmt.Printf("正在写入文件: %s\n", filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建文件: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// runParse 运行parse命令
func runParse(args []string, verbose bool) {
	if len(args) != 1 {
		fmt.Println("错误: parse命令需要一个文件参数")
		fmt.Println("\n用法: leptjson parse FILE")
		return
	}

	filePath := args[0]
	if verbose {
		fmt.Printf("正在解析文件: %s\n", filePath)
	}

	// 尝试解析JSON文件
	_, err := loadJSON(filePath, verbose)
	if err != nil {
		fmt.Printf("解析失败: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("文件格式有效")
}

// runFormat 运行format命令
func runFormat(args []string, verbose bool) {
	if len(args) < 1 {
		fmt.Println("错误: format命令需要至少一个文件参数")
		fmt.Println("\n用法: leptjson format [选项] FILE [OUTPUT]")
		return
	}

	// 解析indent选项
	indentSpaces := 2
	fileArgs := args

	for i, arg := range args {
		if strings.HasPrefix(arg, "--indent=") {
			indentVal := strings.TrimPrefix(arg, "--indent=")
			spaces, err := strconv.Atoi(indentVal)
			if err != nil || spaces < 0 {
				fmt.Printf("错误: 无效的缩进值: %s\n", indentVal)
				return
			}
			indentSpaces = spaces
			// 从参数列表中移除选项
			fileArgs = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(fileArgs) < 1 || len(fileArgs) > 2 {
		fmt.Println("错误: format命令需要1-2个文件参数")
		fmt.Println("\n用法: leptjson format [--indent=SPACES] FILE [OUTPUT]")
		return
	}

	inputFile := fileArgs[0]
	outputFile := ""
	if len(fileArgs) == 2 {
		outputFile = fileArgs[1]
	} else {
		// 默认输出文件名
		outputFile = inputFile + ".formatted.json"
	}

	if verbose {
		fmt.Printf("正在格式化: %s -> %s (缩进: %d空格)\n", inputFile, outputFile, indentSpaces)
	}

	// 加载JSON
	v, err := loadJSON(inputFile, verbose)
	if err != nil {
		fmt.Printf("格式化失败: %s\n", err)
		os.Exit(1)
	}

	// 生成缩进字符串
	indent := strings.Repeat(" ", indentSpaces)

	// 格式化JSON
	formatted, err := formatJSON(v, indent)
	if err != nil {
		fmt.Printf("格式化失败: %s\n", err)
		os.Exit(1)
	}

	// 保存结果
	err = saveJSON(outputFile, formatted, verbose)
	if err != nil {
		fmt.Printf("保存结果失败: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("格式化完成: %s\n", outputFile)
}

// runMinify 运行minify命令
func runMinify(args []string, verbose bool) {
	if len(args) < 1 || len(args) > 2 {
		fmt.Println("错误: minify命令需要1-2个文件参数")
		fmt.Println("\n用法: leptjson minify FILE [OUTPUT]")
		return
	}

	inputFile := args[0]
	outputFile := ""
	if len(args) == 2 {
		outputFile = args[1]
	} else {
		// 默认输出文件名
		outputFile = inputFile + ".min.json"
	}

	if verbose {
		fmt.Printf("正在最小化: %s -> %s\n", inputFile, outputFile)
	}

	// 加载JSON
	v, err := loadJSON(inputFile, verbose)
	if err != nil {
		fmt.Printf("最小化失败: %s\n", err)
		os.Exit(1)
	}

	// 最小化JSON
	minified, err := minifyJSON(v)
	if err != nil {
		fmt.Printf("最小化失败: %s\n", err)
		os.Exit(1)
	}

	// 保存结果
	err = saveJSON(outputFile, minified, verbose)
	if err != nil {
		fmt.Printf("保存结果失败: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("最小化完成: %s\n", outputFile)
}

// runStats 运行stats命令
func runStats(args []string, verbose bool) {
	// 解析选项和参数
	jsonOutput := false
	fileArgs := args

	for i, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			// 从参数列表中移除选项
			fileArgs = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(fileArgs) != 1 {
		fmt.Println("错误: stats命令需要一个文件参数")
		fmt.Println("\n用法: leptjson stats [--json] FILE")
		return
	}

	filePath := fileArgs[0]
	if verbose {
		fmt.Printf("正在分析文件: %s\n", filePath)
	}

	// 加载JSON
	v, err := loadJSON(filePath, verbose)
	if err != nil {
		fmt.Printf("分析失败: %s\n", err)
		os.Exit(1)
	}

	// 计算统计信息
	stats := calculateStats(v)

	// 输出统计信息
	if jsonOutput {
		// 以JSON格式输出
		statsJSON, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			fmt.Printf("生成JSON统计信息失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(statsJSON))
	} else {
		// 以可读格式输出
		fmt.Printf("JSON统计信息 - %s\n", filePath)
		fmt.Printf("对象数量: %d\n", stats.ObjectCount)
		fmt.Printf("数组数量: %d\n", stats.ArrayCount)
		fmt.Printf("字符串数量: %d\n", stats.StringCount)
		fmt.Printf("数字数量: %d\n", stats.NumberCount)
		fmt.Printf("布尔值数量: %d\n", stats.BooleanCount)
		fmt.Printf("null值数量: %d\n", stats.NullCount)
		fmt.Printf("总键数量: %d\n", stats.KeyCount)
		fmt.Printf("最大深度: %d\n", stats.MaxDepth)
		if stats.MaxKeyLength > 0 {
			fmt.Printf("最长键: '%s' (%d字符)\n", stats.LongestKey, stats.MaxKeyLength)
		}
	}
}

// runFind 运行find命令
func runFind(args []string, verbose bool) {
	// 解析选项和参数
	outputFormat := "compact"
	fileArgs := args

	for i, arg := range args {
		if strings.HasPrefix(arg, "--output=") {
			outputFormat = strings.TrimPrefix(arg, "--output=")
			// 检查输出格式是否有效
			if outputFormat != "compact" && outputFormat != "pretty" && outputFormat != "raw" {
				fmt.Printf("错误: 无效的输出格式: %s\n", outputFormat)
				fmt.Println("有效的格式: compact, pretty, raw")
				return
			}
			// 从参数列表中移除选项
			fileArgs = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(fileArgs) != 2 {
		fmt.Println("错误: find命令需要两个参数")
		fmt.Println("\n用法: leptjson find [--output=FORMAT] FILE JSONPATH")
		return
	}

	filePath := fileArgs[0]
	jsonPath := fileArgs[1]

	if verbose {
		fmt.Printf("正在在文件 %s 中查找路径: %s\n", filePath, jsonPath)
	}

	// 加载JSON
	v, err := loadJSON(filePath, verbose)
	if err != nil {
		fmt.Printf("查找失败: %s\n", err)
		os.Exit(1)
	}

	// 执行JSONPath搜索
	result, err := findByPath(v, jsonPath)
	if err != nil {
		fmt.Printf("查找失败: %s\n", err)
		os.Exit(1)
	}

	// 输出结果
	switch outputFormat {
	case "compact":
		// 紧凑输出
		output, err := minifyJSON(result)
		if err != nil {
			fmt.Printf("生成输出失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	case "pretty":
		// 美化输出
		output, err := formatJSON(result, "  ")
		if err != nil {
			fmt.Printf("生成输出失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(output)
	case "raw":
		// 原始值输出
		switch result.Type {
		case STRING:
			fmt.Println(result.S)
		case NUMBER:
			fmt.Println(result.N)
		case TRUE:
			fmt.Println("true")
		case FALSE:
			fmt.Println("false")
		case NULL:
			fmt.Println("null")
		default:
			// 对象和数组默认使用格式化的输出
			output, err := formatJSON(result, "  ")
			if err != nil {
				fmt.Printf("生成输出失败: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(output)
		}
	}
}

// runCompare 运行compare命令
func runCompare(args []string, verbose bool) {
	// 解析选项和参数
	jsonOutput := false
	fileArgs := args

	for i, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			// 从参数列表中移除选项
			fileArgs = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(fileArgs) != 2 {
		fmt.Println("错误: compare命令需要两个文件参数")
		fmt.Println("\n用法: leptjson compare [--json] FILE1 FILE2")
		return
	}

	file1 := fileArgs[0]
	file2 := fileArgs[1]

	if verbose {
		fmt.Printf("正在比较文件: %s 和 %s\n", file1, file2)
	}

	// 加载两个JSON文件
	v1, err := loadJSON(file1, verbose)
	if err != nil {
		fmt.Printf("加载第一个文件失败: %s\n", err)
		os.Exit(1)
	}

	v2, err := loadJSON(file2, verbose)
	if err != nil {
		fmt.Printf("加载第二个文件失败: %s\n", err)
		os.Exit(1)
	}

	// 比较JSON
	differences := compareJSON(v1, v2)

	// 输出差异
	if len(differences) == 0 {
		fmt.Println("文件相同")
		return
	}

	if jsonOutput {
		// 以JSON格式输出差异
		diffJSON, err := json.MarshalIndent(differences, "", "  ")
		if err != nil {
			fmt.Printf("生成JSON差异报告失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(diffJSON))
	} else {
		// 以可读格式输出差异
		fmt.Printf("发现 %d 处差异:\n", len(differences))
		for i, diff := range differences {
			fmt.Printf("%d. %s\n", i+1, diff)
		}
	}
}

// findByPath 使用JSONPath查找值
func findByPath(v *Value, path string) (*Value, error) {
	// 这里实现JSONPath查询
	// 简化版本仅支持基本路径，如 $.store.book[0].title
	if !strings.HasPrefix(path, "$") {
		return nil, fmt.Errorf("JSONPath 必须以 $ 开头")
	}

	current := v
	segments := strings.Split(path[1:], ".")

	for _, segment := range segments {
		// 处理数组访问，如 book[0]
		indexStart := strings.Index(segment, "[")
		if indexStart > 0 {
			propName := segment[:indexStart]
			indexEnd := strings.Index(segment, "]")
			if indexEnd <= indexStart {
				return nil, fmt.Errorf("无效的数组索引语法: %s", segment)
			}

			// 获取属性
			if current.Type != OBJECT {
				return nil, fmt.Errorf("路径 %s 处预期对象类型", propName)
			}

			var found bool
			for _, m := range current.O {
				if m.K == propName {
					current = m.V
					found = true
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("未找到属性: %s", propName)
			}

			// 获取数组索引
			indexStr := segment[indexStart+1 : indexEnd]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("无效的数组索引: %s", indexStr)
			}

			if current.Type != ARRAY {
				return nil, fmt.Errorf("路径 %s 处预期数组类型", propName)
			}

			if index < 0 || index >= len(current.A) {
				return nil, fmt.Errorf("数组索引超出范围: %d", index)
			}

			current = current.A[index]
		} else if segment == "" {
			// 处理空段，如 $.
			continue
		} else {
			// 处理普通属性访问
			if current.Type != OBJECT {
				return nil, fmt.Errorf("路径 %s 处预期对象类型", segment)
			}

			var found bool
			for _, m := range current.O {
				if m.K == segment {
					current = m.V
					found = true
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("未找到属性: %s", segment)
			}
		}
	}

	return current, nil
}

// 验证结果结构
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

// JSON Schema验证的实现函数
func validateWithSchema(schema, data *Value) ValidationResult {
	// 创建验证结果
	result := ValidationResult{
		Valid: true,
	}

	// 调用JSON Schema验证函数
	schemaErrs := validateJSONSchema(schema, data, "$")
	if len(schemaErrs) > 0 {
		result.Valid = false
		result.Errors = schemaErrs
		if len(schemaErrs) == 1 {
			result.Message = "发现1个验证错误"
		} else {
			result.Message = fmt.Sprintf("发现%d个验证错误", len(schemaErrs))
		}
	}

	return result
}

// 实际的JSON Schema验证逻辑
func validateJSONSchema(schema, data *Value, path string) []string {
	errors := []string{}

	// 检查类型验证
	if typeSchema := findObjectKey(schema, "type"); typeSchema != nil {
		typeErrors := validateType(typeSchema, data, path)
		errors = append(errors, typeErrors...)
	}

	// 根据数据类型执行对应的验证
	switch data.Type {
	case NUMBER:
		// 数值验证
		if minimumSchema := findObjectKey(schema, "minimum"); minimumSchema != nil && minimumSchema.Type == NUMBER {
			if data.N < minimumSchema.N {
				errors = append(errors, fmt.Sprintf("位于'%s'的数值%g小于最小值%g", path, data.N, minimumSchema.N))
			}
		}
		if maximumSchema := findObjectKey(schema, "maximum"); maximumSchema != nil && maximumSchema.Type == NUMBER {
			if data.N > maximumSchema.N {
				errors = append(errors, fmt.Sprintf("位于'%s'的数值%g大于最大值%g", path, data.N, maximumSchema.N))
			}
		}
		if multipleOfSchema := findObjectKey(schema, "multipleOf"); multipleOfSchema != nil && multipleOfSchema.Type == NUMBER && multipleOfSchema.N > 0 {
			// 检查是否是multipleOf的倍数
			remainder := math.Mod(data.N, multipleOfSchema.N)
			if math.Abs(remainder) > 1e-10 { // 使用小误差范围来处理浮点数比较
				errors = append(errors, fmt.Sprintf("位于'%s'的数值%g不是%g的倍数", path, data.N, multipleOfSchema.N))
			}
		}

	case STRING:
		// 字符串验证
		if minLengthSchema := findObjectKey(schema, "minLength"); minLengthSchema != nil && minLengthSchema.Type == NUMBER {
			minLen := int(minLengthSchema.N)
			if len(data.S) < minLen {
				errors = append(errors, fmt.Sprintf("位于'%s'的字符串长度%d小于最小长度%d", path, len(data.S), minLen))
			}
		}
		if maxLengthSchema := findObjectKey(schema, "maxLength"); maxLengthSchema != nil && maxLengthSchema.Type == NUMBER {
			maxLen := int(maxLengthSchema.N)
			if len(data.S) > maxLen {
				errors = append(errors, fmt.Sprintf("位于'%s'的字符串长度%d大于最大长度%d", path, len(data.S), maxLen))
			}
		}
		if patternSchema := findObjectKey(schema, "pattern"); patternSchema != nil && patternSchema.Type == STRING {
			pattern := patternSchema.S
			matched, err := regexp.MatchString(pattern, data.S)
			if err != nil || !matched {
				errors = append(errors, fmt.Sprintf("位于'%s'的字符串不匹配正则表达式'%s'", path, pattern))
			}
		}

	case ARRAY:
		// 数组验证
		if minItemsSchema := findObjectKey(schema, "minItems"); minItemsSchema != nil && minItemsSchema.Type == NUMBER {
			minItems := int(minItemsSchema.N)
			if len(data.A) < minItems {
				errors = append(errors, fmt.Sprintf("位于'%s'的数组元素数量%d小于最小数量%d", path, len(data.A), minItems))
			}
		}
		if maxItemsSchema := findObjectKey(schema, "maxItems"); maxItemsSchema != nil && maxItemsSchema.Type == NUMBER {
			maxItems := int(maxItemsSchema.N)
			if len(data.A) > maxItems {
				errors = append(errors, fmt.Sprintf("位于'%s'的数组元素数量%d大于最大数量%d", path, len(data.A), maxItems))
			}
		}

		// 验证items
		if itemsSchema := findObjectKey(schema, "items"); itemsSchema != nil {
			if itemsSchema.Type == OBJECT {
				// 所有项使用相同的schema
				for i, item := range data.A {
					itemPath := fmt.Sprintf("%s[%d]", path, i)
					itemErrors := validateJSONSchema(itemsSchema, item, itemPath)
					errors = append(errors, itemErrors...)
				}
			}
		}

	case OBJECT:
		// 对象验证
		if requiredSchema := findObjectKey(schema, "required"); requiredSchema != nil && requiredSchema.Type == ARRAY {
			for _, reqVal := range requiredSchema.A {
				if reqVal.Type == STRING {
					requiredProp := reqVal.S
					if !hasObjectKey(data, requiredProp) {
						errors = append(errors, fmt.Sprintf("位于'%s'的对象缺少必需的属性'%s'", path, requiredProp))
					}
				}
			}
		}

		// 验证properties
		if propertiesSchema := findObjectKey(schema, "properties"); propertiesSchema != nil && propertiesSchema.Type == OBJECT {
			for _, schemaProp := range propertiesSchema.O {
				propName := schemaProp.K
				if propValue := findObjectKeyValue(data, propName); propValue != nil {
					propPath := fmt.Sprintf("%s.%s", path, propName)
					propErrors := validateJSONSchema(schemaProp.V, propValue, propPath)
					errors = append(errors, propErrors...)
				}
			}
		}
	}

	return errors
}

// 验证数据类型
func validateType(typeSchema, data *Value, path string) []string {
	errors := []string{}

	// 类型可以是单个类型或类型数组
	if typeSchema.Type == STRING {
		expectedType := typeSchema.S
		if !matchesType(data, expectedType) {
			errors = append(errors, fmt.Sprintf("位于'%s'的值类型为'%s'，而不是预期的'%s'",
				path, getValueTypeName(data.Type), expectedType))
		}
	} else if typeSchema.Type == ARRAY {
		// 类型是数组时，值必须匹配其中一种类型
		matched := false
		for _, typeVal := range typeSchema.A {
			if typeVal.Type == STRING && matchesType(data, typeVal.S) {
				matched = true
				break
			}
		}
		if !matched {
			errors = append(errors, fmt.Sprintf("位于'%s'的值类型'%s'不在允许的类型列表中",
				path, getValueTypeName(data.Type)))
		}
	}

	return errors
}

// 检查值是否匹配指定的JSON Schema类型
func matchesType(v *Value, typeName string) bool {
	switch typeName {
	case "null":
		return v.Type == NULL
	case "boolean":
		return v.Type == TRUE || v.Type == FALSE
	case "number":
		return v.Type == NUMBER
	case "integer":
		// 对于整数，需要检查NUMBER类型的值是否为整数
		return v.Type == NUMBER && math.Floor(v.N) == v.N
	case "string":
		return v.Type == STRING
	case "array":
		return v.Type == ARRAY
	case "object":
		return v.Type == OBJECT
	}
	return false
}

// 辅助函数：查找对象中的键值对，返回值指针
func findObjectKey(v *Value, key string) *Value {
	if v == nil || v.Type != OBJECT {
		return nil
	}

	for _, member := range v.O {
		if member.K == key {
			return member.V
		}
	}

	return nil
}

// 辅助函数：检查对象是否包含指定键
func hasObjectKey(v *Value, key string) bool {
	return findObjectKey(v, key) != nil
}

// 辅助函数：查找对象中的键值对值
func findObjectKeyValue(v *Value, key string) *Value {
	return findObjectKey(v, key)
}

// runValidate 实现validate命令
func runValidate(args []string, verbose bool) {
	// 解析选项和参数
	outputFormat := "text" // 默认为文本格式
	fileArgs := args

	for i, arg := range args {
		if strings.HasPrefix(arg, "--format=") {
			outputFormat = strings.TrimPrefix(arg, "--format=")
			// 检查输出格式是否有效
			if outputFormat != "text" && outputFormat != "json" {
				fmt.Printf("错误: 无效的输出格式: %s\n", outputFormat)
				fmt.Println("有效的格式: text, json")
				return
			}
			// 从参数列表中移除选项
			fileArgs = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(fileArgs) != 2 {
		fmt.Println("错误: validate命令需要两个文件参数")
		fmt.Println("\n用法: leptjson validate [--format=FORMAT] SCHEMA FILE")
		return
	}

	schemaFile := fileArgs[0]
	dataFile := fileArgs[1]

	if verbose {
		fmt.Printf("使用Schema '%s' 验证文件 '%s'\n", schemaFile, dataFile)
	}

	// 加载Schema文件
	schema, err := loadJSON(schemaFile, verbose)
	if err != nil {
		fmt.Printf("加载Schema失败: %s\n", err)
		os.Exit(1)
	}

	// 加载数据文件
	data, err := loadJSON(dataFile, verbose)
	if err != nil {
		fmt.Printf("加载数据文件失败: %s\n", err)
		os.Exit(1)
	}

	// 执行验证
	result := validateWithSchema(schema, data)

	// 输出验证结果
	if outputFormat == "json" {
		// JSON格式输出
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("生成JSON结果失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(resultJSON))
	} else {
		// 文本格式输出
		if result.Valid {
			fmt.Println("验证通过: 文件符合Schema定义")
		} else {
			fmt.Printf("验证失败: %s\n", result.Message)
			for i, err := range result.Errors {
				fmt.Printf("%d. %s\n", i+1, err)
			}
		}
	}

	// 如果验证失败，设置退出码
	if !result.Valid {
		os.Exit(2) // 使用非0的退出码表示验证失败
	}
}

// JSON Pointer解析器
type CliJSONPointer struct {
	Tokens []string
}

// 创建新的JSON Pointer
func NewJSONPointer(pointer string) (*CliJSONPointer, error) {
	if pointer == "" {
		// 空指针指向根
		return &CliJSONPointer{Tokens: []string{}}, nil
	}

	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("JSON Pointer必须以/开头: %s", pointer)
	}

	// 分割指针并解码token
	tokens := strings.Split(pointer[1:], "/")
	for i, token := range tokens {
		// 解码 ~1 为 / 和 ~0 为 ~
		token = strings.ReplaceAll(token, "~1", "/")
		token = strings.ReplaceAll(token, "~0", "~")
		tokens[i] = token
	}

	return &CliJSONPointer{Tokens: tokens}, nil
}

// 解析JSON Pointer并返回引用的值
func ResolvePointer(doc *Value, pointer *CliJSONPointer) (*Value, []*Value, error) {
	if doc == nil || pointer == nil {
		return nil, nil, fmt.Errorf("文档或指针为空")
	}

	current := doc
	parents := []*Value{}

	// 如果是根指针，直接返回
	if len(pointer.Tokens) == 0 {
		return current, parents, nil
	}

	// 遍历指针的每个token
	for i, token := range pointer.Tokens {
		switch current.Type {
		case OBJECT:
			found := false
			for _, member := range current.O {
				if member.K == token {
					parents = append(parents, current)
					current = member.V
					found = true
					break
				}
			}
			if !found {
				return nil, nil, fmt.Errorf("对象不包含键 '%s' (在token %d)", token, i)
			}

		case ARRAY:
			// 尝试将token解析为数组索引
			index, err := strconv.Atoi(token)
			if err != nil {
				return nil, nil, fmt.Errorf("无效的数组索引 '%s' (在token %d): %v", token, i, err)
			}

			if index < 0 || index >= len(current.A) {
				return nil, nil, fmt.Errorf("数组索引 %d 超出范围 [0-%d] (在token %d)",
					index, len(current.A)-1, i)
			}

			parents = append(parents, current)
			current = current.A[index]

		default:
			return nil, nil, fmt.Errorf("无法在 %s 类型的值上使用token '%s' (在token %d)",
				getValueTypeName(current.Type), token, i)
		}
	}

	return current, parents, nil
}

// 执行添加操作
func PointerAdd(doc *Value, pointer *CliJSONPointer, value *Value) error {
	// 处理根路径的特殊情况
	if len(pointer.Tokens) == 0 {
		// 不能替换根文档
		return fmt.Errorf("不能添加到根路径")
	}

	// 获取父值
	lastToken := pointer.Tokens[len(pointer.Tokens)-1]
	parentPointer := &CliJSONPointer{Tokens: pointer.Tokens[:len(pointer.Tokens)-1]}

	parent, _, err := ResolvePointer(doc, parentPointer)
	if err != nil {
		return fmt.Errorf("找不到父路径: %v", err)
	}

	// 根据父值类型执行添加操作
	switch parent.Type {
	case OBJECT:
		// 检查键是否已存在
		for i, member := range parent.O {
			if member.K == lastToken {
				// 替换现有值
				parent.O[i].V = value
				return nil
			}
		}
		// 添加新键值对
		parent.O = append(parent.O, Member{K: lastToken, V: value})

	case ARRAY:
		index, err := strconv.Atoi(lastToken)
		if err != nil {
			return fmt.Errorf("无效的数组索引: %s", lastToken)
		}

		// 检查是否是末尾的特殊值"-"
		if lastToken == "-" {
			// 添加到数组末尾
			parent.A = append(parent.A, value)
			return nil
		}

		// 检查索引是否在有效范围内
		if index < 0 || index > len(parent.A) {
			return fmt.Errorf("数组索引超出范围: %d", index)
		}

		// 在索引位置插入
		if index == len(parent.A) {
			// 添加到末尾
			parent.A = append(parent.A, value)
		} else {
			// 插入到中间
			parent.A = append(parent.A[:index+1], parent.A[index:]...)
			parent.A[index] = value
		}

	default:
		return fmt.Errorf("无法添加到 %s 类型的值", getValueTypeName(parent.Type))
	}

	return nil
}

// 执行删除操作
func PointerRemove(doc *Value, pointer *CliJSONPointer) error {
	// 处理根路径的特殊情况
	if len(pointer.Tokens) == 0 {
		// 不能删除根文档
		return fmt.Errorf("不能删除根路径")
	}

	// 获取目标值和父值
	_, parents, err := ResolvePointer(doc, pointer)
	if err != nil {
		return fmt.Errorf("找不到要删除的路径: %v", err)
	}

	parent := parents[len(parents)-1]
	lastToken := pointer.Tokens[len(pointer.Tokens)-1]

	// 根据父值类型执行删除操作
	switch parent.Type {
	case OBJECT:
		// 查找并删除成员
		for i, member := range parent.O {
			if member.K == lastToken {
				// 删除匹配的成员
				parent.O = append(parent.O[:i], parent.O[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("对象不包含键: %s", lastToken)

	case ARRAY:
		index, err := strconv.Atoi(lastToken)
		if err != nil {
			return fmt.Errorf("无效的数组索引: %s", lastToken)
		}

		// 检查索引是否在有效范围内
		if index < 0 || index >= len(parent.A) {
			return fmt.Errorf("数组索引超出范围: %d", index)
		}

		// 删除元素
		parent.A = append(parent.A[:index], parent.A[index+1:]...)

	default:
		return fmt.Errorf("无法从 %s 类型的值中删除", getValueTypeName(parent.Type))
	}

	return nil
}

// 执行替换操作
func PointerReplace(doc *Value, pointer *CliJSONPointer, value *Value) error {
	// 处理根路径的特殊情况
	if len(pointer.Tokens) == 0 {
		// 替换整个文档
		*doc = *value
		return nil
	}

	// 获取父值
	lastToken := pointer.Tokens[len(pointer.Tokens)-1]
	parentPointer := &CliJSONPointer{Tokens: pointer.Tokens[:len(pointer.Tokens)-1]}

	parent, _, err := ResolvePointer(doc, parentPointer)
	if err != nil {
		return fmt.Errorf("找不到父路径: %v", err)
	}

	// 根据父值类型执行替换操作
	switch parent.Type {
	case OBJECT:
		found := false
		for i, member := range parent.O {
			if member.K == lastToken {
				// 替换现有值
				parent.O[i].V = value
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("对象不包含键: %s", lastToken)
		}

	case ARRAY:
		index, err := strconv.Atoi(lastToken)
		if err != nil {
			return fmt.Errorf("无效的数组索引: %s", lastToken)
		}

		// 检查索引是否在有效范围内
		if index < 0 || index >= len(parent.A) {
			return fmt.Errorf("数组索引超出范围: %d", index)
		}

		// 替换元素
		parent.A[index] = value

	default:
		return fmt.Errorf("无法在 %s 类型的值上替换", getValueTypeName(parent.Type))
	}

	return nil
}

// 解析JSON值字符串
func parseJSONValue(jsonStr string) (*Value, error) {
	var value Value
	err := Parse(&value, jsonStr)
	if err != PARSE_OK {
		return nil, fmt.Errorf("无效的JSON值: %s", err)
	}
	return &value, nil
}

// 运行pointer命令
func runPointer(args []string, verbose bool) {
	// 默认选项
	operation := "get" // 默认操作是获取
	jsonValue := ""    // add和replace操作的值
	outputFile := ""   // 输出文件
	fileArgs := args   // 不包含选项的参数

	// 解析选项
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--operation=") {
			operation = strings.TrimPrefix(arg, "--operation=")
			// 验证操作类型
			if operation != "get" && operation != "add" && operation != "remove" && operation != "replace" {
				fmt.Printf("错误: 无效的操作类型: %s\n", operation)
				fmt.Println("有效的操作: get, add, remove, replace")
				return
			}
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}

		if strings.HasPrefix(arg, "--value=") {
			jsonValue = strings.TrimPrefix(arg, "--value=")
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}

		if strings.HasPrefix(arg, "--output=") {
			outputFile = strings.TrimPrefix(arg, "--output=")
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}
	}

	// 检查必要的参数
	if len(fileArgs) != 2 {
		fmt.Println("错误: pointer命令需要两个参数")
		fmt.Println("\n用法: leptjson pointer [选项] FILE POINTER")
		return
	}

	inputFile := fileArgs[0]
	pointerStr := fileArgs[1]

	// 默认输出到原文件
	if outputFile == "" && (operation == "add" || operation == "remove" || operation == "replace") {
		outputFile = inputFile
	}

	// 验证add和replace操作需要value参数
	if (operation == "add" || operation == "replace") && jsonValue == "" {
		fmt.Printf("错误: %s操作需要--value选项\n", operation)
		return
	}

	if verbose {
		fmt.Printf("对文件 '%s' 执行 %s 操作，pointer: '%s'\n", inputFile, operation, pointerStr)
	}

	// 加载JSON文档
	doc, err := loadJSON(inputFile, verbose)
	if err != nil {
		fmt.Printf("加载JSON文档失败: %s\n", err)
		os.Exit(1)
	}

	// 解析JSON Pointer
	pointer, err := NewJSONPointer(pointerStr)
	if err != nil {
		fmt.Printf("解析JSON Pointer失败: %s\n", err)
		os.Exit(1)
	}

	// 根据操作类型执行不同的操作
	switch operation {
	case "get":
		// 获取值
		value, _, err := ResolvePointer(doc, pointer)
		if err != nil {
			fmt.Printf("解析指针失败: %s\n", err)
			os.Exit(1)
		}

		// 格式化并输出结果
		result, err := formatJSON(value, "  ")
		if err != nil {
			fmt.Printf("格式化结果失败: %s\n", err)
			os.Exit(1)
		}

		fmt.Println(result)

	case "add":
		// 添加或替换值
		valueObj, err := parseJSONValue(jsonValue)
		if err != nil {
			fmt.Printf("解析JSON值失败: %s\n", err)
			os.Exit(1)
		}

		// 执行添加操作
		if err := PointerAdd(doc, pointer, valueObj); err != nil {
			fmt.Printf("添加值失败: %s\n", err)
			os.Exit(1)
		}

		// 保存修改后的文档
		if outputFile != "" {
			jsonStr, err := formatJSON(doc, "  ")
			if err != nil {
				fmt.Printf("格式化JSON失败: %s\n", err)
				os.Exit(1)
			}

			if err := saveJSON(outputFile, jsonStr, verbose); err != nil {
				fmt.Printf("保存文件失败: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("已成功添加值并保存到 %s\n", outputFile)
		}

	case "remove":
		// 删除值
		if err := PointerRemove(doc, pointer); err != nil {
			fmt.Printf("删除值失败: %s\n", err)
			os.Exit(1)
		}

		// 保存修改后的文档
		if outputFile != "" {
			jsonStr, err := formatJSON(doc, "  ")
			if err != nil {
				fmt.Printf("格式化JSON失败: %s\n", err)
				os.Exit(1)
			}

			if err := saveJSON(outputFile, jsonStr, verbose); err != nil {
				fmt.Printf("保存文件失败: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("已成功删除值并保存到 %s\n", outputFile)
		}

	case "replace":
		// 替换值
		valueObj, err := parseJSONValue(jsonValue)
		if err != nil {
			fmt.Printf("解析JSON值失败: %s\n", err)
			os.Exit(1)
		}

		// 执行替换操作
		if err := PointerReplace(doc, pointer, valueObj); err != nil {
			fmt.Printf("替换值失败: %s\n", err)
			os.Exit(1)
		}

		// 保存修改后的文档
		if outputFile != "" {
			jsonStr, err := formatJSON(doc, "  ")
			if err != nil {
				fmt.Printf("格式化JSON失败: %s\n", err)
				os.Exit(1)
			}

			if err := saveJSON(outputFile, jsonStr, verbose); err != nil {
				fmt.Printf("保存文件失败: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("已成功替换值并保存到 %s\n", outputFile)
		}
	}
}

// JSON Patch操作类型
const (
	OpAdd     = "add"
	OpRemove  = "remove"
	OpReplace = "replace"
	OpMove    = "move"
	OpCopy    = "copy"
	OpTest    = "test"
)

// JSON Patch操作
type CliPatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	From  string `json:"from,omitempty"`
	Value *Value `json:"value,omitempty"`
}

// 解析JSON Patch
func parsePatch(patchDoc *Value) ([]CliPatchOperation, error) {
	if patchDoc.Type != ARRAY {
		return nil, fmt.Errorf("JSON Patch必须是操作数组")
	}

	operations := make([]CliPatchOperation, 0, len(patchDoc.A))

	for i, op := range patchDoc.A {
		if op.Type != OBJECT {
			return nil, fmt.Errorf("Patch操作 #%d 必须是对象", i+1)
		}

		// 获取操作类型
		opType := findObjectKeyValue(op, "op")
		if opType == nil || opType.Type != STRING {
			return nil, fmt.Errorf("Patch操作 #%d 缺少有效的 'op' 字段", i+1)
		}

		// 获取路径
		pathVal := findObjectKeyValue(op, "path")
		if pathVal == nil || pathVal.Type != STRING {
			return nil, fmt.Errorf("Patch操作 #%d 缺少有效的 'path' 字段", i+1)
		}

		// 创建操作对象
		operation := CliPatchOperation{
			Op:   opType.S,
			Path: pathVal.S,
		}

		// 根据操作类型处理额外字段
		switch opType.S {
		case OpAdd, OpReplace, OpTest:
			valueVal := findObjectKeyValue(op, "value")
			if valueVal == nil {
				return nil, fmt.Errorf("Patch操作 #%d (%s) 缺少必需的 'value' 字段", i+1, opType.S)
			}
			operation.Value = valueVal

		case OpMove, OpCopy:
			fromVal := findObjectKeyValue(op, "from")
			if fromVal == nil || fromVal.Type != STRING {
				return nil, fmt.Errorf("Patch操作 #%d (%s) 缺少有效的 'from' 字段", i+1, opType.S)
			}
			operation.From = fromVal.S

		case OpRemove:
			// remove只需要path

		default:
			return nil, fmt.Errorf("Patch操作 #%d 包含无效的操作类型: %s", i+1, opType.S)
		}

		operations = append(operations, operation)
	}

	return operations, nil
}

// 应用JSON Patch
func applyPatch(doc *Value, operations []CliPatchOperation, testOnly bool) error {
	// 执行所有操作
	for i, op := range operations {
		// 解析路径
		path, err := NewJSONPointer(op.Path)
		if err != nil {
			return fmt.Errorf("操作 #%d: 无效的路径 '%s': %v", i+1, op.Path, err)
		}

		// 根据操作类型执行不同的操作
		switch op.Op {
		case OpAdd:
			if testOnly {
				continue
			}
			if err := PointerAdd(doc, path, op.Value); err != nil {
				return fmt.Errorf("操作 #%d (add): %v", i+1, err)
			}

		case OpRemove:
			if testOnly {
				// 在测试模式下，只检查路径是否存在
				_, _, err := ResolvePointer(doc, path)
				if err != nil {
					return fmt.Errorf("操作 #%d (remove): %v", i+1, err)
				}
			} else {
				if err := PointerRemove(doc, path); err != nil {
					return fmt.Errorf("操作 #%d (remove): %v", i+1, err)
				}
			}

		case OpReplace:
			if testOnly {
				// 在测试模式下，只检查路径是否存在
				_, _, err := ResolvePointer(doc, path)
				if err != nil {
					return fmt.Errorf("操作 #%d (replace): %v", i+1, err)
				}
			} else {
				if err := PointerReplace(doc, path, op.Value); err != nil {
					return fmt.Errorf("操作 #%d (replace): %v", i+1, err)
				}
			}

		case OpMove:
			if testOnly {
				continue
			}

			// 解析源路径
			fromPath, err := NewJSONPointer(op.From)
			if err != nil {
				return fmt.Errorf("操作 #%d: 无效的源路径 '%s': %v", i+1, op.From, err)
			}

			// 获取源值
			fromValue, _, err := ResolvePointer(doc, fromPath)
			if err != nil {
				return fmt.Errorf("操作 #%d (move): 源路径错误: %v", i+1, err)
			}

			// 复制源值
			valueClone := &Value{}
			*valueClone = *fromValue

			// 移除源
			if err := PointerRemove(doc, fromPath); err != nil {
				return fmt.Errorf("操作 #%d (move): 删除源失败: %v", i+1, err)
			}

			// 添加到目标
			if err := PointerAdd(doc, path, valueClone); err != nil {
				return fmt.Errorf("操作 #%d (move): 添加到目标失败: %v", i+1, err)
			}

		case OpCopy:
			if testOnly {
				continue
			}

			// 解析源路径
			fromPath, err := NewJSONPointer(op.From)
			if err != nil {
				return fmt.Errorf("操作 #%d: 无效的源路径 '%s': %v", i+1, op.From, err)
			}

			// 获取源值
			fromValue, _, err := ResolvePointer(doc, fromPath)
			if err != nil {
				return fmt.Errorf("操作 #%d (copy): 源路径错误: %v", i+1, err)
			}

			// 复制源值
			valueClone := &Value{}
			*valueClone = *fromValue

			// 添加到目标
			if err := PointerAdd(doc, path, valueClone); err != nil {
				return fmt.Errorf("操作 #%d (copy): 添加到目标失败: %v", i+1, err)
			}

		case OpTest:
			// 获取当前值
			current, _, err := ResolvePointer(doc, path)
			if err != nil {
				return fmt.Errorf("操作 #%d (test): 路径不存在: %v", i+1, err)
			}

			// 比较值
			if !ValuesEqual(current, op.Value) {
				actualJSON, _ := Stringify(current)
				expectedJSON, _ := Stringify(op.Value)
				return fmt.Errorf("操作 #%d (test): 值不匹配\n  路径: %s\n  期望: %s\n  实际: %s",
					i+1, op.Path, expectedJSON, actualJSON)
			}

		default:
			return fmt.Errorf("操作 #%d: 不支持的操作类型: %s", i+1, op.Op)
		}
	}

	return nil
}

// 比较两个值是否相等
func ValuesEqual(a, b *Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 类型必须相同
	if a.Type != b.Type {
		return false
	}

	// 根据类型比较值
	switch a.Type {
	case NULL:
		return true
	case TRUE, FALSE:
		return a.Type == b.Type
	case NUMBER:
		return a.N == b.N
	case STRING:
		return a.S == b.S
	case ARRAY:
		if len(a.A) != len(b.A) {
			return false
		}
		for i := range a.A {
			if !ValuesEqual(a.A[i], b.A[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		if len(a.O) != len(b.O) {
			return false
		}

		// 创建b的键值映射，用于快速查找
		bMembers := make(map[string]*Value)
		for _, member := range b.O {
			bMembers[member.K] = member.V
		}

		// 检查a中的每个键值对是否都在b中且值相等
		for _, member := range a.O {
			bValue, ok := bMembers[member.K]
			if !ok || !ValuesEqual(member.V, bValue) {
				return false
			}
		}
		return true
	}

	return false
}

// 运行patch命令
func runPatch(args []string, verbose bool) {
	// 解析选项
	inPlace := false
	testOnly := false
	fileArgs := args

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--in-place" {
			inPlace = true
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}

		if arg == "--test" {
			testOnly = true
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}
	}

	// 检查必要的参数
	if len(fileArgs) < 2 || len(fileArgs) > 3 {
		fmt.Println("错误: patch命令需要2-3个参数")
		fmt.Println("\n用法: leptjson patch [选项] PATCH FILE [OUTPUT]")
		return
	}

	patchFile := fileArgs[0]
	targetFile := fileArgs[1]
	outputFile := ""

	if len(fileArgs) == 3 {
		outputFile = fileArgs[2]
	} else if inPlace {
		outputFile = targetFile
	} else {
		// 默认输出文件名
		ext := filepath.Ext(targetFile)
		baseName := targetFile[:len(targetFile)-len(ext)]
		outputFile = baseName + ".patched" + ext
	}

	if verbose {
		if testOnly {
			fmt.Printf("测试补丁 '%s' 应用于 '%s'\n", patchFile, targetFile)
		} else {
			fmt.Printf("应用补丁 '%s' 到 '%s'，输出到 '%s'\n", patchFile, targetFile, outputFile)
		}
	}

	// 加载补丁文件
	patchDoc, err := loadJSON(patchFile, verbose)
	if err != nil {
		fmt.Printf("加载补丁失败: %s\n", err)
		os.Exit(1)
	}

	// 解析补丁操作
	operations, err := parsePatch(patchDoc)
	if err != nil {
		fmt.Printf("解析补丁失败: %s\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("找到 %d 个补丁操作\n", len(operations))
		for i, op := range operations {
			fmt.Printf("  %d. %s %s\n", i+1, op.Op, op.Path)
		}
	}

	// 加载目标文件
	targetDoc, err := loadJSON(targetFile, verbose)
	if err != nil {
		fmt.Printf("加载目标文件失败: %s\n", err)
		os.Exit(1)
	}

	// 应用补丁
	err = applyPatch(targetDoc, operations, testOnly)
	if err != nil {
		fmt.Printf("应用补丁失败: %s\n", err)
		os.Exit(1)
	}

	if testOnly {
		fmt.Println("补丁测试成功: 所有操作都可以成功应用")
		return
	}

	// 保存结果
	resultJSON, err := formatJSON(targetDoc, "  ")
	if err != nil {
		fmt.Printf("格式化结果失败: %s\n", err)
		os.Exit(1)
	}

	if err := saveJSON(outputFile, resultJSON, verbose); err != nil {
		fmt.Printf("保存结果失败: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("补丁应用成功: 输出保存到 %s\n", outputFile)
}

// 应用JSON Merge Patch
func applyMergePatch(target, patch *Value) error {
	// 如果patch不是对象，则直接替换整个目标
	if patch.Type != OBJECT {
		*target = *patch
		return nil
	}

	// 确保目标是对象
	if target.Type != OBJECT {
		// 如果目标不是对象，就创建一个空对象
		target.Type = OBJECT
		target.O = []Member{}
	}

	// 遍历patch中的所有成员
	for _, patchMember := range patch.O {
		patchKey := patchMember.K
		patchValue := patchMember.V

		// 如果补丁值是null，从目标中删除该键
		if patchValue.Type == NULL {
			// 查找并删除目标中的键
			for i, targetMember := range target.O {
				if targetMember.K == patchKey {
					// 删除这个成员
					target.O = append(target.O[:i], target.O[i+1:]...)
					break
				}
			}
			continue
		}

		// 查找目标中是否存在该键
		var targetValue *Value
		targetIndex := -1
		for i, targetMember := range target.O {
			if targetMember.K == patchKey {
				targetValue = targetMember.V
				targetIndex = i
				break
			}
		}

		// 如果patch的值是对象并且目标中存在该键且值也是对象，递归合并
		if patchValue.Type == OBJECT && targetValue != nil && targetValue.Type == OBJECT {
			// 递归合并对象
			if err := applyMergePatch(targetValue, patchValue); err != nil {
				return fmt.Errorf("合并键 '%s' 时出错: %v", patchKey, err)
			}
		} else {
			// 否则，直接替换/添加值
			if targetIndex >= 0 {
				// 替换现有值
				target.O[targetIndex].V = patchValue
			} else {
				// 添加新键值对
				target.O = append(target.O, Member{K: patchKey, V: patchValue})
			}
		}
	}

	return nil
}

// 运行merge-patch命令
func runMergePatch(args []string, verbose bool) {
	// 解析选项
	inPlace := false
	fileArgs := args

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--in-place" {
			inPlace = true
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}
	}

	// 检查必要的参数
	if len(fileArgs) < 2 || len(fileArgs) > 3 {
		fmt.Println("错误: merge-patch命令需要2-3个参数")
		fmt.Println("\n用法: leptjson merge-patch [选项] PATCH FILE [OUTPUT]")
		return
	}

	patchFile := fileArgs[0]
	targetFile := fileArgs[1]
	outputFile := ""

	if len(fileArgs) == 3 {
		outputFile = fileArgs[2]
	} else if inPlace {
		outputFile = targetFile
	} else {
		// 默认输出文件名
		ext := filepath.Ext(targetFile)
		baseName := targetFile[:len(targetFile)-len(ext)]
		outputFile = baseName + ".merged" + ext
	}

	if verbose {
		fmt.Printf("应用Merge Patch '%s' 到 '%s'，输出到 '%s'\n", patchFile, targetFile, outputFile)
	}

	// 加载补丁文件
	patchDoc, err := loadJSON(patchFile, verbose)
	if err != nil {
		fmt.Printf("加载Merge Patch失败: %s\n", err)
		os.Exit(1)
	}

	// 加载目标文件
	targetDoc, err := loadJSON(targetFile, verbose)
	if err != nil {
		fmt.Printf("加载目标文件失败: %s\n", err)
		os.Exit(1)
	}

	// 应用Merge Patch
	if err := applyMergePatch(targetDoc, patchDoc); err != nil {
		fmt.Printf("应用Merge Patch失败: %s\n", err)
		os.Exit(1)
	}

	// 保存结果
	resultJSON, err := formatJSON(targetDoc, "  ")
	if err != nil {
		fmt.Printf("格式化结果失败: %s\n", err)
		os.Exit(1)
	}

	if err := saveJSON(outputFile, resultJSON, verbose); err != nil {
		fmt.Printf("保存结果失败: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Merge Patch应用成功: 输出保存到 %s\n", outputFile)
}

// 实现runPath命令
func runPath(args []string, verbose bool) {
	// 解析选项
	outputFormat := "pretty" // 默认为美化输出
	showAll := false         // 默认只显示前10个结果
	csvFile := ""            // CSV输出文件
	showPath := true         // 显示路径信息
	fileArgs := args

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--output=") {
			outputFormat = strings.TrimPrefix(arg, "--output=")
			// 验证输出格式
			if outputFormat != "compact" && outputFormat != "pretty" &&
				outputFormat != "raw" && outputFormat != "table" {
				fmt.Printf("错误: 无效的输出格式: %s\n", outputFormat)
				fmt.Println("有效的格式: compact, pretty, raw, table")
				return
			}
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}

		if arg == "--all" {
			showAll = true
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}

		if strings.HasPrefix(arg, "--csv=") {
			csvFile = strings.TrimPrefix(arg, "--csv=")
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}

		if arg == "--no-path" {
			showPath = false
			// 从参数列表中移除
			fileArgs = append(fileArgs[:i], fileArgs[i+1:]...)
			i--
			continue
		}
	}

	// 检查必要参数
	if len(fileArgs) != 2 {
		fmt.Println("错误: path命令需要两个参数")
		fmt.Println("\n用法: leptjson path [选项] FILE JSONPATH")
		return
	}

	filePath := fileArgs[0]
	jsonPathExpr := fileArgs[1]

	if verbose {
		fmt.Printf("在文件 %s 中查询 JSONPath: %s\n", filePath, jsonPathExpr)
	}

	// 加载JSON
	doc, err := loadJSON(filePath, verbose)
	if err != nil {
		fmt.Printf("加载JSON失败: %s\n", err)
		os.Exit(1)
	}

	// 解析JSONPath并执行查询
	path, err := NewJSONPath(jsonPathExpr)
	if err != nil {
		fmt.Printf("解析JSONPath失败: %s\n", err)
		os.Exit(1)
	}

	results, err := path.Query(doc)
	if err != nil {
		fmt.Printf("执行查询失败: %s\n", err)
		os.Exit(1)
	}

	// 显示结果数量
	totalResults := len(results)
	if verbose {
		fmt.Printf("找到 %d 个匹配结果\n", totalResults)
	}

	// 如果没有结果，提前返回
	if totalResults == 0 {
		fmt.Println("没有找到匹配的结果")
		return
	}

	// 限制结果数量（除非使用--all选项）
	displayResults := results
	if !showAll && totalResults > 10 {
		displayResults = results[:10]
		fmt.Printf("显示前10个结果（共 %d 个匹配项）。使用 --all 查看所有结果。\n", totalResults)
	}

	// 如果需要CSV输出
	if csvFile != "" {
		if err := saveResultsAsCSV(displayResults, csvFile); err != nil {
			fmt.Printf("保存CSV失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("结果已保存到CSV文件: %s\n", csvFile)
	}

	// 根据输出格式显示结果
	switch outputFormat {
	case "compact":
		// 紧凑输出
		for i, result := range displayResults {
			output, err := minifyJSON(result)
			if err != nil {
				fmt.Printf("格式化结果 #%d 失败: %s\n", i+1, err)
				continue
			}
			if showPath {
				fmt.Printf("结果 #%d: %s\n", i+1, output)
			} else {
				fmt.Println(output)
			}
		}

	case "pretty":
		// 美化输出
		for i, result := range displayResults {
			output, err := formatJSON(result, "  ")
			if err != nil {
				fmt.Printf("格式化结果 #%d 失败: %s\n", i+1, err)
				continue
			}
			if showPath {
				fmt.Printf("结果 #%d:\n%s\n", i+1, output)
			} else {
				fmt.Println(output)
			}
		}

	case "raw":
		// 原始值输出
		for i, result := range displayResults {
			if showPath {
				fmt.Printf("结果 #%d: ", i+1)
			}

			switch result.Type {
			case STRING:
				fmt.Println(result.S)
			case NUMBER:
				fmt.Println(result.N)
			case TRUE:
				fmt.Println("true")
			case FALSE:
				fmt.Println("false")
			case NULL:
				fmt.Println("null")
			default:
				// 对象和数组使用格式化输出
				output, err := formatJSON(result, "  ")
				if err != nil {
					fmt.Printf("格式化结果失败: %s\n", err)
					continue
				}
				fmt.Println(output)
			}
		}

	case "table":
		// 表格输出 (仅适用于数组内的对象且具有相同的结构)
		printResultsAsTable(displayResults)
	}
}

// 将结果保存为CSV文件
func saveResultsAsCSV(results []*Value, filename string) error {
	// 创建CSV文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建CSV文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果结果是对象数组，尝试提取所有键作为标题
	allKeys := make(map[string]bool)
	objectResults := []*Value{}

	for _, result := range results {
		if result.Type == OBJECT {
			objectResults = append(objectResults, result)
			for _, member := range result.O {
				allKeys[member.K] = true
			}
		}
	}

	if len(objectResults) > 0 {
		// 将所有键排序，以确保一致的列顺序
		headers := make([]string, 0, len(allKeys))
		for key := range allKeys {
			headers = append(headers, key)
		}
		sort.Strings(headers)

		// 写入标题行
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("写入CSV标题失败: %w", err)
		}

		// 写入每个对象的值
		for _, obj := range objectResults {
			row := make([]string, len(headers))
			for i, key := range headers {
				// 查找对象中的键
				for _, member := range obj.O {
					if member.K == key {
						// 将值转换为字符串
						row[i] = valueToString(member.V)
						break
					}
				}
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("写入CSV行失败: %w", err)
			}
		}
	} else {
		// 对于简单值，直接写入
		row := make([]string, len(results))
		for i, result := range results {
			row[i] = valueToString(result)
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入CSV数据失败: %w", err)
		}
	}

	return nil
}

// 将结果打印为表格
func printResultsAsTable(results []*Value) {
	// 表格输出仅对对象数组有意义
	objectResults := []*Value{}
	for _, result := range results {
		if result.Type == OBJECT {
			objectResults = append(objectResults, result)
		}
	}

	if len(objectResults) == 0 {
		fmt.Println("无法以表格形式显示结果: 需要对象数组")
		// 回退到JSON格式
		for i, result := range results {
			output, _ := formatJSON(result, "  ")
			fmt.Printf("结果 #%d:\n%s\n", i+1, output)
		}
		return
	}

	// 收集所有对象的键
	allKeys := make(map[string]bool)
	for _, obj := range objectResults {
		for _, member := range obj.O {
			allKeys[member.K] = true
		}
	}

	// 将所有键排序以确保一致的列顺序
	headers := make([]string, 0, len(allKeys))
	for key := range allKeys {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	// 构建表格数据
	rows := make([][]string, len(objectResults))
	for i, obj := range objectResults {
		row := make([]string, len(headers))
		for j, key := range headers {
			// 查找对象中的键
			found := false
			for _, member := range obj.O {
				if member.K == key {
					// 将值转换为字符串
					row[j] = valueToString(member.V)
					found = true
					break
				}
			}
			if !found {
				row[j] = ""
			}
		}
		rows[i] = row
	}

	// 计算每列的最大宽度
	colWidths := make([]int, len(headers))
	for j, header := range headers {
		colWidths[j] = len(header)
		for i := range rows {
			if len(rows[i][j]) > colWidths[j] {
				colWidths[j] = len(rows[i][j])
			}
		}
		// 限制列宽
		if colWidths[j] > 30 {
			colWidths[j] = 30
		}
	}

	// 打印表头
	fmt.Print("|")
	for j, header := range headers {
		fmt.Printf(" %-*s |", colWidths[j], truncateString(header, colWidths[j]))
	}
	fmt.Println()

	// 打印分隔线
	fmt.Print("|")
	for j := range headers {
		fmt.Print(strings.Repeat("-", colWidths[j]+2) + "|")
	}
	fmt.Println()

	// 打印行
	for i := range rows {
		fmt.Print("|")
		for j := range headers {
			fmt.Printf(" %-*s |", colWidths[j], truncateString(rows[i][j], colWidths[j]))
		}
		fmt.Println()
	}
}

// 将Value转换为字符串
func valueToString(v *Value) string {
	if v == nil {
		return "null"
	}

	switch v.Type {
	case NULL:
		return "null"
	case TRUE:
		return "true"
	case FALSE:
		return "false"
	case NUMBER:
		return fmt.Sprintf("%g", v.N)
	case STRING:
		return v.S
	case OBJECT:
		// 对于对象，返回键数量摘要
		return fmt.Sprintf("{...} (%d keys)", len(v.O))
	case ARRAY:
		// 对于数组，返回元素数量摘要
		return fmt.Sprintf("[...] (%d items)", len(v.A))
	default:
		return "unknown"
	}
}

// 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
