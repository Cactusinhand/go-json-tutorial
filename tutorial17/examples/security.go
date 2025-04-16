package security

import (
	"fmt"
	"strings"

	leptjson ".." // 使用相对路径
)

// RunSecurityExample 演示安全功能的使用
func RunSecurityExample() {
	fmt.Println("=== 安全性功能演示 ===")

	// 1. 创建恶意构造的大型输入
	maliciousLongString := `"` + strings.Repeat("x", 10000) + `"`                     // 非常长的字符串
	maliciousDeepArray := strings.Repeat("[", 1000) + "1" + strings.Repeat("]", 1000) // 深度嵌套数组
	maliciousLargeArray := "[" + strings.Repeat("1,", 20000) + "1]"                   // 超大数组

	// 构建超大对象
	var objBuilder strings.Builder
	objBuilder.WriteString("{")
	for i := 0; i < 20000; i++ {
		if i > 0 {
			objBuilder.WriteString(",")
		}
		objBuilder.WriteString(fmt.Sprintf(`"key%d":%d`, i, i))
	}
	objBuilder.WriteString("}")
	maliciousLargeObject := objBuilder.String()

	maliciousLargeNumber := "1e1000" // 非常大的数字

	// 2. 设置安全选项
	secureOptions := leptjson.DefaultParseOptions()
	// 可以根据需要调整这些值
	secureOptions.MaxDepth = 100             // 限制嵌套深度
	secureOptions.MaxStringLength = 1000     // 限制字符串长度
	secureOptions.MaxArraySize = 1000        // 限制数组元素数量
	secureOptions.MaxObjectSize = 1000       // 限制对象成员数量
	secureOptions.MaxTotalSize = 1024 * 1024 // 限制总输入大小为1MB
	secureOptions.MaxNumberValue = 1e100     // 限制最大数值
	secureOptions.MinNumberValue = -1e100    // 限制最小数值
	secureOptions.EnabledSecurity = true     // 启用安全检查

	// 3. 演示各种安全检查
	testSecurity("超长字符串", maliciousLongString, secureOptions)
	testSecurity("深度嵌套数组", maliciousDeepArray, secureOptions)
	testSecurity("超大数组", maliciousLargeArray, secureOptions)
	testSecurity("超大对象", maliciousLargeObject, secureOptions)
	testSecurity("超大数值", maliciousLargeNumber, secureOptions)

	// 4. 演示禁用安全检查（注意：这可能导致程序崩溃或性能问题）
	fmt.Println("\n禁用安全检查示例（注意风险）:")
	unsafeOptions := secureOptions
	unsafeOptions.EnabledSecurity = false

	// 尝试解析一个安全的输入
	smallInput := `{"name":"John","age":30,"isStudent":false}`
	testSecurity("禁用安全检查-安全输入", smallInput, unsafeOptions)

	fmt.Println("\n=== 安全性功能演示结束 ===")
}

// testSecurity 辅助函数，测试解析安全性
func testSecurity(name string, input string, options leptjson.ParseOptions) {
	fmt.Printf("\n测试: %s\n", name)
	fmt.Printf("输入长度: %d 字节\n", len(input))

	var v leptjson.Value
	err := leptjson.ParseWithOptions(&v, input, options)

	if err == leptjson.PARSE_OK {
		fmt.Println("解析成功!")
	} else {
		fmt.Printf("解析失败: %s\n", leptjson.GetErrorMessage(err))
	}
}

// 如果直接运行这个文件，调用示例函数
func main() {
	RunSecurityExample()
}
