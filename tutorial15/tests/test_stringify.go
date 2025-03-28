package tutorial15_test

import (
	"fmt"
	"os"

	"github.com/Cactusinhand/go-json-tutorial/tutorial15"
)

func main() {
	// 创建测试对象
	v := &tutorial15.Value{}
	tutorial15.SetObject(v)

	// 添加属性
	nameVal := &tutorial15.Value{}
	tutorial15.SetString(nameVal, "张三")
	objName := tutorial15.SetObjectValue(v, "name")
	objName.Type = nameVal.Type
	objName.S = nameVal.S

	ageVal := &tutorial15.Value{}
	tutorial15.SetNumber(ageVal, 30)
	objAge := tutorial15.SetObjectValue(v, "age")
	objAge.Type = ageVal.Type
	objAge.N = ageVal.N

	// 测试紧凑格式
	fmt.Println("=== 紧凑格式 ===")
	compactJSON, err := tutorial15.StringifyIndent(v, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "格式化失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(compactJSON)

	// 测试缩进格式
	fmt.Println("\n=== 两空格缩进 ===")
	indentedJSON, err := tutorial15.StringifyIndent(v, "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "格式化失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(indentedJSON)
}
