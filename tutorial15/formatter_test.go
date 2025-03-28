package tutorial15

import (
	"strings"
	"testing"
)

func TestStringifyIndent(t *testing.T) {
	// 创建一个复杂的JSON对象进行测试
	v := &Value{}
	SetObject(v)

	// 添加各种类型的字段
	nameVal := &Value{}
	SetString(nameVal, "张三")
	nameObj := SetObjectValue(v, "name")
	nameObj.Type = nameVal.Type
	nameObj.S = nameVal.S

	ageVal := &Value{}
	SetNumber(ageVal, 30)
	ageObj := SetObjectValue(v, "age")
	ageObj.Type = ageVal.Type
	ageObj.N = ageVal.N

	marriedVal := &Value{}
	SetBool(marriedVal, false)
	marriedObj := SetObjectValue(v, "married")
	marriedObj.Type = marriedVal.Type
	marriedObj.B = marriedVal.B

	// 添加数组
	hobbiesVal := &Value{}
	SetArray(hobbiesVal)
	hobby1 := &Value{}
	SetString(hobby1, "阅读")
	PushBackArrayElement(hobbiesVal, hobby1)
	hobby2 := &Value{}
	SetString(hobby2, "旅行")
	PushBackArrayElement(hobbiesVal, hobby2)
	hobby3 := &Value{}
	SetString(hobby3, "编程")
	PushBackArrayElement(hobbiesVal, hobby3)

	hobbiesObj := SetObjectValue(v, "hobbies")
	hobbiesObj.Type = hobbiesVal.Type
	hobbiesObj.Elements = hobbiesVal.Elements

	// 添加嵌套对象
	addressVal := &Value{}
	SetObject(addressVal)
	cityVal := &Value{}
	SetString(cityVal, "北京")
	cityObj := SetObjectValue(addressVal, "city")
	cityObj.Type = cityVal.Type
	cityObj.S = cityVal.S

	streetVal := &Value{}
	SetString(streetVal, "中关村大街")
	streetObj := SetObjectValue(addressVal, "street")
	streetObj.Type = streetVal.Type
	streetObj.S = streetVal.S

	addressObj := SetObjectValue(v, "address")
	addressObj.Type = addressVal.Type
	addressObj.Members = addressVal.Members

	// 添加null值
	nullVal := &Value{}
	SetNull(nullVal)
	nullObj := SetObjectValue(v, "nothing")
	nullObj.Type = nullVal.Type

	// 测试用例
	testCases := []struct {
		name      string
		indent    string
		checkFn   func(string) bool
		lineCount int
	}{
		{
			name:   "无缩进(紧凑格式)",
			indent: "",
			checkFn: func(s string) bool {
				return !strings.Contains(s, "\n") && strings.Contains(s, "\"name\":\"张三\"")
			},
			lineCount: 1,
		},
		{
			name:   "两个空格缩进",
			indent: "  ",
			checkFn: func(s string) bool {
				return strings.Contains(s, "\"name\": \"张三\"") && strings.Contains(s, "  \"age\": 30")
			},
			lineCount: 17, // 预计的行数
		},
		{
			name:   "四个空格缩进",
			indent: "    ",
			checkFn: func(s string) bool {
				return strings.Contains(s, "    \"name\": \"张三\"") && strings.Contains(s, "    \"married\": false")
			},
			lineCount: 17, // 预计的行数
		},
		{
			name:   "Tab缩进",
			indent: "\t",
			checkFn: func(s string) bool {
				return strings.Contains(s, "\t\"hobbies\": [") && strings.Contains(s, "\t\t\"阅读\"")
			},
			lineCount: 17, // 预计的行数
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StringifyIndent(v, tc.indent)
			if err != nil {
				t.Fatalf("格式化失败: %v", err)
			}

			// 检查格式化结果是否符合期望
			if !tc.checkFn(result) {
				t.Errorf("格式化结果不符合预期:\n%s", result)
			}

			// 检查行数是否符合预期
			lines := strings.Split(result, "\n")
			if len(lines) != tc.lineCount {
				t.Errorf("行数不符合预期: 实际=%d, 预期=%d", len(lines), tc.lineCount)
			}

			// 检查格式化后的JSON是否可以被解析回来
			var parsed Value
			err = Parse(&parsed, result)
			if err != nil {
				t.Errorf("无法解析格式化后的JSON: %v", err)
			}

			// 检查解析回来的对象是否与原对象相等
			if !Equal(&parsed, v) {
				t.Errorf("解析回来的对象与原对象不相等")
			}

			// 打印格式化结果供手动检查
			t.Logf("格式化结果:\n%s", result)
		})
	}
}
