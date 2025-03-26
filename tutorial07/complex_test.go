package leptjson

import (
	"strconv"
	"strings"
	"testing"
)

// TestComplexCombinations 测试复杂的JSON对象/数组组合
func TestComplexCombinations(t *testing.T) {
	// 测试具有各种类型的复杂JSON
	t.Run("MixedTypes", func(t *testing.T) {
		v := Value{}
		complexJSON := `{
			"null": null,
			"boolean": true,
			"number": 123.456,
			"string": "Hello, World!",
			"array": [1, false, "text", {"key": "value"}],
			"object": {"a": 1, "b": "text", "c": [1,2,3]}
		}`

		if err := Parse(&v, complexJSON); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT {
			t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
		}
		if GetObjectSize(&v) != 6 {
			t.Errorf("期望对象大小为6，但得到: %d", GetObjectSize(&v))
		}

		// 验证各个字段
		if nullValue := GetObjectValueByKey(&v, "null"); GetType(nullValue) != NULL {
			t.Errorf("期望\"null\"字段为NULL类型，但得到: %v", GetType(nullValue))
		}

		if boolValue := GetObjectValueByKey(&v, "boolean"); GetType(boolValue) != TRUE {
			t.Errorf("期望\"boolean\"字段为TRUE类型，但得到: %v", GetType(boolValue))
		}

		if numValue := GetObjectValueByKey(&v, "number"); GetType(numValue) != NUMBER || GetNumber(numValue) != 123.456 {
			t.Errorf("期望\"number\"字段为123.456，但得到: %v", numValue)
		}

		if strValue := GetObjectValueByKey(&v, "string"); GetType(strValue) != STRING || GetString(strValue) != "Hello, World!" {
			t.Errorf("期望\"string\"字段为\"Hello, World!\"，但得到: %v", strValue)
		}

		// 验证数组
		arrayValue := GetObjectValueByKey(&v, "array")
		if GetType(arrayValue) != ARRAY || GetArraySize(arrayValue) != 4 {
			t.Errorf("期望\"array\"字段为长度为4的数组，但得到: %v", arrayValue)
		} else {
			if GetType(GetArrayElement(arrayValue, 0)) != NUMBER || GetNumber(GetArrayElement(arrayValue, 0)) != 1.0 {
				t.Errorf("期望数组第一个元素为数字1，但得到: %v", GetArrayElement(arrayValue, 0))
			}
			if GetType(GetArrayElement(arrayValue, 1)) != FALSE {
				t.Errorf("期望数组第二个元素为FALSE，但得到: %v", GetArrayElement(arrayValue, 1))
			}
			if GetType(GetArrayElement(arrayValue, 2)) != STRING || GetString(GetArrayElement(arrayValue, 2)) != "text" {
				t.Errorf("期望数组第三个元素为字符串\"text\"，但得到: %v", GetArrayElement(arrayValue, 2))
			}
			if GetType(GetArrayElement(arrayValue, 3)) != OBJECT {
				t.Errorf("期望数组第四个元素为对象，但得到: %v", GetArrayElement(arrayValue, 3))
			}
		}

		// 验证嵌套对象
		objectValue := GetObjectValueByKey(&v, "object")
		if GetType(objectValue) != OBJECT || GetObjectSize(objectValue) != 3 {
			t.Errorf("期望\"object\"字段为大小为3的对象，但得到: %v", objectValue)
		} else {
			if numValue := GetObjectValueByKey(objectValue, "a"); GetType(numValue) != NUMBER || GetNumber(numValue) != 1.0 {
				t.Errorf("期望对象的\"a\"字段为数字1，但得到: %v", numValue)
			}
			if strValue := GetObjectValueByKey(objectValue, "b"); GetType(strValue) != STRING || GetString(strValue) != "text" {
				t.Errorf("期望对象的\"b\"字段为字符串\"text\"，但得到: %v", strValue)
			}
			if arrValue := GetObjectValueByKey(objectValue, "c"); GetType(arrValue) != ARRAY || GetArraySize(arrValue) != 3 {
				t.Errorf("期望对象的\"c\"字段为长度为3的数组，但得到: %v", arrValue)
			}
		}
	})
}

// TestPerformanceEdgeCases 测试性能边缘情况
func TestPerformanceEdgeCases(t *testing.T) {
	// 注意：这些测试可能会消耗较多资源，仅在需要时运行

	// 测试处理大型JSON数组
	t.Run("LargeArray", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过大数组测试")
		}

		// 创建一个包含1000个元素的数组
		var jsonBuilder strings.Builder
		jsonBuilder.WriteString("[")
		for i := 0; i < 1000; i++ {
			if i > 0 {
				jsonBuilder.WriteString(",")
			}
			jsonBuilder.WriteString(strconv.Itoa(i))
		}
		jsonBuilder.WriteString("]")

		v := Value{}
		if err := Parse(&v, jsonBuilder.String()); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != ARRAY || GetArraySize(&v) != 1000 {
			t.Errorf("期望数组大小为1000，但得到: %d", GetArraySize(&v))
		}
	})

	// 测试处理大型JSON对象
	t.Run("LargeObject", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过大对象测试")
		}

		// A 创建一个包含1000个键值对的对象
		var jsonBuilder strings.Builder
		jsonBuilder.WriteString("{")
		for i := 0; i < 1000; i++ {
			if i > 0 {
				jsonBuilder.WriteString(",")
			}
			key := "key" + strconv.Itoa(i)
			jsonBuilder.WriteString("\"" + key + "\":" + strconv.Itoa(i))
		}
		jsonBuilder.WriteString("}")

		v := Value{}
		if err := Parse(&v, jsonBuilder.String()); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != OBJECT || GetObjectSize(&v) != 1000 {
			t.Errorf("期望对象大小为1000，但得到: %d", GetObjectSize(&v))
		}
	})
}
