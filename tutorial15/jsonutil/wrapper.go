// wrapper.go - 封装leptjson包API，解决类型兼容性问题
package jsonutil

import (
	"fmt"

	"github.com/Cactusinhand/go-json-tutorial/tutorial15/leptjson"
)

// 包装Parse函数
func Parse(v *leptjson.Value, json string) error {
	code := leptjson.Parse(v, json)
	if code != 0 { // PARSE_OK
		return fmt.Errorf("JSON解析错误: %d", code)
	}
	return nil
}

// 包装ParseConcurrent函数
func ParseConcurrent(json string, options leptjson.ConcurrentParseOptions) (*leptjson.Value, error) {
	// 这里我们假设leptjson包有一个ParseConcurrent函数
	// 实际上这可能需要单独实现
	v := &leptjson.Value{}
	err := Parse(v, json)
	return v, err
}

// 包装Stringify函数
func Stringify(v *leptjson.Value) (string, error) {
	return leptjson.Stringify(v)
}

// 包装StringifyIndent函数
func StringifyIndent(v *leptjson.Value, indent string) (string, error) {
	return leptjson.StringifyIndent(v, indent)
}

// 常量定义
const (
	JSON_NULL   = leptjson.JSON_NULL
	JSON_FALSE  = leptjson.JSON_FALSE
	JSON_TRUE   = leptjson.JSON_TRUE
	JSON_NUMBER = leptjson.JSON_NUMBER
	JSON_STRING = leptjson.JSON_STRING
	JSON_ARRAY  = leptjson.JSON_ARRAY
	JSON_OBJECT = leptjson.JSON_OBJECT
)

// Copy 深度复制JSON值
func Copy(dst, src *leptjson.Value) error {
	if dst == nil || src == nil {
		return fmt.Errorf("源或目标为空")
	}

	// 根据json.go中的Value结构适配字段
	dst.Type = src.Type

	switch dst.Type {
	case leptjson.JSON_NULL:
		// 无需处理
	case leptjson.JSON_TRUE, leptjson.JSON_FALSE:
		// 布尔值通过类型表示
	case leptjson.JSON_NUMBER:
		dst.N = src.N
	case leptjson.JSON_STRING:
		dst.S = src.S
	case leptjson.JSON_ARRAY:
		// 数组处理适配模式
		if len(dst.A) > 0 {
			dst.A = nil // 清空现有数组
		}
		for _, elem := range src.A {
			newElem := &leptjson.Value{}
			if err := Copy(newElem, elem); err != nil {
				return err
			}
			dst.A = append(dst.A, newElem)
		}
	case leptjson.JSON_OBJECT:
		// 对象处理适配模式
		if len(dst.M) > 0 {
			dst.M = nil // 清空现有对象
		}
		dst.M = make([][2]string, len(src.M))
		copy(dst.M, src.M)
	}

	return nil
}

// FindObjectValue 查找对象中的键值
func FindObjectValue(v *leptjson.Value, key string) (*leptjson.Value, bool) {
	if v == nil || v.Type != leptjson.JSON_OBJECT {
		return nil, false
	}

	// 遍历对象成员
	for i := 0; i < len(v.M); i++ {
		if v.M[i][0] == key {
			// 解析字符串表示的值
			val := &leptjson.Value{}
			if err := Parse(val, v.M[i][1]); err != nil {
				return nil, false
			}
			return val, true
		}
	}

	return nil, false
}

// JSONStats 统计结构体
type JSONStats struct {
	NullCount     int     // null值数量
	BoolCount     int     // 布尔值数量
	TrueCount     int     // true值数量
	FalseCount    int     // false值数量
	NumberCount   int     // 数字数量
	NumberSum     float64 // 数字总和
	StringCount   int     // 字符串数量
	StringLength  int     // 字符串总长度
	ArrayCount    int     // 数组数量
	ArrayElements int     // 数组元素总数
	ObjectCount   int     // 对象数量
	TotalFields   int     // 对象字段总数
	TotalElements int     // 总元素数量
	MaxDepth      int     // 最大嵌套深度
}

// CalculateStats 计算JSON统计信息
func CalculateStats(v *leptjson.Value) *JSONStats {
	stats := &JSONStats{}
	calcStatsRecursive(v, stats, 0)
	return stats
}

// calcStatsRecursive 递归计算统计信息
func calcStatsRecursive(v *leptjson.Value, stats *JSONStats, depth int) int {
	if v == nil {
		return depth
	}

	// 更新最大深度
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	// 总元素计数
	stats.TotalElements++

	// 按类型计数
	switch v.Type {
	case leptjson.JSON_NULL:
		stats.NullCount++
	case leptjson.JSON_TRUE:
		stats.BoolCount++
		stats.TrueCount++
	case leptjson.JSON_FALSE:
		stats.BoolCount++
		stats.FalseCount++
	case leptjson.JSON_NUMBER:
		stats.NumberCount++
		stats.NumberSum += v.N
	case leptjson.JSON_STRING:
		stats.StringCount++
		stats.StringLength += len(v.S)
	case leptjson.JSON_ARRAY:
		stats.ArrayCount++
		if v.A != nil {
			elements := len(v.A)
			stats.ArrayElements += elements

			// 递归处理数组元素
			maxChildDepth := depth
			for i := 0; i < elements; i++ {
				childDepth := calcStatsRecursive(v.A[i], stats, depth+1)
				if childDepth > maxChildDepth {
					maxChildDepth = childDepth
				}
			}
			return maxChildDepth
		}
	case leptjson.JSON_OBJECT:
		stats.ObjectCount++
		if v.M != nil {
			fields := len(v.M)
			stats.TotalFields += fields

			// 递归处理对象字段值
			maxChildDepth := depth
			for i := 0; i < fields; i++ {
				val := &leptjson.Value{}
				if Parse(val, v.M[i][1]) == nil {
					childDepth := calcStatsRecursive(val, stats, depth+1)
					if childDepth > maxChildDepth {
						maxChildDepth = childDepth
					}
				}
			}
			return maxChildDepth
		}
	}

	return depth
}

// GetTypeString 获取类型的字符串表示
func GetTypeString(t int) string {
	switch t {
	case leptjson.JSON_NULL:
		return "null"
	case leptjson.JSON_FALSE:
		return "false"
	case leptjson.JSON_TRUE:
		return "true"
	case leptjson.JSON_NUMBER:
		return "number"
	case leptjson.JSON_STRING:
		return "string"
	case leptjson.JSON_ARRAY:
		return "array"
	case leptjson.JSON_OBJECT:
		return "object"
	default:
		return "unknown"
	}
}
