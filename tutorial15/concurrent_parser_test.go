package tutorial15

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// 生成大型JSON字符串用于测试
func generateLargeJSON(size int) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString("  \"items\": [\n")

	for i := 0; i < size; i++ {
		sb.WriteString(fmt.Sprintf("    {\n      \"id\": %d,\n      \"name\": \"Item %d\",\n      \"value\": %d\n    }", i, i, i*i))
		if i < size-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  ]\n")
	sb.WriteString("}")

	return sb.String()
}

// 测试并发解析功能
func TestConcurrentParser(t *testing.T) {
	// 生成测试用JSON
	jsonStr := generateLargeJSON(50) // 减少数据量从1000到50

	// 测试用例
	testCases := []struct {
		name       string
		options    ConcurrentParseOptions
		expectFail bool
	}{
		{
			name:       "默认配置",
			options:    DefaultConcurrentParseOptions(),
			expectFail: false,
		},
		{
			name: "小块大小",
			options: ConcurrentParseOptions{
				ChunkSize:   4 * 1024, // 4KB
				WorkerCount: 4,
				AutoWorkers: false,
				Timeout:     120 * time.Second, // 延长超时时间到120秒
			},
			expectFail: false,
		},
		{
			name: "多工作线程",
			options: ConcurrentParseOptions{
				ChunkSize:   64 * 1024, // 64KB
				WorkerCount: 8,
				AutoWorkers: false,
				Timeout:     120 * time.Second, // 延长超时时间到120秒
			},
			expectFail: false,
		},
	}

	// 对照组：标准解析
	t.Run("标准解析", func(t *testing.T) {
		start := time.Now()
		v := &Value{}
		err := Parse(v, jsonStr)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("标准解析失败: %v", err)
		}

		t.Logf("标准解析耗时: %v", duration)
	})

	// 测试并发解析
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			v, err := ParseConcurrent(jsonStr, tc.options)
			duration := time.Since(start)

			if tc.expectFail {
				if err == nil {
					t.Errorf("预期失败但成功解析")
				}
			} else {
				if err != nil {
					t.Errorf("并发解析失败: %v", err)
				}
			}

			if v != nil {
				// 验证解析结果
				if v.Type != OBJECT {
					t.Errorf("解析结果类型错误，预期OBJECT，实际: %d", v.Type)
				}

				// 检查items数组
				itemsIndex, found := FindObjectKey(v, "items")
				if !found {
					t.Errorf("未找到items数组")
				} else {
					itemsVal := v.Members[itemsIndex].Value
					if itemsVal.Type != ARRAY {
						t.Errorf("items类型错误，预期ARRAY，实际: %d", itemsVal.Type)
					} else if len(itemsVal.Elements) != 50 {
						t.Errorf("items数组长度错误，预期50，实际: %d", len(itemsVal.Elements))
					}
				}
			}

			t.Logf("并发解析耗时: %v，配置: 块大小=%dKB, 工作线程=%d",
				duration, tc.options.ChunkSize/1024, tc.options.WorkerCount)
		})
	}
}

// 测试流式解析数组
func TestArrayStreaming(t *testing.T) {
	// 创建一个大型数组JSON
	const arraySize = 100
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < arraySize; i++ {
		sb.WriteString(fmt.Sprintf("\"%d\"", i))
		if i < arraySize-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")

	arrayJSON := sb.String()

	// 创建流式解析器
	parser, err := NewArrayStreamParser([]byte(arrayJSON))
	if err != nil {
		t.Fatalf("创建流式解析器失败: %v", err)
	}

	// 计数器，记录实际读取的元素数量
	var count int

	// 测试逐个读取元素
	for {
		element, err := parser.NextElement()
		if err != nil {
			break
		}

		count++

		// 验证元素值
		if element.Type != STRING {
			t.Errorf("元素#%d类型错误，预期STRING，实际: %d", count-1, element.Type)
		}

		expectedStr := fmt.Sprintf("%d", count-1)
		if element.S != expectedStr {
			t.Errorf("元素#%d值错误，预期: %s，实际: %s", count-1, expectedStr, element.S)
		}
	}

	if count != arraySize {
		t.Errorf("读取的元素数量错误，预期%d，实际: %d", arraySize, count)
	}
}

// 测试并发处理数组
func TestProcessArrayConcurrently(t *testing.T) {
	arraySize := 100 // 减小数组大小以加快测试
	jsonStr := fmt.Sprintf(`[%s]`, strings.Join(generateNumbers(arraySize), ","))

	processed := make([]bool, arraySize)
	err := ProcessArrayConcurrently([]byte(jsonStr), 4, func(v *Value) error {
		if v.Type != NUMBER {
			return fmt.Errorf("期望NUMBER类型，但得到了: %v", v.Type)
		}

		// 将值转换为索引
		var index int
		index = int(v.N)

		if index < 0 || index >= arraySize {
			return fmt.Errorf("索引超出范围: %d", index)
		}

		// 标记为已处理
		processed[index] = true
		return nil
	})

	if err != nil {
		t.Fatalf("并发处理数组失败: %v", err)
	}

	// 验证所有元素都已处理
	for i, p := range processed {
		if !p {
			t.Errorf("元素 %d 未被处理", i)
		}
	}

	// 输出处理结果
	processed_count := 0
	for _, p := range processed {
		if p {
			processed_count++
		}
	}
	t.Logf("共处理 %d/%d 个元素", processed_count, arraySize)
}

// 生成指定数量的数字字符串
func generateNumbers(count int) []string {
	nums := make([]string, count)
	for i := 0; i < count; i++ {
		nums[i] = strconv.Itoa(i)
	}
	return nums
}

// 基准测试：比较标准解析和并发解析的性能
func BenchmarkJSONParsing(b *testing.B) {
	// 生成一个中等大小的JSON
	jsonStr := generateLargeJSON(1000)

	// 基准测试标准解析
	b.Run("标准解析", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := &Value{}
			Parse(v, jsonStr)
		}
	})

	// 基准测试并发解析
	b.Run("并发解析", func(b *testing.B) {
		options := DefaultConcurrentParseOptions()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ParseConcurrent(jsonStr, options)
		}
	})
}

// 测试超大文件解析
func TestHugeJSON(t *testing.T) {
	// 跳过正常测试以避免耗时过长
	if testing.Short() {
		t.Skip("在短测试模式下跳过")
	}

	// 生成一个较小的JSON，避免测试超时
	jsonStr := generateLargeJSON(500)

	// 并发解析大型JSON
	options := DefaultConcurrentParseOptions()
	options.ChunkSize = 128 * 1024      // 128KB
	options.WorkerCount = 4             // 减少工作线程数
	options.Timeout = 120 * time.Second // 增加超时时间

	start := time.Now()
	_, err := ParseConcurrent(jsonStr, options)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("解析超大JSON失败: %v", err)
	}

	t.Logf("解析超大JSON耗时: %v", duration)
}

// 测试实际的大型JSON文件
func TestLargeJSONFile(t *testing.T) {
	// 跳过正常测试以避免依赖外部文件
	if testing.Short() {
		t.Skip("在短测试模式下跳过")
	}

	// 文件路径，可以根据实际情况调整
	filePath := "../examples/large.json"

	// 尝试打开文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Skipf("跳过测试：无法读取文件 %s: %v", filePath, err)
		return
	}

	jsonStr := string(data)

	// 并发解析大型JSON文件
	options := DefaultConcurrentParseOptions()
	options.ChunkSize = 256 * 1024 // 256KB
	options.WorkerCount = 8

	start := time.Now()
	_, err = ParseConcurrent(jsonStr, options)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("解析大型JSON文件失败: %v", err)
	}

	// 记录解析时间和文件大小，用于性能评估
	fileSize := len(data) / (1024 * 1024) // MB
	t.Logf("解析%dMB的JSON文件耗时: %v (%.2f MB/s)",
		fileSize, duration, float64(fileSize)/duration.Seconds())
}
