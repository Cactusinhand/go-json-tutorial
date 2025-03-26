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
				t.Errorf("期望值为%g，但得到: %g", tc.expected, GetNumber(&v))
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
		{`"\u0024"`, "$"},       // 基本多语言平面内的字符U+0024
		{`"\u00A2"`, "¢"},       // 基本多语言平面内的字符U+00A2
		{`"\u20AC"`, "€"},       // 基本多语言平面内的字符U+20AC
		{`"\uD834\uDD1E"`, "𝄞"}, // 辅助平面字符U+1D11E (𝄞)
		{`"\ud834\udd1e"`, "𝄞"}, // 辅助平面字符U+1D11E (𝄞)，小写表示
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

	// 测试无效数字
	invalidNumbers := []string{
		"+0",    // 不允许正号
		"+1",    // 不允许正号
		".123",  // 小数点前必须有数字
		"1.",    // 小数点后必须有数字
		"INF",   // 不是合法的JSON数字
		"inf",   // 不是合法的JSON数字
		"NAN",   // 不是合法的JSON数字
		"nan",   // 不是合法的JSON数字
		"0123",  // 前导零后不能有数字
		"0x0",   // 不支持十六进制
		"0x123", // 不支持十六进制
		"0123",  // 不允许前导零
		"1e",    // 指数部分不完整
		"1e+",   // 指数部分不完整
		"1e-",   // 指数部分不完整
	}

	for _, invalidNum := range invalidNumbers {
		t.Run(invalidNum, func(t *testing.T) {
			v := Value{Type: TRUE}
			if err := Parse(&v, invalidNum); err != PARSE_INVALID_VALUE {
				t.Errorf("期望错误PARSE_INVALID_VALUE，但得到: %v，输入: %s", err, invalidNum)
			}
			if GetType(&v) != NULL {
				t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
			}
		})
	}
}

// 测试解析字符串错误
func TestParseInvalidString(t *testing.T) {
	// 测试缺少引号
	v := Value{Type: TRUE}
	if err := Parse(&v, "\""); err != PARSE_MISS_QUOTATION_MARK {
		t.Errorf("期望错误PARSE_MISS_QUOTATION_MARK，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}

	// 测试无效的转义字符
	invalidEscapes := []string{
		"\"\\v\"",   // \v不是有效的转义字符
		"\"\\0\"",   // \0不是有效的转义字符
		"\"\\x12\"", // \x不是有效的转义字符
	}

	for _, invalidEsc := range invalidEscapes {
		t.Run(invalidEsc, func(t *testing.T) {
			v := Value{Type: TRUE}
			if err := Parse(&v, invalidEsc); err != PARSE_INVALID_STRING_ESCAPE {
				t.Errorf("期望错误PARSE_INVALID_STRING_ESCAPE，但得到: %v，输入: %s", err, invalidEsc)
			}
			if GetType(&v) != NULL {
				t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
			}
		})
	}

	// 测试无效的字符
	invalidChars := []string{
		"\"\x01\"", // 控制字符U+0001
		"\"\x1F\"", // 控制字符U+001F
	}

	for _, invalidChar := range invalidChars {
		t.Run(fmt.Sprintf("InvalidChar-%X", invalidChar[1]), func(t *testing.T) {
			v := Value{Type: TRUE}
			if err := Parse(&v, invalidChar); err != PARSE_INVALID_STRING_CHAR {
				t.Errorf("期望错误PARSE_INVALID_STRING_CHAR，但得到: %v", err)
			}
			if GetType(&v) != NULL {
				t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
			}
		})
	}

	// 测试无效的Unicode十六进制
	invalidUnicodeHex := []string{
		"\"\\u\"",     // 缺少4位十六进制数字
		"\"\\u0\"",    // 不足4位十六进制数字
		"\"\\u01\"",   // 不足4位十六进制数字
		"\"\\u012\"",  // 不足4位十六进制数字
		"\"\\u012Z\"", // 包含非十六进制字符
		"\"\\u000G\"", // 包含非十六进制字符
	}

	for _, invalidHex := range invalidUnicodeHex {
		t.Run(invalidHex, func(t *testing.T) {
			v := Value{Type: TRUE}
			if err := Parse(&v, invalidHex); err != PARSE_INVALID_UNICODE_HEX {
				t.Errorf("期望错误PARSE_INVALID_UNICODE_HEX，但得到: %v，输入: %s", err, invalidHex)
			}
			if GetType(&v) != NULL {
				t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
			}
		})
	}

	// 测试无效的Unicode代理对
	invalidSurrogates := []string{
		"\"\\uD800\"",        // 只有高代理项，缺少低代理项
		"\"\\uDBFF\"",        // 只有高代理项，缺少低代理项
		"\"\\uD800\\\"",      // 高代理项后面不是\u
		"\"\\uD800\\uE000\"", // 高代理项后面不是低代理项
		"\"\\uD800\\uDBFF\"", // 高代理项后面是另一个高代理项
	}

	for _, invalidSurrogate := range invalidSurrogates {
		t.Run(invalidSurrogate, func(t *testing.T) {
			v := Value{Type: TRUE}
			if err := Parse(&v, invalidSurrogate); err != PARSE_INVALID_UNICODE_SURROGATE {
				t.Errorf("期望错误PARSE_INVALID_UNICODE_SURROGATE，但得到: %v，输入: %s", err, invalidSurrogate)
			}
			if GetType(&v) != NULL {
				t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
			}
		})
	}
}

// 测试数字太大
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
	// Output: Hello, World!
}

// 示例 - 解析Unicode字符串
func ExampleParse_unicodeString() {
	v := Value{}
	Parse(&v, "\"\\uD834\\uDD1E\"")
	fmt.Println(v.String())
	// Output: 𝄞
}

// 示例 - 错误处理
func ExampleParse_error() {
	v := Value{}
	err := Parse(&v, "invalid")
	fmt.Println(err)
	// Output: 无效的值
}
