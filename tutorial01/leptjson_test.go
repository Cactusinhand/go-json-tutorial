// leptjson_test.go - Go语言版JSON库测试
package leptjson

import (
	"fmt"
	"runtime"
	"strings"
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

// 测试解析根节点不唯一
func TestParseRootNotSingular(t *testing.T) {
	v := Value{Type: FALSE}
	if err := Parse(&v, "null x"); err != PARSE_ROOT_NOT_SINGULAR {
		t.Errorf("期望错误PARSE_ROOT_NOT_SINGULAR，但得到: %v", err)
	}
	if GetType(&v) != NULL {
		t.Errorf("期望类型为NULL，但得到: %v", GetType(&v))
	}
}

// 测试边界条件
func TestParseBoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ParseError
	}{
		{"EmptyInput", "", PARSE_EXPECT_VALUE},
		{"OnlyWhitespace", "   ", PARSE_EXPECT_VALUE},
		{"MultipleWhitespace", "\t\n\r ", PARSE_EXPECT_VALUE},
		{"InvalidWhitespace", "\x00", PARSE_INVALID_VALUE},
		{"InvalidUnicode", "\uFFFE", PARSE_INVALID_VALUE},
		{"InvalidUTF8", "\x80", PARSE_INVALID_VALUE},
		{"InvalidControlChar", "\x1F", PARSE_INVALID_VALUE},
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

// 测试大输入性能
func BenchmarkParseLargeInput(b *testing.B) {
	// 生成大量重复的null值
	largeInput := strings.Repeat("null,", 1000) + "null"
	v := Value{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(&v, largeInput)
	}
}

// 测试并发解析
func TestConcurrentParse(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			v := Value{}
			for j := 0; j < numIterations; j++ {
				Parse(&v, "null")
				if GetType(&v) != NULL {
					t.Errorf("并发解析失败: 期望类型为NULL，但得到: %v", GetType(&v))
				}
			}
		}()
	}

	wg.Wait()
}

// 测试内存使用
func TestMemoryUsage(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beforeAlloc := m.Alloc

	// 创建大量Value对象
	values := make([]Value, 1000)
	for i := range values {
		Parse(&values[i], "null")
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
	t.Run("TestParseExpectValue", TestParseExpectValue)
	t.Run("TestParseInvalidValue", TestParseInvalidValue)
	t.Run("TestParseRootNotSingular", TestParseRootNotSingular)
	t.Run("TestParseBoundaryConditions", TestParseBoundaryConditions)
	t.Run("TestConcurrentParse", TestConcurrentParse)
	t.Run("TestMemoryUsage", TestMemoryUsage)
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

// 示例 - 错误处理
func ExampleParse_error() {
	v := Value{}
	err := Parse(&v, "invalid")
	fmt.Println(err)
	// Output: 无效的值
}
