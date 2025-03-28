package tutorial15

import (
	"strings"
	"testing"
)

func TestMainStringifyIndent(t *testing.T) {
	// 创建一个测试对象
	v := &Value{}
	SetObject(v)

	// 设置一些属性
	nameVal := &Value{}
	SetString(nameVal, "张三")
	objName := SetObjectValue(v, "name")
	objName.Type = nameVal.Type
	objName.S = nameVal.S

	ageVal := &Value{}
	SetNumber(ageVal, 30)
	objAge := SetObjectValue(v, "age")
	objAge.Type = ageVal.Type
	objAge.N = ageVal.N

	// 测试不同缩进
	testCases := []struct {
		name      string
		indentStr string
	}{
		{"紧凑格式", ""},
		{"两空格缩进", "  "},
		{"四空格缩进", "    "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StringifyIndent(v, tc.indentStr)
			if err != nil {
				t.Fatalf("StringifyIndent failed: %v", err)
			}

			t.Logf("格式化结果:\n%s", result)

			// 基本验证
			if tc.indentStr == "" && strings.Contains(result, "\n") {
				t.Errorf("紧凑格式不应该包含换行符")
			}

			// 解析回来确保正确
			parsedValue := &Value{}
			if err := Parse(parsedValue, result); err != nil {
				t.Errorf("无法解析格式化后的JSON: %v", err)
			} else if !Equal(parsedValue, v) {
				t.Errorf("解析回来的值与原始值不同")
			}
		})
	}
}
