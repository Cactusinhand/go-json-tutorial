// leptjson_test.go - Go语言版JSON库测试
package leptjson

import (
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
	tests := []struct {
		input    string
		expected string
	}{
		{`""`, ""},
		{`"Hello"`, "Hello"},
		{`"Hello\nWorld"`, "Hello\nWorld"},
		{`"\" \\ \/ \b \f \n \r \t"`, "\" \\ / \b \f \n \r \t"},
		{`"\u0024"`, "$"},       // Basic ASCII
		{`"\u00A2"`, "¢"},       // Cents sign
		{`"\u20AC"`, "€"},       // Euro sign
		{`"\uD834\uDD1E"`, "𝄞"}, // G clef (surrogate pair)
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, test.input); err != PARSE_OK {
				t.Errorf("Parse failed: %v", err)
			}
			if GetType(&v) != STRING {
				t.Errorf("Expected STRING type, got %v", GetType(&v))
			}
			if got := GetString(&v); got != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, got)
			}
		})
	}
}

// 测试字符串API
func TestStringAPI(t *testing.T) {
	v := Value{}
	SetString(&v, "Hello")
	if GetType(&v) != STRING {
		t.Errorf("期望类型为STRING，但得到: %v", GetType(&v))
	}
	if GetString(&v) != "Hello" {
		t.Errorf("期望值为Hello，但得到: %s", GetString(&v))
	}
	if GetStringLength(&v) != 5 {
		t.Errorf("期望长度为5，但得到: %d", GetStringLength(&v))
	}

	// 测试空字符串
	SetString(&v, "")
	if GetType(&v) != STRING {
		t.Errorf("期望类型为STRING，但得到: %v", GetType(&v))
	}
	if GetString(&v) != "" {
		t.Errorf("期望值为空字符串，但得到: %s", GetString(&v))
	}
	if GetStringLength(&v) != 0 {
		t.Errorf("期望长度为0，但得到: %d", GetStringLength(&v))
	}

	// 测试包含特殊字符的字符串
	SetString(&v, "Hello\nWorld")
	if GetType(&v) != STRING {
		t.Errorf("期望类型为STRING，但得到: %v", GetType(&v))
	}
	if GetString(&v) != "Hello\nWorld" {
		t.Errorf("期望值为Hello\nWorld，但得到: %s", GetString(&v))
	}
	if GetStringLength(&v) != 11 {
		t.Errorf("期望长度为11，但得到: %d", GetStringLength(&v))
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

// 测试字符串解析错误
func TestParseInvalidString(t *testing.T) {
	tests := []string{
		`"`,                // Missing quotation mark
		`"abc`,             // Missing quotation mark
		`"\v"`,             // Invalid escape character
		`"\0"`,             // Invalid escape character
		`"\x12"`,           // Invalid escape character
		`"abc\tabc"def"`,   // Additional content
		`"\uD800"`,         // Invalid surrogate pair (missing low surrogate)
		`"\uDBFF\uDBFF"`,   // Invalid surrogate pair
		`"abc\n\tabc"def"`, // Additional content after valid string
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, test); err == PARSE_OK {
				t.Errorf("Expected parsing to fail for %q", test)
			}
		})
	}
}
