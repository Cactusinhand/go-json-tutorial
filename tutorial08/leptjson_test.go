// leptjson_test.go - Go语言版JSON库测试
package leptjson

import (
	"fmt"
	"math"
	"testing"
)

// 测试解析null值
func TestParseNull(t *testing.T) {
	v := Value{Type: TRUE} // 初始化为非NULL类型
	if err := Parse(&v, "null"); err != PARSE_OK {
		t.Errorf("期望解析成功，但返回错误: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}

// 测试解析true值
func TestParseTrue(t *testing.T) {
	v := Value{Type: FALSE} // 初始化为非TRUE类型
	if err := Parse(&v, "true"); err != PARSE_OK {
		t.Errorf("期望解析成功，但返回错误: %v", err)
	}
	if GetType(&v) != TRUE {
		t.Errorf("期望类型为TRUE，但得到: %v", GetType(&v))
	}
}

// 测试解析false值
func TestParseFalse(t *testing.T) {
	v := Value{Type: TRUE} // 初始化为非FALSE类型
	if err := Parse(&v, "false"); err != PARSE_OK {
		t.Errorf("期望解析成功，但返回错误: %v", err)
	}
	if GetType(&v) != FALSE {
		t.Errorf("期望类型为FALSE，但得到: %v", GetType(&v))
	}
}

// 测试解析数字
func TestParseNumber(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
	}{
		{"0", 0.0},
		{"-0", 0.0},
		{"-0.0", 0.0},
		{"1", 1.0},
		{"-1", -1.0},
		{"1.5", 1.5},
		{"-1.5", -1.5},
		{"3.1416", 3.1416},
		{"1E10", 1e10},
		{"1e10", 1e10},
		{"1E+10", 1e+10},
		{"1E-10", 1e-10},
		{"-1E10", -1e10},
		{"-1e10", -1e10},
		{"-1E+10", -1e+10},
		{"-1E-10", -1e-10},
		{"1.234E+10", 1.234e+10},
		{"1.234E-10", 1.234e-10},
		{"1e-10000", 0.0}, // 下溢出
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, tc.input); err != PARSE_OK {
				t.Errorf("期望解析成功，但返回错误: %v", err)
			}
			if GetType(&v) != NUMBER {
				t.Errorf("期望类型为NUMBER，但得到: %v", GetType(&v))
			}
			if GetNumber(&v) != tc.expected {
				t.Errorf("期望值为%v，但得到: %v", tc.expected, GetNumber(&v))
			}
		})
	}
}

// 测试解析字符串
func TestParseString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{`""`, ""},
		{`"Hello"`, "Hello"},
		{`"Hello\nWorld"`, "Hello\nWorld"},
		{`"\"\\\/\b\f\n\r\t"`, "\"\\/\b\f\n\r\t"},
		{`"Hello\u0000World"`, "Hello\u0000World"},
		{`"\u0024"`, "$"},                // U+0024 是 $
		{`"\u00A2"`, "\u00A2"},           // U+00A2 是 ¢
		{`"\u20AC"`, "\u20AC"},           // U+20AC 是 €
		{`"\uD834\uDD1E"`, "\U0001D11E"}, // U+1D11E 是 𝄞
		{`"\ud834\udd1e"`, "\U0001D11E"}, // U+1D11E 是 𝄞
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, tc.input); err != PARSE_OK {
				t.Errorf("期望解析成功，但返回错误: %v", err)
			}
			if GetType(&v) != STRING {
				t.Errorf("期望类型为STRING，但得到: %v", GetType(&v))
			}
			if GetString(&v) != tc.expected {
				t.Errorf("期望值为%q，但得到: %q", tc.expected, GetString(&v))
			}
		})
	}
}

// 测试解析数组
func TestParseArray(t *testing.T) {
	t.Run("EmptyArray", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[ ]"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != ARRAY {
			t.Errorf("期望类型为ARRAY，但得到: %v", GetType(&v))
		}
		if GetArraySize(&v) != 0 {
			t.Errorf("期望数组大小为0，但得到: %v", GetArraySize(&v))
		}
	})

	t.Run("OneElement", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[null]"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != ARRAY {
			t.Errorf("期望类型为ARRAY，但得到: %v", GetType(&v))
		}
		if GetArraySize(&v) != 1 {
			t.Errorf("期望数组大小为1，但得到: %v", GetArraySize(&v))
		}
		if GetType(GetArrayElement(&v, 0)) != NULL {
			t.Errorf("期望第一个元素为NULL，但得到: %v", GetType(GetArrayElement(&v, 0)))
		}
	})

	t.Run("MultipleElements", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[null, false, true, 123, \"abc\"]"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != ARRAY {
			t.Errorf("期望类型为ARRAY，但得到: %v", GetType(&v))
		}
		if GetArraySize(&v) != 5 {
			t.Errorf("期望数组大小为5，但得到: %v", GetArraySize(&v))
		}
		if GetType(GetArrayElement(&v, 0)) != NULL {
			t.Errorf("期望第一个元素为NULL，但得到: %v", GetType(GetArrayElement(&v, 0)))
		}
		if GetType(GetArrayElement(&v, 1)) != FALSE {
			t.Errorf("期望第二个元素为FALSE，但得到: %v", GetType(GetArrayElement(&v, 1)))
		}
		if GetType(GetArrayElement(&v, 2)) != TRUE {
			t.Errorf("期望第三个元素为TRUE，但得到: %v", GetType(GetArrayElement(&v, 2)))
		}
		if GetType(GetArrayElement(&v, 3)) != NUMBER || GetNumber(GetArrayElement(&v, 3)) != 123.0 {
			t.Errorf("期望第四个元素为数字123，但得到: %v", GetArrayElement(&v, 3))
		}
		if GetType(GetArrayElement(&v, 4)) != STRING || GetString(GetArrayElement(&v, 4)) != "abc" {
			t.Errorf("期望第五个元素为字符串\"abc\"，但得到: %v", GetArrayElement(&v, 4))
		}
	})

	t.Run("NestedArray", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[[1, 2], [3, [4, 5]], 6]"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != ARRAY {
			t.Errorf("期望类型为ARRAY，但得到: %v", GetType(&v))
		}
		if GetArraySize(&v) != 3 {
			t.Errorf("期望数组大小为3，但得到: %v", GetArraySize(&v))
		}

		// 验证第一个元素 [1, 2]
		element0 := GetArrayElement(&v, 0)
		if GetType(element0) != ARRAY || GetArraySize(element0) != 2 {
			t.Errorf("期望第一个元素为大小为2的数组，但得到: %v", element0)
		}
		if GetType(GetArrayElement(element0, 0)) != NUMBER || GetNumber(GetArrayElement(element0, 0)) != 1.0 {
			t.Errorf("期望[1, 2]的第一个元素为数字1，但得到: %v", GetArrayElement(element0, 0))
		}
		if GetType(GetArrayElement(element0, 1)) != NUMBER || GetNumber(GetArrayElement(element0, 1)) != 2.0 {
			t.Errorf("期望[1, 2]的第二个元素为数字2，但得到: %v", GetArrayElement(element0, 1))
		}

		// 验证第二个元素 [3, [4, 5]]
		element1 := GetArrayElement(&v, 1)
		if GetType(element1) != ARRAY || GetArraySize(element1) != 2 {
			t.Errorf("期望第二个元素为大小为2的数组，但得到: %v", element1)
		}
		if GetType(GetArrayElement(element1, 0)) != NUMBER || GetNumber(GetArrayElement(element1, 0)) != 3.0 {
			t.Errorf("期望[3, [4, 5]]的第一个元素为数字3，但得到: %v", GetArrayElement(element1, 0))
		}

		// 验证嵌套数组 [4, 5]
		element2 := GetArrayElement(element1, 1)
		if GetType(element2) != ARRAY || GetArraySize(element2) != 2 {
			t.Errorf("期望[3, [4, 5]]的第二个元素为大小为2的数组，但得到: %v", element2)
		}
		if GetType(GetArrayElement(element2, 0)) != NUMBER || GetNumber(GetArrayElement(element2, 0)) != 4.0 {
			t.Errorf("期望[4, 5]的第一个元素为数字4，但得到: %v", GetArrayElement(element2, 0))
		}
		if GetType(GetArrayElement(element2, 1)) != NUMBER || GetNumber(GetArrayElement(element2, 1)) != 5.0 {
			t.Errorf("期望[4, 5]的第二个元素为数字5，但得到: %v", GetArrayElement(element2, 1))
		}

		// 验证第三个元素 6
		if GetType(GetArrayElement(&v, 2)) != NUMBER || GetNumber(GetArrayElement(&v, 2)) != 6.0 {
			t.Errorf("期望第三个元素为数字6，但得到: %v", GetArrayElement(&v, 2))
		}
	})
}

// 测试解析对象
func TestParseObject(t *testing.T) {
	t.Run("EmptyObject", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{ }"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT {
			t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
		}
		if GetObjectSize(&v) != 0 {
			t.Errorf("期望对象大小为0，但得到: %v", GetObjectSize(&v))
		}
	})

	t.Run("OneKeyValue", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\":\"value\"}"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT {
			t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
		}
		if GetObjectSize(&v) != 1 {
			t.Errorf("期望对象大小为1，但得到: %v", GetObjectSize(&v))
		}
		if GetObjectKey(&v, 0) != "name" {
			t.Errorf("期望第一个键为\"name\"，但得到: %v", GetObjectKey(&v, 0))
		}
		if GetType(GetObjectValue(&v, 0)) != STRING || GetString(GetObjectValue(&v, 0)) != "value" {
			t.Errorf("期望第一个值为字符串\"value\"，但得到: %v", GetObjectValue(&v, 0))
		}
	})

	t.Run("MultipleKeyValues", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\":\"value\", \"age\":30, \"isStudent\":false}"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT {
			t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
		}
		if GetObjectSize(&v) != 3 {
			t.Errorf("期望对象大小为3，但得到: %v", GetObjectSize(&v))
		}

		// 验证第一个键值对
		if GetObjectKey(&v, 0) != "name" {
			t.Errorf("期望第一个键为\"name\"，但得到: %v", GetObjectKey(&v, 0))
		}
		if GetType(GetObjectValue(&v, 0)) != STRING || GetString(GetObjectValue(&v, 0)) != "value" {
			t.Errorf("期望第一个值为字符串\"value\"，但得到: %v", GetObjectValue(&v, 0))
		}

		// 验证第二个键值对
		if GetObjectKey(&v, 1) != "age" {
			t.Errorf("期望第二个键为\"age\"，但得到: %v", GetObjectKey(&v, 1))
		}
		if GetType(GetObjectValue(&v, 1)) != NUMBER || GetNumber(GetObjectValue(&v, 1)) != 30.0 {
			t.Errorf("期望第二个值为数字30，但得到: %v", GetObjectValue(&v, 1))
		}

		// 验证第三个键值对
		if GetObjectKey(&v, 2) != "isStudent" {
			t.Errorf("期望第三个键为\"isStudent\"，但得到: %v", GetObjectKey(&v, 2))
		}
		if GetType(GetObjectValue(&v, 2)) != FALSE {
			t.Errorf("期望第三个值为FALSE，但得到: %v", GetObjectValue(&v, 2))
		}
	})

	t.Run("NestedObject", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"person\":{\"name\":\"John\", \"age\":30}, \"isActive\":true}"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT {
			t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
		}
		if GetObjectSize(&v) != 2 {
			t.Errorf("期望对象大小为2，但得到: %v", GetObjectSize(&v))
		}

		// 验证第一个键值对 "person":{...}
		if GetObjectKey(&v, 0) != "person" {
			t.Errorf("期望第一个键为\"person\"，但得到: %v", GetObjectKey(&v, 0))
		}
		personObj := GetObjectValue(&v, 0)
		if GetType(personObj) != OBJECT {
			t.Errorf("期望第一个值为OBJECT，但得到: %v", GetType(personObj))
		}
		if GetObjectSize(personObj) != 2 {
			t.Errorf("期望person对象大小为2，但得到: %v", GetObjectSize(personObj))
		}

		// 验证嵌套对象的键值对
		if GetObjectKey(personObj, 0) != "name" {
			t.Errorf("期望person对象第一个键为\"name\"，但得到: %v", GetObjectKey(personObj, 0))
		}
		if GetType(GetObjectValue(personObj, 0)) != STRING || GetString(GetObjectValue(personObj, 0)) != "John" {
			t.Errorf("期望person对象第一个值为字符串\"John\"，但得到: %v", GetObjectValue(personObj, 0))
		}
		if GetObjectKey(personObj, 1) != "age" {
			t.Errorf("期望person对象第二个键为\"age\"，但得到: %v", GetObjectKey(personObj, 1))
		}
		if GetType(GetObjectValue(personObj, 1)) != NUMBER || GetNumber(GetObjectValue(personObj, 1)) != 30.0 {
			t.Errorf("期望person对象第二个值为数字30，但得到: %v", GetObjectValue(personObj, 1))
		}

		// 验证第二个键值对 "isActive":true
		if GetObjectKey(&v, 1) != "isActive" {
			t.Errorf("期望第二个键为\"isActive\"，但得到: %v", GetObjectKey(&v, 1))
		}
		if GetType(GetObjectValue(&v, 1)) != TRUE {
			t.Errorf("期望第二个值为TRUE，但得到: %v", GetObjectValue(&v, 1))
		}
	})

	t.Run("ComplexObject", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\":\"John\", \"age\":30, \"address\":{\"city\":\"New York\", \"zip\":\"10001\"}, \"hobbies\":[\"reading\", \"gaming\", {\"sport\":\"football\"}]}"); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT {
			t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
		}
		if GetObjectSize(&v) != 4 {
			t.Errorf("期望对象大小为4，但得到: %v", GetObjectSize(&v))
		}

		// 验证基本键值对
		if GetObjectKey(&v, 0) != "name" || GetType(GetObjectValue(&v, 0)) != STRING || GetString(GetObjectValue(&v, 0)) != "John" {
			t.Errorf("期望第一个键值对为\"name\":\"John\"，但得到: %v:%v", GetObjectKey(&v, 0), GetObjectValue(&v, 0))
		}
		if GetObjectKey(&v, 1) != "age" || GetType(GetObjectValue(&v, 1)) != NUMBER || GetNumber(GetObjectValue(&v, 1)) != 30.0 {
			t.Errorf("期望第二个键值对为\"age\":30，但得到: %v:%v", GetObjectKey(&v, 1), GetObjectValue(&v, 1))
		}

		// 验证嵌套对象
		if GetObjectKey(&v, 2) != "address" {
			t.Errorf("期望第三个键为\"address\"，但得到: %v", GetObjectKey(&v, 2))
		}
		addressObj := GetObjectValue(&v, 2)
		if GetType(addressObj) != OBJECT {
			t.Errorf("期望第三个值为OBJECT，但得到: %v", GetType(addressObj))
		}
		if GetObjectSize(addressObj) != 2 {
			t.Errorf("期望address对象大小为2，但得到: %v", GetObjectSize(addressObj))
		}
		if GetObjectKey(addressObj, 0) != "city" || GetString(GetObjectValue(addressObj, 0)) != "New York" {
			t.Errorf("期望address对象第一个键值对为\"city\":\"New York\"，但得到: %v:%v", GetObjectKey(addressObj, 0), GetObjectValue(addressObj, 0))
		}
		if GetObjectKey(addressObj, 1) != "zip" || GetString(GetObjectValue(addressObj, 1)) != "10001" {
			t.Errorf("期望address对象第二个键值对为\"zip\":\"10001\"，但得到: %v:%v", GetObjectKey(addressObj, 1), GetObjectValue(addressObj, 1))
		}

		// 验证数组值
		if GetObjectKey(&v, 3) != "hobbies" {
			t.Errorf("期望第四个键为\"hobbies\"，但得到: %v", GetObjectKey(&v, 3))
		}
		hobbiesArr := GetObjectValue(&v, 3)
		if GetType(hobbiesArr) != ARRAY {
			t.Errorf("期望第四个值为ARRAY，但得到: %v", GetType(hobbiesArr))
		}
		if GetArraySize(hobbiesArr) != 3 {
			t.Errorf("期望hobbies数组大小为3，但得到: %v", GetArraySize(hobbiesArr))
		}
		if GetType(GetArrayElement(hobbiesArr, 0)) != STRING || GetString(GetArrayElement(hobbiesArr, 0)) != "reading" {
			t.Errorf("期望hobbies数组第一个元素为\"reading\"，但得到: %v", GetArrayElement(hobbiesArr, 0))
		}
		if GetType(GetArrayElement(hobbiesArr, 1)) != STRING || GetString(GetArrayElement(hobbiesArr, 1)) != "gaming" {
			t.Errorf("期望hobbies数组第二个元素为\"gaming\"，但得到: %v", GetArrayElement(hobbiesArr, 1))
		}

		// 验证数组中的对象
		sportObj := GetArrayElement(hobbiesArr, 2)
		if GetType(sportObj) != OBJECT {
			t.Errorf("期望hobbies数组第三个元素为OBJECT，但得到: %v", GetType(sportObj))
		}
		if GetObjectSize(sportObj) != 1 {
			t.Errorf("期望sport对象大小为1，但得到: %v", GetObjectSize(sportObj))
		}
		if GetObjectKey(sportObj, 0) != "sport" || GetString(GetObjectValue(sportObj, 0)) != "football" {
			t.Errorf("期望sport对象键值对为\"sport\":\"football\"，但得到: %v:%v", GetObjectKey(sportObj, 0), GetObjectValue(sportObj, 0))
		}
	})
}

// 测试解析对象错误
func TestParseObjectError(t *testing.T) {
	// 测试缺少右花括号
	t.Run("MissingRightBrace", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\": \"value\""); err != PARSE_MISS_COMMA_OR_CURLY_BRACKET {
			t.Errorf("期望错误PARSE_MISS_COMMA_OR_CURLY_BRACKET，但得到: %v", err)
		}
	})

	// 测试缺少键
	t.Run("MissingKey", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{:1}"); err != PARSE_MISS_KEY {
			t.Errorf("期望错误PARSE_MISS_KEY，但得到: %v", err)
		}
	})

	// 测试键不是字符串
	t.Run("KeyNotString", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{1:1}"); err != PARSE_MISS_KEY {
			t.Errorf("期望错误PARSE_MISS_KEY，但得到: %v", err)
		}
	})

	// 测试缺少冒号
	t.Run("MissingColon", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\" 1}"); err != PARSE_MISS_COLON {
			t.Errorf("期望错误PARSE_MISS_COLON，但得到: %v", err)
		}
	})

	// 测试缺少逗号
	t.Run("MissingComma", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\":\"value\" \"age\":30}"); err != PARSE_MISS_COMMA_OR_CURLY_BRACKET {
			t.Errorf("期望错误PARSE_MISS_COMMA_OR_CURLY_BRACKET，但得到: %v", err)
		}
	})

	// 测试对象中的无效值
	t.Run("InvalidValue", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\":?}"); err != PARSE_INVALID_VALUE {
			t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v", err)
		}
	})

	// 测试对象后有额外内容
	t.Run("ExtraContent", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "{\"name\":\"value\"} null"); err != PARSE_ROOT_NOT_SINGULAR {
			t.Errorf("期望错误PARSE_ROOT_NOT_SINGULAR，但得到: %v", err)
		}
	})
}

// 测试查找对象成员
func TestFindObjectMember(t *testing.T) {
	v := Value{}
	if err := Parse(&v, "{\"name\":\"John\", \"age\":30, \"gender\":\"male\"}"); err != PARSE_OK {
		t.Errorf("期望解析成功，但返回错误: %v", err)
	}

	// 测试查找存在的键
	if idx := FindObjectIndex(&v, "name"); idx != 0 {
		t.Errorf("期望查找\"name\"的索引为0，但得到: %v", idx)
	}
	if idx := FindObjectIndex(&v, "age"); idx != 1 {
		t.Errorf("期望查找\"age\"的索引为1，但得到: %v", idx)
	}
	if idx := FindObjectIndex(&v, "gender"); idx != 2 {
		t.Errorf("期望查找\"gender\"的索引为2，但得到: %v", idx)
	}

	// 测试查找不存在的键
	if idx := FindObjectIndex(&v, "address"); idx != -1 {
		t.Errorf("期望查找不存在的键返回-1，但得到: %v", idx)
	}

	// 测试根据键获取值
	nameValue := GetObjectValueByKey(&v, "name")
	if nameValue == nil || GetType(nameValue) != STRING || GetString(nameValue) != "John" {
		t.Errorf("期望获取\"name\"的值为\"John\"，但得到: %v", nameValue)
	}

	ageValue := GetObjectValueByKey(&v, "age")
	if ageValue == nil || GetType(ageValue) != NUMBER || GetNumber(ageValue) != 30.0 {
		t.Errorf("期望获取\"age\"的值为30，但得到: %v", ageValue)
	}

	// 测试获取不存在的键的值
	if value := GetObjectValueByKey(&v, "address"); value != nil {
		t.Errorf("期望获取不存在的键的值为nil，但得到: %v", value)
	}
}

// 测试解析数组错误
// 测试解析数组错误
func TestParseArrayError(t *testing.T) {
	// 测试缺少右方括号
	t.Run("MissingRightBracket", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[1, 2"); err != PARSE_MISS_COMMA_OR_SQUARE_BRACKET {
			t.Errorf("期望错误PARSE_MISS_COMMA_OR_SQUARE_BRACKET，但得到: %v", err)
		}
	})

	// 测试数组中的无效值
	t.Run("InvalidValue", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[1, ?]"); err != PARSE_INVALID_VALUE {
			t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v", err)
		}
	})

	// 测试缺少逗号
	t.Run("MissingComma", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[1 2]"); err != PARSE_MISS_COMMA_OR_SQUARE_BRACKET {
			t.Errorf("期望错误PARSE_MISS_COMMA_OR_SQUARE_BRACKET，但得到: %v", err)
		}
	})

	// 测试空数组后有额外内容
	t.Run("ExtraContent", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "[] null"); err != PARSE_ROOT_NOT_SINGULAR {
			t.Errorf("期望错误PARSE_ROOT_NOT_SINGULAR，但得到: %v", err)
		}
	})
}

// 测试解析期望值
func TestParseExpectValue(t *testing.T) {
	v := Value{Type: FALSE}
	if err := Parse(&v, ""); err != PARSE_EXPECT_VALUE {
		t.Errorf("期望错误PARSE_EXPECT_VALUE，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	v = Value{Type: FALSE}
	if err := Parse(&v, " "); err != PARSE_EXPECT_VALUE {
		t.Errorf("期望错误PARSE_EXPECT_VALUE，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}

// 测试解析无效值
func TestParseInvalidValue(t *testing.T) {
	v := Value{Type: FALSE}
	if err := Parse(&v, "nul"); err != PARSE_INVALID_VALUE {
		t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	v = Value{Type: FALSE}
	if err := Parse(&v, "?"); err != PARSE_INVALID_VALUE {
		t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}

// 测试解析无效字符串
func TestParseInvalidString(t *testing.T) {
	// 测试缺少引号
	t.Run("MissingQuotationMark", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "\""); err != PARSE_MISS_QUOTATION_MARK {
			t.Errorf("期望错误PARSE_MISS_QUOTATION_MARK，但得到: %v", err)
		}
	})

	// 测试无效的转义序列
	t.Run("InvalidEscapeSequence", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "\"\\v\""); err != PARSE_INVALID_STRING_ESCAPE {
			t.Errorf("期望错误PARSE_INVALID_STRING_ESCAPE，但得到: %v", err)
		}
	})

	// 测试无效的字符
	t.Run("InvalidChar", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "\"\x01\""); err != PARSE_INVALID_STRING_CHAR {
			t.Errorf("期望错误PARSE_INVALID_STRING_CHAR，但得到: %v", err)
		}
	})

	// 测试无效的Unicode十六进制
	t.Run("InvalidUnicodeHex", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "\"\\u123\""); err != PARSE_INVALID_UNICODE_HEX {
			t.Errorf("期望错误PARSE_INVALID_UNICODE_HEX，但得到: %v", err)
		}
	})

	// 测试无效的Unicode代理对
	t.Run("InvalidUnicodeSurrogate", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "\"\\uD800\""); err != PARSE_INVALID_UNICODE_SURROGATE {
			t.Errorf("期望错误PARSE_INVALID_UNICODE_SURROGATE，但得到: %v", err)
		}
		v = Value{}
		if err := Parse(&v, "\"\\uD800\\uDBFF\""); err != PARSE_INVALID_UNICODE_SURROGATE {
			t.Errorf("期望错误PARSE_INVALID_UNICODE_SURROGATE，但得到: %v", err)
		}
	})
}

// 测试解析数字太大
func TestParseNumberTooBig(t *testing.T) {
	v := Value{Type: FALSE}
	if err := Parse(&v, "1e309"); err != PARSE_NUMBER_TOO_BIG {
		t.Errorf("期望错误PARSE_NUMBER_TOO_BIG，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	v = Value{Type: FALSE}
	if err := Parse(&v, "-1e309"); err != PARSE_NUMBER_TOO_BIG {
		t.Errorf("期望错误PARSE_NUMBER_TOO_BIG，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}

// 测试解析根节点不唯一
func TestParseRootNotSingular(t *testing.T) {
	v := Value{Type: FALSE}
	if err := Parse(&v, "null x"); err != PARSE_ROOT_NOT_SINGULAR {
		t.Errorf("期望错误PARSE_ROOT_NOT_SINGULAR，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	// 测试数字后有额外内容
	v = Value{Type: FALSE}
	if err := Parse(&v, "0123"); err != PARSE_INVALID_VALUE {
		t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	v = Value{Type: FALSE}
	if err := Parse(&v, "0x0"); err != PARSE_INVALID_VALUE {
		t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}

// 运行所有测试
func TestParse(t *testing.T) {
	t.Run("TestParseNull", TestParseNull)
	t.Run("TestParseTrue", TestParseTrue)
	t.Run("TestParseFalse", TestParseFalse)
	t.Run("TestParseNumber", TestParseNumber)
	t.Run("TestParseString", TestParseString)
	t.Run("TestParseArray", TestParseArray)
	t.Run("TestParseObject", TestParseObject)
	t.Run("TestParseArrayError", TestParseArrayError)
	t.Run("TestParseObjectError", TestParseObjectError)
	t.Run("TestFindObjectMember", TestFindObjectMember)
	t.Run("TestParseExpectValue", TestParseExpectValue)
	t.Run("TestParseInvalidValue", TestParseInvalidValue)
	t.Run("TestParseInvalidString", TestParseInvalidString)
	t.Run("TestParseNumberTooBig", TestParseNumberTooBig)
	t.Run("TestParseRootNotSingular", TestParseRootNotSingular)
}

// 基准测试 - 解析null值
func BenchmarkParseNull(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "null")
	}
}

// 基准测试 - 解析true值
func BenchmarkParseTrue(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "true")
	}
}

// 基准测试 - 解析false值
func BenchmarkParseFalse(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "false")
	}
}

// 基准测试 - 解析数字
func BenchmarkParseNumber(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "123.456e+789")
	}
}

// 基准测试 - 解析字符串
func BenchmarkParseString(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "\"Hello\\nWorld\"")
	}
}

// 基准测试 - 解析Unicode字符串
func BenchmarkParseUnicodeString(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "\"\\uD834\\uDD1E\"")
	}
}

// 基准测试 - 解析数组
func BenchmarkParseArray(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "[null,false,true,123,\"abc\",[1,2,3]]")
	}
}

// 基准测试 - 解析对象
func BenchmarkParseObject(b *testing.B) {
	v := Value{}
	for i := 0; i < b.N; i++ {
		Parse(&v, "{\"name\":\"value\",\"age\":30,\"address\":{\"city\":\"New York\"}}")
	}
}

// 示例 - 解析null值
func ExampleParse_null() {
	v := Value{}
	Parse(&v, "null")
	fmt.Println(v.String())
	// Output: null
}

// 示例 - 解析true值
func ExampleParse_true() {
	v := Value{}
	Parse(&v, "true")
	fmt.Println(v.String())
	// Output: true
}

// 示例 - 解析false值
func ExampleParse_false() {
	v := Value{}
	Parse(&v, "false")
	fmt.Println(v.String())
	// Output: false
}

// 示例 - 解析数字
func ExampleParse_number() {
	v := Value{}
	Parse(&v, "123.456")
	fmt.Println(v.String())
	// Output: 123.456
}

// 示例 - 解析字符串
func ExampleParse_string() {
	v := Value{}
	Parse(&v, "\"Hello, World!\"")
	fmt.Println(v.String())
	// Output: "Hello, World!"
}

// 示例 - 解析数组
func ExampleParse_array() {
	v := Value{}
	Parse(&v, "[1,2,3]")
	fmt.Println(v.String())
	// Output: [1,2,3]
}

// 示例 - 解析嵌套数组
func ExampleParse_nestedArray() {
	v := Value{}
	Parse(&v, "[[1,2],[3,4],5]")
	fmt.Println(v.String())
	// Output: [[1,2],[3,4],5]
}

// 示例 - 解析对象
func ExampleParse_object() {
	v := Value{}
	Parse(&v, "{\"name\":\"John\",\"age\":30}")
	fmt.Println(v.String())
	// Output: {"name":"John","age":30}
}

// 示例 - 解析嵌套对象
func ExampleParse_nestedObject() {
	v := Value{}
	Parse(&v, "{\"person\":{\"name\":\"John\",\"age\":30},\"isActive\":true}")
	fmt.Println(v.String())
	// Output: {"person":{"name":"John","age":30},"isActive":true}
}

// 示例 - 错误处理
func ExampleParse_error() {
	v := Value{}
	err := Parse(&v, "[1,2,")
	fmt.Println(err)
	// Output: 缺少逗号或方括号
}

func ExampleParse_error02() {
	v := Value{}
	err := Parse(&v, "{\"name\":\"John\",\"age\":")
	fmt.Println(err)
	// Output: 期望一个值
}

// 测试Stringify函数：往返测试
func testRoundTrip(t *testing.T, json string) {
	var v Value
	err := Parse(&v, json)
	if err != PARSE_OK {
		t.Errorf("解析失败，错误：%v，JSON：%s", err, json)
		return
	}

	jsonOut, errStr := Stringify(&v)
	if errStr != STRINGIFY_OK {
		t.Errorf("字符串化失败，错误：%v，JSON：%s", errStr, json)
		return
	}

	var v2 Value
	err = Parse(&v2, jsonOut)
	if err != PARSE_OK {
		t.Errorf("重新解析失败，错误：%v，JSON：%s，字符串化结果：%s", err, json, jsonOut)
		return
	}

	// 不直接比较字符串，因为数字和空白格式可能不同
	// 但结构应该是相同的
	if ok, path := compareValue(&v, &v2); !ok {
		t.Errorf("往返测试失败，路径：%s，原始JSON：%s，字符串化结果：%s", path, json, jsonOut)
	}
}

// 比较两个Value是否相等，返回(是否相等, 不同的路径)
func compareValue(v1, v2 *Value) (bool, string) {
	if v1.Type != v2.Type {
		return false, "type"
	}

	switch v1.Type {
	case NULL:
		return true, ""
	case FALSE, TRUE:
		return true, ""
	case NUMBER:
		// 允许数字有小误差
		if v1.N != v2.N {
			// 对于非常接近的数字，考虑它们相等
			if (v1.N == 0.0 && v2.N == 0.0) || (v1.N != 0.0 && v2.N != 0.0 &&
				(v1.N-v2.N)/v1.N < 1e-15) {
				return true, ""
			}
			return false, "number"
		}
		return true, ""
	case STRING:
		if v1.S != v2.S {
			return false, "string"
		}
		return true, ""
	case ARRAY:
		if len(v1.A) != len(v2.A) {
			return false, "array.length"
		}
		for i := 0; i < len(v1.A); i++ {
			if ok, path := compareValue(v1.A[i], v2.A[i]); !ok {
				return false, fmt.Sprintf("array[%d].%s", i, path)
			}
		}
		return true, ""
	case OBJECT:
		if len(v1.O) != len(v2.O) {
			return false, "object.length"
		}
		// 由于对象是无序的，我们需要通过键来查找成员
		for _, m1 := range v1.O {
			j := -1
			for k, m2 := range v2.O {
				if m1.K == m2.K {
					j = k
					break
				}
			}
			if j == -1 {
				return false, fmt.Sprintf("object.missing key %s", m1.K)
			}
			if ok, path := compareValue(m1.V, v2.O[j].V); !ok {
				return false, fmt.Sprintf("object[%s].%s", m1.K, path)
			}
		}
		return true, ""
	default:
		return false, "unknown type"
	}
}

// TestStringify 测试Stringify函数
func TestStringify(t *testing.T) {
	t.Run("Null", func(t *testing.T) {
		testRoundTrip(t, "null")
	})

	t.Run("Boolean", func(t *testing.T) {
		testRoundTrip(t, "true")
		testRoundTrip(t, "false")
	})

	t.Run("Number", func(t *testing.T) {
		testRoundTrip(t, "0")
		testRoundTrip(t, "1")
		testRoundTrip(t, "-1")
		testRoundTrip(t, "1.5")
		testRoundTrip(t, "-1.5")
		testRoundTrip(t, "3.1416")
		testRoundTrip(t, "1e20")
		testRoundTrip(t, "1.234e+10")
		testRoundTrip(t, "1.234e-10")
	})

	t.Run("String", func(t *testing.T) {
		testRoundTrip(t, `""`)
		testRoundTrip(t, `"Hello"`)
		testRoundTrip(t, `"Hello\nWorld"`)
		testRoundTrip(t, `"\"\\\/\b\f\n\r\t"`)
		testRoundTrip(t, `"\u0024"`)       // $
		testRoundTrip(t, `"\u00A2"`)       // ¢
		testRoundTrip(t, `"\u20AC"`)       // €
		testRoundTrip(t, `"\uD834\uDD1E"`) // 𝄞
	})

	t.Run("Array", func(t *testing.T) {
		testRoundTrip(t, `[]`)
		testRoundTrip(t, `[null]`)
		testRoundTrip(t, `[null,false,true,123,"abc"]`)
		testRoundTrip(t, `[[1,2],[3,[4,5]],6]`)
	})

	t.Run("Object", func(t *testing.T) {
		testRoundTrip(t, `{}`)
		testRoundTrip(t, `{"foo":"bar"}`)
		testRoundTrip(t, `{"a":1,"b":true,"c":["hello"]}`)
		testRoundTrip(t, `{"a":{"b":{"c":{"d":123}}}}`)
	})

	t.Run("Complex", func(t *testing.T) {
		testRoundTrip(t, `{"a":[1,2],"b":{"c":0,"d":[false,true,null,"Hello"]}}`)
	})
}

// TestStringifySpecific 测试Stringify函数特定行为
func TestStringifySpecific(t *testing.T) {
	t.Run("StringEscaping", func(t *testing.T) {
		v := Value{Type: STRING, S: "\"\\/\b\f\n\r\t"}
		json, err := Stringify(&v)
		if err != STRINGIFY_OK {
			t.Errorf("字符串化失败，错误：%v", err)
			return
		}
		expected := `"\"\\\/\b\f\n\r\t"`
		if json != expected {
			t.Errorf("字符串转义不正确，期望：%s，得到：%s", expected, json)
		}
	})

	t.Run("ControlChars", func(t *testing.T) {
		v := Value{Type: STRING, S: string([]byte{0x01, 0x1F})}
		json, err := Stringify(&v)
		if err != STRINGIFY_OK {
			t.Errorf("字符串化失败，错误：%v", err)
			return
		}
		expected := `"\u0001\u001f"`
		if json != expected {
			t.Errorf("控制字符转义不正确，期望：%s，得到：%s", expected, json)
		}
	})

	// 测试输出数字的格式
	t.Run("NumberFormat", func(t *testing.T) {
		numbers := []struct {
			value    float64
			expected string
		}{
			{0.0, "0"},
			{1.0, "1"},
			{-1.0, "-1"},
			{1.5, "1.5"},
			{-1.5, "-1.5"},
			{1e20, "1e+20"},
			{1.234e10, "1.234e+10"},
			{1.234e-10, "1.234e-10"},
		}

		for _, n := range numbers {
			v := Value{Type: NUMBER, N: n.value}
			json, err := Stringify(&v)
			if err != STRINGIFY_OK {
				t.Errorf("字符串化失败，错误：%v，值：%f", err, n.value)
				continue
			}
			// 注意：由于浮点数格式差异，我们不直接比较字符串
			// 而是解析回来比较值
			var v2 Value
			if err := Parse(&v2, json); err != PARSE_OK {
				t.Errorf("解析失败，错误：%v，字符串：%s", err, json)
				continue
			}
			if v2.Type != NUMBER || math.Abs(v2.N-n.value)/math.Max(1.0, math.Abs(n.value)) > 1e-15 {
				t.Errorf("数字往返测试失败，原始值：%f，得到：%f，字符串：%s", n.value, v2.N, json)
			}
		}
	})
}

// BenchmarkStringifyNull 测试生成null值的性能
func BenchmarkStringifyNull(b *testing.B) {
	v := Value{Type: NULL}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyTrue 测试生成true值的性能
func BenchmarkStringifyTrue(b *testing.B) {
	v := Value{Type: TRUE}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyFalse 测试生成false值的性能
func BenchmarkStringifyFalse(b *testing.B) {
	v := Value{Type: FALSE}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyNumber 测试生成数字值的性能
func BenchmarkStringifyNumber(b *testing.B) {
	v := Value{Type: NUMBER, N: 123.456}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyString 测试生成字符串值的性能
func BenchmarkStringifyString(b *testing.B) {
	v := Value{Type: STRING, S: "Hello, World!"}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyArray 测试生成数组值的性能
func BenchmarkStringifyArray(b *testing.B) {
	v := Value{
		Type: ARRAY,
		A: []*Value{
			{Type: NULL},
			{Type: FALSE},
			{Type: TRUE},
			{Type: NUMBER, N: 123},
			{Type: STRING, S: "abc"},
		},
	}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyObject 测试生成对象值的性能
func BenchmarkStringifyObject(b *testing.B) {
	v := Value{
		Type: OBJECT,
		O: []Member{
			{K: "null", V: &Value{Type: NULL}},
			{K: "false", V: &Value{Type: FALSE}},
			{K: "true", V: &Value{Type: TRUE}},
			{K: "number", V: &Value{Type: NUMBER, N: 123}},
			{K: "string", V: &Value{Type: STRING, S: "abc"}},
		},
	}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// BenchmarkStringifyComplex 测试生成复杂值的性能
func BenchmarkStringifyComplex(b *testing.B) {
	v := Value{
		Type: OBJECT,
		O: []Member{
			{K: "a", V: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 2},
				},
			}},
			{K: "b", V: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "c", V: &Value{Type: NUMBER, N: 0}},
					{K: "d", V: &Value{
						Type: ARRAY,
						A: []*Value{
							{Type: FALSE},
							{Type: TRUE},
							{Type: NULL},
							{Type: STRING, S: "Hello"},
						},
					}},
				},
			}},
		},
	}
	for i := 0; i < b.N; i++ {
		_, _ = Stringify(&v)
	}
}

// ExampleStringify_null 展示如何字符串化null值
func ExampleStringify_null() {
	v := Value{Type: NULL}
	json, _ := Stringify(&v)
	fmt.Println(json)
	// Output: null
}

// ExampleStringify_boolean 展示如何字符串化布尔值
func ExampleStringify_boolean() {
	v1 := Value{Type: TRUE}
	json1, _ := Stringify(&v1)
	fmt.Println(json1)

	v2 := Value{Type: FALSE}
	json2, _ := Stringify(&v2)
	fmt.Println(json2)
	// Output: true
	// false
}

// ExampleStringify_number 展示如何字符串化数字值
func ExampleStringify_number() {
	v := Value{Type: NUMBER, N: 123.456}
	json, _ := Stringify(&v)
	fmt.Println(json)
	// Output: 123.456
}

// ExampleStringify_string 展示如何字符串化字符串值
func ExampleStringify_string() {
	v := Value{Type: STRING, S: "Hello, World!"}
	json, _ := Stringify(&v)
	fmt.Println(json)
	// Output: "Hello, World!"
}

// ExampleStringify_array 展示如何字符串化数组值
func ExampleStringify_array() {
	v := Value{
		Type: ARRAY,
		A: []*Value{
			{Type: NULL},
			{Type: FALSE},
			{Type: TRUE},
			{Type: NUMBER, N: 123},
			{Type: STRING, S: "abc"},
		},
	}
	json, _ := Stringify(&v)
	fmt.Println(json)
	// Output: [null,false,true,123,"abc"]
}

// ExampleStringify_object 展示如何字符串化对象值
func ExampleStringify_object() {
	v := Value{
		Type: OBJECT,
		O: []Member{
			{K: "null", V: &Value{Type: NULL}},
			{K: "false", V: &Value{Type: FALSE}},
			{K: "true", V: &Value{Type: TRUE}},
			{K: "number", V: &Value{Type: NUMBER, N: 123}},
			{K: "string", V: &Value{Type: STRING, S: "abc"}},
		},
	}
	json, _ := Stringify(&v)
	fmt.Println(json)
	// Output: {"null":null,"false":false,"true":true,"number":123,"string":"abc"}
}

// ExampleStringify_complex 展示如何字符串化复杂值
func ExampleStringify_complex() {
	v := Value{
		Type: OBJECT,
		O: []Member{
			{K: "a", V: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 2},
				},
			}},
			{K: "b", V: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "c", V: &Value{Type: NUMBER, N: 0}},
					{K: "d", V: &Value{
						Type: ARRAY,
						A: []*Value{
							{Type: FALSE},
							{Type: TRUE},
							{Type: NULL},
							{Type: STRING, S: "Hello"},
						},
					}},
				},
			}},
		},
	}
	json, _ := Stringify(&v)
	fmt.Println(json)
	// Output: {"a":[1,2],"b":{"c":0,"d":[false,true,null,"Hello"]}}
}

// ExampleParse_stringify 展示解析和字符串化的结合使用
func ExampleParse_stringify() {
	jsonText := `{"a":[1,2],"b":{"c":0,"d":[false,true,null,"Hello"]}}`
	var v Value

	if err := Parse(&v, jsonText); err == PARSE_OK {
		jsonOut, _ := Stringify(&v)
		fmt.Println(jsonOut)
	}
	// Output: {"a":[1,2],"b":{"c":0,"d":[false,true,null,"Hello"]}}
}

// 测试Equal函数
func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		lhs      *Value
		rhs      *Value
		expected bool
	}{
		{
			name:     "NullValues",
			lhs:      &Value{Type: NULL},
			rhs:      &Value{Type: NULL},
			expected: true,
		},
		{
			name:     "TrueValues",
			lhs:      &Value{Type: TRUE},
			rhs:      &Value{Type: TRUE},
			expected: true,
		},
		{
			name:     "FalseValues",
			lhs:      &Value{Type: FALSE},
			rhs:      &Value{Type: FALSE},
			expected: true,
		},
		{
			name:     "DifferentTypes",
			lhs:      &Value{Type: TRUE},
			rhs:      &Value{Type: FALSE},
			expected: false,
		},
		{
			name:     "EqualNumbers",
			lhs:      &Value{Type: NUMBER, N: 123.45},
			rhs:      &Value{Type: NUMBER, N: 123.45},
			expected: true,
		},
		{
			name:     "DifferentNumbers",
			lhs:      &Value{Type: NUMBER, N: 123.45},
			rhs:      &Value{Type: NUMBER, N: 123.46},
			expected: false,
		},
		{
			name:     "EqualStrings",
			lhs:      &Value{Type: STRING, S: "hello"},
			rhs:      &Value{Type: STRING, S: "hello"},
			expected: true,
		},
		{
			name:     "DifferentStrings",
			lhs:      &Value{Type: STRING, S: "hello"},
			rhs:      &Value{Type: STRING, S: "world"},
			expected: false,
		},
		{
			name: "EqualArrays",
			lhs: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 2},
				},
			},
			rhs: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 2},
				},
			},
			expected: true,
		},
		{
			name: "DifferentArrayLength",
			lhs: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 2},
				},
			},
			rhs: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
				},
			},
			expected: false,
		},
		{
			name: "DifferentArrayElements",
			lhs: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 2},
				},
			},
			rhs: &Value{
				Type: ARRAY,
				A: []*Value{
					{Type: NUMBER, N: 1},
					{Type: NUMBER, N: 3},
				},
			},
			expected: false,
		},
		{
			name: "EqualObjects",
			lhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			rhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			expected: true,
		},
		{
			name: "EqualObjectsDifferentOrder",
			lhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			rhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
				},
			},
			expected: true,
		},
		{
			name: "DifferentObjectLength",
			lhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			rhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
				},
			},
			expected: false,
		},
		{
			name: "DifferentObjectValues",
			lhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			rhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 3}},
				},
			},
			expected: false,
		},
		{
			name: "DifferentObjectKeys",
			lhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "b", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			rhs: &Value{
				Type: OBJECT,
				O: []Member{
					{K: "a", V: &Value{Type: NUMBER, N: 1}},
					{K: "c", V: &Value{Type: NUMBER, N: 2}},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Equal(test.lhs, test.rhs)
			if actual != test.expected {
				t.Errorf("Equal(%v, %v) = %v, expected %v", test.lhs, test.rhs, actual, test.expected)
			}
		})
	}
}

// 测试Copy函数
func TestCopy(t *testing.T) {
	t.Run("CopyNull", func(t *testing.T) {
		src := &Value{Type: NULL}
		dst := &Value{Type: TRUE}
		Copy(dst, src)
		if dst.Type != NULL {
			t.Errorf("Copy null值失败，目标类型为 %v，期望 NULL", dst.Type)
		}
	})

	t.Run("CopyBoolean", func(t *testing.T) {
		src := &Value{Type: TRUE}
		dst := &Value{Type: NULL}
		Copy(dst, src)
		if dst.Type != TRUE {
			t.Errorf("Copy true值失败，目标类型为 %v，期望 TRUE", dst.Type)
		}
	})

	t.Run("CopyNumber", func(t *testing.T) {
		src := &Value{Type: NUMBER, N: 123.45}
		dst := &Value{Type: NULL}
		Copy(dst, src)
		if dst.Type != NUMBER || dst.N != 123.45 {
			t.Errorf("Copy number值失败，目标为 {%v, %v}，期望 {NUMBER, 123.45}", dst.Type, dst.N)
		}
	})

	t.Run("CopyString", func(t *testing.T) {
		src := &Value{Type: STRING, S: "hello"}
		dst := &Value{Type: NULL}
		Copy(dst, src)
		if dst.Type != STRING || dst.S != "hello" {
			t.Errorf("Copy string值失败，目标为 {%v, %v}，期望 {STRING, hello}", dst.Type, dst.S)
		}
	})

	t.Run("CopyArray", func(t *testing.T) {
		src := &Value{
			Type: ARRAY,
			A: []*Value{
				{Type: NUMBER, N: 1},
				{Type: NUMBER, N: 2},
			},
		}
		dst := &Value{Type: NULL}
		Copy(dst, src)
		if dst.Type != ARRAY || len(dst.A) != 2 {
			t.Errorf("Copy array值失败，目标类型为 %v，长度为 %v", dst.Type, len(dst.A))
			return
		}
		if dst.A[0].Type != NUMBER || dst.A[0].N != 1 {
			t.Errorf("Copy array第一个元素失败，为 {%v, %v}，期望 {NUMBER, 1}", dst.A[0].Type, dst.A[0].N)
		}
		if dst.A[1].Type != NUMBER || dst.A[1].N != 2 {
			t.Errorf("Copy array第二个元素失败，为 {%v, %v}，期望 {NUMBER, 2}", dst.A[1].Type, dst.A[1].N)
		}
	})

	t.Run("CopyObject", func(t *testing.T) {
		src := &Value{
			Type: OBJECT,
			O: []Member{
				{K: "a", V: &Value{Type: NUMBER, N: 1}},
				{K: "b", V: &Value{Type: NUMBER, N: 2}},
			},
		}
		dst := &Value{Type: NULL}
		Copy(dst, src)
		if dst.Type != OBJECT || len(dst.O) != 2 {
			t.Errorf("Copy object值失败，目标类型为 %v，成员数为 %v", dst.Type, len(dst.O))
			return
		}
		if dst.O[0].K != "a" || dst.O[0].V.Type != NUMBER || dst.O[0].V.N != 1 {
			t.Errorf("Copy object第一个成员失败，为 {%v, {%v, %v}}，期望 {a, {NUMBER, 1}}",
				dst.O[0].K, dst.O[0].V.Type, dst.O[0].V.N)
		}
		if dst.O[1].K != "b" || dst.O[1].V.Type != NUMBER || dst.O[1].V.N != 2 {
			t.Errorf("Copy object第二个成员失败，为 {%v, {%v, %v}}，期望 {b, {NUMBER, 2}}",
				dst.O[1].K, dst.O[1].V.Type, dst.O[1].V.N)
		}
	})
}

// 测试Move函数
func TestMove(t *testing.T) {
	t.Run("MoveNumber", func(t *testing.T) {
		src := &Value{Type: NUMBER, N: 123.45}
		dst := &Value{Type: NULL}
		Move(dst, src)
		if dst.Type != NUMBER || dst.N != 123.45 {
			t.Errorf("Move number值失败，目标为 {%v, %v}，期望 {NUMBER, 123.45}", dst.Type, dst.N)
		}
		if src.Type != NULL {
			t.Errorf("Move后源值类型应为NULL，但为 %v", src.Type)
		}
	})

	t.Run("MoveArray", func(t *testing.T) {
		src := &Value{
			Type: ARRAY,
			A: []*Value{
				{Type: NUMBER, N: 1},
				{Type: NUMBER, N: 2},
			},
		}
		dst := &Value{Type: NULL}
		Move(dst, src)
		if dst.Type != ARRAY || len(dst.A) != 2 {
			t.Errorf("Move array值失败，目标类型为 %v，长度为 %v", dst.Type, len(dst.A))
			return
		}
		if src.Type != NULL || src.A != nil {
			t.Errorf("Move后源值应为NULL，数组应为nil，但类型为 %v，数组为 %v", src.Type, src.A)
		}
	})
}

// 测试Swap函数
func TestSwap(t *testing.T) {
	t.Run("SwapDifferentTypes", func(t *testing.T) {
		a := &Value{Type: NUMBER, N: 123}
		b := &Value{Type: STRING, S: "hello"}
		Swap(a, b)
		if a.Type != STRING || a.S != "hello" {
			t.Errorf("Swap后a应为STRING类型，值为hello，但为 {%v, %v}", a.Type, a.S)
		}
		if b.Type != NUMBER || b.N != 123 {
			t.Errorf("Swap后b应为NUMBER类型，值为123，但为 {%v, %v}", b.Type, b.N)
		}
	})
}

// 测试动态数组
func TestDynamicArray(t *testing.T) {
	t.Run("SetArray", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 10)
		if v.Type != ARRAY {
			t.Errorf("SetArray后类型应为ARRAY，但为 %v", v.Type)
		}
		if cap(v.A) != 10 {
			t.Errorf("SetArray后容量应为10，但为 %v", cap(v.A))
		}
		if len(v.A) != 0 {
			t.Errorf("SetArray后长度应为0，但为 %v", len(v.A))
		}
	})

	t.Run("PushBack", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 2)
		e1 := PushBackArrayElement(v)
		SetNumber(e1, 1)
		e2 := PushBackArrayElement(v)
		SetNumber(e2, 2)
		e3 := PushBackArrayElement(v) // 触发扩容
		SetNumber(e3, 3)

		if len(v.A) != 3 {
			t.Errorf("PushBack后长度应为3，但为 %v", len(v.A))
		}
		if cap(v.A) <= 2 {
			t.Errorf("PushBack应触发扩容，但容量为 %v", cap(v.A))
		}
		if v.A[0].N != 1 || v.A[1].N != 2 || v.A[2].N != 3 {
			t.Errorf("PushBack的元素值错误，为 [%v, %v, %v]，期望 [1, 2, 3]",
				v.A[0].N, v.A[1].N, v.A[2].N)
		}
	})

	t.Run("PopBack", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 0)
		e1 := PushBackArrayElement(v)
		SetNumber(e1, 1)
		e2 := PushBackArrayElement(v)
		SetNumber(e2, 2)

		PopBackArrayElement(v)
		if len(v.A) != 1 {
			t.Errorf("PopBack后长度应为1，但为 %v", len(v.A))
		}
		if v.A[0].N != 1 {
			t.Errorf("PopBack后数组元素错误，为 %v，期望 1", v.A[0].N)
		}
	})

	t.Run("Insert", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 0)
		e1 := PushBackArrayElement(v)
		SetNumber(e1, 1)
		e2 := PushBackArrayElement(v)
		SetNumber(e2, 3)

		// 在索引1插入元素
		e3 := InsertArrayElement(v, 1)
		SetNumber(e3, 2)

		if len(v.A) != 3 {
			t.Errorf("Insert后长度应为3，但为 %v", len(v.A))
		}
		if v.A[0].N != 1 || v.A[1].N != 2 || v.A[2].N != 3 {
			t.Errorf("Insert的元素值错误，为 [%v, %v, %v]，期望 [1, 2, 3]",
				v.A[0].N, v.A[1].N, v.A[2].N)
		}
	})

	t.Run("Erase", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 0)
		for i := 1; i <= 5; i++ {
			e := PushBackArrayElement(v)
			SetNumber(e, float64(i))
		}

		// 删除索引1开始的2个元素
		EraseArrayElement(v, 1, 2)

		if len(v.A) != 3 {
			t.Errorf("Erase后长度应为3，但为 %v", len(v.A))
		}
		if v.A[0].N != 1 || v.A[1].N != 4 || v.A[2].N != 5 {
			t.Errorf("Erase后元素值错误，为 [%v, %v, %v]，期望 [1, 4, 5]",
				v.A[0].N, v.A[1].N, v.A[2].N)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 0)
		for i := 1; i <= 3; i++ {
			e := PushBackArrayElement(v)
			SetNumber(e, float64(i))
		}

		originalCap := cap(v.A)
		ClearArray(v)

		if len(v.A) != 0 {
			t.Errorf("Clear后长度应为0，但为 %v", len(v.A))
		}
		if cap(v.A) != originalCap {
			t.Errorf("Clear后容量应保持不变，期望 %v，但为 %v", originalCap, cap(v.A))
		}
	})

	t.Run("Shrink", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetArray(v, 10)
		for i := 1; i <= 3; i++ {
			e := PushBackArrayElement(v)
			SetNumber(e, float64(i))
		}

		if cap(v.A) != 10 {
			t.Errorf("初始容量应为10，但为 %v", cap(v.A))
		}

		ShrinkArray(v)

		if cap(v.A) != 3 {
			t.Errorf("Shrink后容量应为3，但为 %v", cap(v.A))
		}
		if len(v.A) != 3 {
			t.Errorf("Shrink后长度应为3，但为 %v", len(v.A))
		}
	})
}

// 测试动态对象
func TestDynamicObject(t *testing.T) {
	t.Run("SetObject", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetObject(v)
		if v.Type != OBJECT {
			t.Errorf("SetObject后类型应为OBJECT，但为 %v", v.Type)
		}
		if len(v.O) != 0 {
			t.Errorf("SetObject后成员数应为0，但为 %v", len(v.O))
		}
	})

	t.Run("SetObjectValue", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetObject(v)

		// 添加键值对
		e1 := SetObjectValue(v, "a")
		SetNumber(e1, 1)
		e2 := SetObjectValue(v, "b")
		SetNumber(e2, 2)

		if len(v.O) != 2 {
			t.Errorf("添加键值对后成员数应为2，但为 %v", len(v.O))
		}

		// 修改已存在的键
		e3 := SetObjectValue(v, "a")
		SetNumber(e3, 3)

		if len(v.O) != 2 {
			t.Errorf("修改已存在键后成员数应仍为2，但为 %v", len(v.O))
		}

		// 检查值是否正确
		if v.O[0].K != "a" || v.O[0].V.N != 3 {
			t.Errorf("第一个成员应为 {a, 3}，但为 {%v, %v}", v.O[0].K, v.O[0].V.N)
		}
		if v.O[1].K != "b" || v.O[1].V.N != 2 {
			t.Errorf("第二个成员应为 {b, 2}，但为 {%v, %v}", v.O[1].K, v.O[1].V.N)
		}
	})

	t.Run("RemoveObjectValue", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetObject(v)

		// 添加键值对
		e1 := SetObjectValue(v, "a")
		SetNumber(e1, 1)
		e2 := SetObjectValue(v, "b")
		SetNumber(e2, 2)
		e3 := SetObjectValue(v, "c")
		SetNumber(e3, 3)

		// 删除中间的键值对
		RemoveObjectValue(v, 1)

		if len(v.O) != 2 {
			t.Errorf("删除后成员数应为2，但为 %v", len(v.O))
		}

		// 检查值是否正确
		if v.O[0].K != "a" || v.O[0].V.N != 1 {
			t.Errorf("第一个成员应为 {a, 1}，但为 {%v, %v}", v.O[0].K, v.O[0].V.N)
		}
		if v.O[1].K != "c" || v.O[1].V.N != 3 {
			t.Errorf("第二个成员应为 {c, 3}，但为 {%v, %v}", v.O[1].K, v.O[1].V.N)
		}
	})

	t.Run("ClearObject", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetObject(v)

		// 添加键值对
		for i := 0; i < 3; i++ {
			key := string([]rune{'a' + rune(i)})
			e := SetObjectValue(v, key)
			SetNumber(e, float64(i+1))
		}

		originalCap := cap(v.O)
		ClearObject(v)

		if len(v.O) != 0 {
			t.Errorf("Clear后成员数应为0，但为 %v", len(v.O))
		}
		if cap(v.O) != originalCap {
			t.Errorf("Clear后容量应保持不变，期望 %v，但为 %v", originalCap, cap(v.O))
		}
	})

	t.Run("ObjectCapacity", func(t *testing.T) {
		v := &Value{Type: NULL}
		SetObject(v)

		// 手动设置容量
		v.O = make([]Member, 0, 10)

		if GetObjectCapacity(v) != 10 {
			t.Errorf("对象容量应为10，但为 %v", GetObjectCapacity(v))
		}

		// 添加5个成员
		for i := 0; i < 5; i++ {
			key := string([]rune{'a' + rune(i)})
			e := SetObjectValue(v, key)
			SetNumber(e, float64(i+1))
		}

		// 收缩容量
		ShrinkObject(v)

		if GetObjectCapacity(v) != 5 {
			t.Errorf("收缩后容量应为5，但为 %v", GetObjectCapacity(v))
		}

		// 扩充容量
		ReserveObject(v, 20)

		if GetObjectCapacity(v) != 20 {
			t.Errorf("扩充后容量应为20，但为 %v", GetObjectCapacity(v))
		}
	})
}

// 测试FindObjectKey函数
func TestFindObjectKey(t *testing.T) {
	t.Run("FindExistingKey", func(t *testing.T) {
		v := &Value{
			Type: OBJECT,
			O: []Member{
				{K: "a", V: &Value{Type: NUMBER, N: 1}},
				{K: "b", V: &Value{Type: NUMBER, N: 2}},
				{K: "c", V: &Value{Type: NUMBER, N: 3}},
			},
		}

		// 查找存在的键
		value, found := FindObjectKey(v, "b")
		if !found {
			t.Errorf("未找到应存在的键 'b'")
		}
		if value.Type != NUMBER || value.N != 2 {
			t.Errorf("找到的值不匹配，期望 {NUMBER, 2}，实际为 {%v, %v}", value.Type, value.N)
		}
	})

	t.Run("FindNonExistingKey", func(t *testing.T) {
		v := &Value{
			Type: OBJECT,
			O: []Member{
				{K: "a", V: &Value{Type: NUMBER, N: 1}},
				{K: "b", V: &Value{Type: NUMBER, N: 2}},
			},
		}

		// 查找不存在的键
		_, found := FindObjectKey(v, "x")
		if found {
			t.Errorf("错误地找到了不存在的键 'x'")
		}
	})

	t.Run("FindInNullValue", func(t *testing.T) {
		v := &Value{Type: NULL}

		// 在NULL值中查找键
		_, found := FindObjectKey(v, "a")
		if found {
			t.Errorf("在NULL值中错误地找到了键")
		}
	})

	t.Run("FindInNilValue", func(t *testing.T) {
		// 在nil值中查找键
		_, found := FindObjectKey(nil, "a")
		if found {
			t.Errorf("在nil值中错误地找到了键")
		}
	})
}
