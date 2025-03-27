// leptjson_test.go - Go语言版JSON库测试
package leptjson

import (
	"fmt"
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
