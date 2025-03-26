// leptjson_test.go - Go语言版JSON库测试
package leptjson

import (
	"fmt"
	"runtime"
	"sync"
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

// 测试数字边界条件
func TestParseNumberBoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ParseError
	}{
		{"MaxFloat64", "1.7976931348623157e+308", PARSE_OK},
		{"MinFloat64", "-1.7976931348623157e+308", PARSE_OK},
		{"MaxInt64", "9223372036854775807", PARSE_OK},
		{"MinInt64", "-9223372036854775808", PARSE_OK},
		{"TooLarge", "1.7976931348623157e+309", PARSE_NUMBER_TOO_BIG},
		{"TooSmall", "-1.7976931348623157e+309", PARSE_NUMBER_TOO_BIG},
		{"InvalidExponent", "1e", PARSE_INVALID_VALUE},
		{"InvalidExponentSign", "1e+", PARSE_INVALID_VALUE},
		{"InvalidDecimal", "1.", PARSE_INVALID_VALUE},
		{"InvalidNegative", "-", PARSE_INVALID_VALUE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, tt.input); err != tt.expected {
				t.Errorf("期望错误 %v，但得到: %v", tt.expected, err)
			}
		})
	}
}

// 测试数字精度
func TestNumberPrecision(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Zero", "0", 0},
		{"NegativeZero", "-0", 0},
		{"One", "1", 1},
		{"NegativeOne", "-1", -1},
		{"Decimal", "1.5", 1.5},
		{"Scientific", "1.5e2", 150},
		{"SmallNumber", "1e-10", 1e-10},
		{"LargeNumber", "1e10", 1e10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Value{}
			if err := Parse(&v, tt.input); err != PARSE_OK {
				t.Errorf("解析失败: %v", err)
			}
			if GetNumber(&v) != tt.expected {
				t.Errorf("期望值 %g，但得到: %g", tt.expected, GetNumber(&v))
			}
		})
	}
}

// 测试大数字性能
func BenchmarkParseLargeNumbers(b *testing.B) {
	tests := []string{
		"1.7976931348623157e+308",
		"-1.7976931348623157e+308",
		"9223372036854775807",
		"-9223372036854775808",
	}

	for _, test := range tests {
		b.Run(test, func(b *testing.B) {
			v := Value{}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Parse(&v, test)
			}
		})
	}
}

// 测试并发数字解析
func TestConcurrentNumberParse(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			v := Value{}
			for j := 0; j < numIterations; j++ {
				Parse(&v, "123.456")
				if GetType(&v) != NUMBER {
					t.Errorf("并发解析失败: 期望类型为NUMBER，但得到: %v", GetType(&v))
				}
				if GetNumber(&v) != 123.456 {
					t.Errorf("并发解析失败: 期望值为123.456，但得到: %g", GetNumber(&v))
				}
			}
		}()
	}

	wg.Wait()
}

// 测试内存使用
func TestNumberMemoryUsage(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beforeAlloc := m.Alloc

	// 创建大量包含数字的Value对象
	values := make([]Value, 1000)
	for i := range values {
		Parse(&values[i], "123.456")
	}

	runtime.ReadMemStats(&m)
	afterAlloc := m.Alloc

	// 检查内存增长是否合理
	if afterAlloc-beforeAlloc > 100*1024 { // 100KB
		t.Errorf("内存使用过高: %d bytes", afterAlloc-beforeAlloc)
	}
}

// 运行所有测试
func TestParse(t *testing.T) {
	t.Run("TestParseNull", TestParseNull)
	t.Run("TestParseTrue", TestParseTrue)
	t.Run("TestParseFalse", TestParseFalse)
	t.Run("TestParseNumber", TestParseNumber)
	t.Run("TestParseExpectValue", TestParseExpectValue)
	t.Run("TestParseInvalidValue", TestParseInvalidValue)
	t.Run("TestParseNumberTooBig", TestParseNumberTooBig)
	t.Run("TestParseRootNotSingular", TestParseRootNotSingular)
	t.Run("TestParseNumberBoundaryConditions", TestParseNumberBoundaryConditions)
	t.Run("TestNumberPrecision", TestNumberPrecision)
	t.Run("TestConcurrentNumberParse", TestConcurrentNumberParse)
	t.Run("TestNumberMemoryUsage", TestNumberMemoryUsage)
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

// 示例 - 错误处理
func ExampleParse_error() {
	v := Value{}
	err := Parse(&v, "invalid")
	fmt.Println(err)
	// Output: 无效的值
}
