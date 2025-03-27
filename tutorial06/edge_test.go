package leptjson

import (
	"testing"
)

// TestEdgeCases 测试各种边缘情况
func TestEdgeCases(t *testing.T) {
	// 测试各种空白字符的处理
	t.Run("WhitespaceHandling", func(t *testing.T) {
		v := Value{}
		if err := Parse(&v, "   \t\r\n  null  \t\r\n  "); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != NULL {
			t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
		}
	})

	// 测试非常长的数字
	t.Run("VeryLongNumber", func(t *testing.T) {
		v := Value{}
		longNumber := "1234567890123456789012345678901234567890"
		if err := Parse(&v, longNumber); err != PARSE_OK {
			t.Errorf("期望解析成功，但返回错误: %v", err)
		}
		if GetType(&v) != NUMBER {
			t.Errorf("期望类型为NUMBER，但得到: %v", GetType(&v))
		}
	})

	// 测试极端情况的数字
	t.Run("ExtremeNumbers", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected float64
		}{
			{"1e-10000", 0.0}, // 非常小的数，接近0
			{"1e-323", 0.0},   // 接近Go语言double精度下限
			{"1.7976931348623157e+308", 1.7976931348623157e+308}, // 接近Go语言double精度上限
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
			})
		}
	})

	// 测试各种Unicode字符和转义序列
	t.Run("UnicodeEdgeCases", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{`"\u0000"`, "\u0000"},           // NULL字符
			{`"\u001F"`, "\u001F"},           // 单位分隔符
			{`"\u0080"`, "\u0080"},           // 第一个扩展ASCII字符
			{`"\u00FF"`, "\u00FF"},           // 最后一个扩展ASCII字符
			{`"\u2000"`, "\u2000"},           // En Quad
			{`"\uFFFF"`, "\uFFFF"},           // 最大的BMP字符
			{`"\uD834\uDD1E"`, "\U0001D11E"}, // 𝄞 - 高音谱号，超出BMP
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
	})

	// 测试深度嵌套的数组和对象
	t.Run("DeepNesting", func(t *testing.T) {
		// 深度嵌套的数组
		t.Run("DeepNestedArray", func(t *testing.T) {
			v := Value{}
			deepArray := "[[[[[[[[[[1]]]]]]]]]]" // 10层嵌套
			if err := Parse(&v, deepArray); err != PARSE_OK {
				t.Errorf("期望解析成功，但返回错误: %v", err)
			}
			if GetType(&v) != ARRAY {
				t.Errorf("期望类型为ARRAY，但得到: %v", GetType(&v))
			}

			// 验证嵌套结构
			current := &v
			for i := 0; i < 10; i++ {
				if GetArraySize(current) != 1 {
					t.Errorf("期望数组大小为1，但得到: %d", GetArraySize(current))
					break
				}
				current = GetArrayElement(current, 0)
				if i < 9 && GetType(current) != ARRAY {
					t.Errorf("期望第%d层为ARRAY，但得到: %v", i+1, GetType(current))
					break
				}
			}
			if GetType(current) != NUMBER || GetNumber(current) != 1.0 {
				t.Errorf("期望最内层元素为数字1，但得到: %v", current)
			}
		})

		// 深度嵌套的对象
		t.Run("DeepNestedObject", func(t *testing.T) {
			v := Value{}
			deepObject := `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":1}}}}}}}}}}` // 10层嵌套
			if err := Parse(&v, deepObject); err != PARSE_OK {
				t.Errorf("期望解析成功，但返回错误: %v", err)
			}
			if GetType(&v) != OBJECT {
				t.Errorf("期望类型为OBJECT，但得到: %v", GetType(&v))
			}

			// 验证嵌套结构
			keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
			current := &v
			for i, key := range keys {
				if GetObjectSize(current) != 1 {
					t.Errorf("期望对象大小为1，但得到: %d", GetObjectSize(current))
					break
				}
				if GetObjectKey(current, 0) != key {
					t.Errorf("期望键为%s，但得到: %s", key, GetObjectKey(current, 0))
					break
				}
				current = GetObjectValue(current, 0)
				if i < 9 && GetType(current) != OBJECT {
					t.Errorf("期望第%d层为OBJECT，但得到: %v", i+1, GetType(current))
					break
				}
			}
			if GetType(current) != NUMBER || GetNumber(current) != 1.0 {
				t.Errorf("期望最内层值为数字1，但得到: %v", current)
			}
		})
	})

	// 测试空对象和空数组的各种形式
	t.Run("EmptyContainers", func(t *testing.T) {
		testCases := []struct {
			name  string
			input string
			type_ ValueType
		}{
			{"EmptyArray", "[]", ARRAY},
			{"EmptyArrayWithSpace", "[ ]", ARRAY},
			{"EmptyArrayWithNewLines", "[\n]", ARRAY},
			{"EmptyObject", "{}", OBJECT},
			{"EmptyObjectWithSpace", "{ }", OBJECT},
			{"EmptyObjectWithNewLines", "{\n}", OBJECT},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				v := Value{}
				if err := Parse(&v, tc.input); err != PARSE_OK {
					t.Errorf("期望解析成功，但返回错误: %v", err)
				}
				if GetType(&v) != tc.type_ {
					t.Errorf("期望类型为%v，但得到: %v", tc.type_, GetType(&v))
				}
				if tc.type_ == ARRAY && GetArraySize(&v) != 0 {
					t.Errorf("期望数组大小为0，但得到: %d", GetArraySize(&v))
				}
				if tc.type_ == OBJECT && GetObjectSize(&v) != 0 {
					t.Errorf("期望对象大小为0，但得到: %d", GetObjectSize(&v))
				}
			})
		}
	})

	// 测试错误的JSON格式（应该被拒绝）
	t.Run("InvalidJSON", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected ParseError
		}{
			{"InvalidNumber_LeadingPlus", "+1", PARSE_INVALID_VALUE},
			{"InvalidNumber_LeadingDot", ".123", PARSE_INVALID_VALUE},
			{"InvalidNumber_TrailingDot", "1.", PARSE_INVALID_VALUE},
			{"InvalidNumber_OnlyDot", ".", PARSE_INVALID_VALUE},
			{"InvalidNumber_PlusDot", "+.", PARSE_INVALID_VALUE},
			{"InvalidNumber_ExponentOnly", "e1", PARSE_INVALID_VALUE},
			{"InvalidNumber_ExponentDot", "1e.", PARSE_INVALID_VALUE},
			{"InvalidNumber_NaN", "NaN", PARSE_INVALID_VALUE},
			{"InvalidNumber_Infinity", "Infinity", PARSE_INVALID_VALUE},
			{"InvalidNumber_HexFormat", "0xF", PARSE_INVALID_VALUE},

			{"IncompleteTrue", "tru", PARSE_INVALID_VALUE},
			{"IncompleteFalse", "fals", PARSE_INVALID_VALUE},
			{"IncompleteNull", "nul", PARSE_INVALID_VALUE},

			{"MissingQuote", "\"abc", PARSE_MISS_QUOTATION_MARK},
			{"InvalidEscape", "\"\\x\"", PARSE_INVALID_STRING_ESCAPE},
			{"InvalidUnicode", "\"\\u123\"", PARSE_INVALID_UNICODE_HEX},
			{"InvalidUnicodeHigh", "\"\\uD800\"", PARSE_INVALID_UNICODE_SURROGATE},
			{"InvalidUnicodeLow", "\"\\uDC00\"", PARSE_INVALID_UNICODE_SURROGATE},
			{"InvalidSurrogate", "\"\\uD800\\u0000\"", PARSE_INVALID_UNICODE_SURROGATE},

			{"MissingArrayBracket", "[1,2", PARSE_MISS_COMMA_OR_SQUARE_BRACKET},
			{"TrailingCommaArray", "[1,2,]", PARSE_INVALID_VALUE},
			{"MissingCommaArray", "[1 2]", PARSE_MISS_COMMA_OR_SQUARE_BRACKET},

			{"MissingObjectBrace", "{\"a\":1", PARSE_MISS_COMMA_OR_CURLY_BRACKET},
			{"InvalidObjectKey", "{1:1}", PARSE_MISS_KEY},
			{"MissingObjectKey", "{:1}", PARSE_MISS_KEY},
			{"MissingObjectColon", "{\"a\" 1}", PARSE_MISS_COLON},
			{"TrailingCommaObject", "{\"a\":1,}", PARSE_MISS_KEY},
			{"MissingCommaObject", "{\"a\":1 \"b\":2}", PARSE_MISS_COMMA_OR_CURLY_BRACKET},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				v := Value{}
				if err := Parse(&v, tc.input); err != tc.expected {
					t.Errorf("期望错误%v，但得到: %v", tc.expected, err)
				}
			})
		}
	})
}
