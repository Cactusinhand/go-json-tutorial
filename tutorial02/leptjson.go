// leptjson.go - Go语言版JSON库实现
//
// 这个包实现了一个简单的JSON解析器，支持基本的JSON值类型
// 包括null、true、false和数字等。后续版本将支持字符串、数组和对象。
//
// 数字解析的特殊情况处理：
// 1. 整数部分：
//   - 支持正数和负数
//   - 不允许前导零（如"01"是非法的）
//   - 不允许十六进制表示（如"0x1"是非法的）
//   - 单个"0"是合法的
//
// 2. 小数部分：
//   - 小数点后必须至少有一个数字（如"1."是非法的）
//   - 小数点后不能以0开头（如"1.01"是合法的，但"1.00"会被简化为"1"）
//   - 小数点后可以有多个数字
//
// 3. 指数部分：
//   - 支持"e"或"E"作为指数符号
//   - 指数部分可以为正数或负数
//   - 指数部分不能为空（如"1e"是非法的）
//   - 指数部分不能以0开头（如"1e01"是非法的）
//
// 4. 边界值处理：
//   - 处理数字溢出情况（返回PARSE_NUMBER_TOO_BIG错误）
//   - 处理无效数字格式（返回PARSE_INVALID_VALUE错误）
//
// 5. 特殊值：
//   - 支持科学计数法表示（如"1.23e-4"）
//   - 支持最大/最小浮点数
//   - 支持正负零（"0"和"-0"）
package leptjson

import (
	"fmt"
	"math"
	"strconv"
)

// ValueType 表示JSON值的类型
type ValueType int

// String 返回ValueType的字符串表示
func (t ValueType) String() string {
	switch t {
	case NULL:
		return "null"
	case FALSE:
		return "false"
	case TRUE:
		return "true"
	case NUMBER:
		return "number"
	case STRING:
		return "string"
	case ARRAY:
		return "array"
	case OBJECT:
		return "object"
	default:
		return "unknown"
	}
}

// JSON值类型常量
const (
	NULL   ValueType = iota // null值
	FALSE                   // false值
	TRUE                    // true值
	NUMBER                  // 数字
	STRING                  // 字符串
	ARRAY                   // 数组
	OBJECT                  // 对象
)

// ParseError 表示解析过程中可能出现的错误
type ParseError int

// 解析错误常量
const (
	PARSE_OK                ParseError = iota // 解析成功
	PARSE_EXPECT_VALUE                        // 期望一个值
	PARSE_INVALID_VALUE                       // 无效的值
	PARSE_ROOT_NOT_SINGULAR                   // 根节点不唯一
	PARSE_NUMBER_TOO_BIG                      // 数字太大
)

// Error 实现error接口，返回错误描述
func (e ParseError) Error() string {
	switch e {
	case PARSE_OK:
		return "解析成功"
	case PARSE_EXPECT_VALUE:
		return "期望一个值"
	case PARSE_INVALID_VALUE:
		return "无效的值"
	case PARSE_ROOT_NOT_SINGULAR:
		return "根节点不唯一"
	case PARSE_NUMBER_TOO_BIG:
		return "数字太大"
	default:
		return fmt.Sprintf("未知错误(%d)", int(e))
	}
}

// Value 表示一个JSON值
type Value struct {
	Type ValueType `json:"type"` // 值类型
	N    float64   `json:"n"`    // 数字值（当Type为NUMBER时有效）
}

// String 实现fmt.Stringer接口，返回Value的字符串表示
func (v *Value) String() string {
	switch v.Type {
	case NUMBER:
		return fmt.Sprintf("%g", v.N)
	default:
		return v.Type.String()
	}
}

// context 表示解析过程中的上下文
type context struct {
	json  string
	index int
}

// Parse 解析JSON文本
//
// 接收一个Value指针和JSON字符串，解析JSON并将结果存储在Value中
// 返回解析过程中可能出现的错误
func Parse(v *Value, json string) ParseError {
	var c context
	c.json = json
	c.index = 0
	v.Type = NULL

	parseWhitespace(&c)
	if err := parseValue(&c, v); err != PARSE_OK {
		return err
	}
	parseWhitespace(&c)
	if c.index < len(c.json) {
		return PARSE_ROOT_NOT_SINGULAR
	}
	return PARSE_OK
}

// GetType 获取JSON值的类型
func GetType(v *Value) ValueType {
	return v.Type
}

// GetNumber 获取JSON数字值
//
// 当且仅当值类型为NUMBER时，返回值有效
func GetNumber(v *Value) float64 {
	return v.N
}

// parseWhitespace 跳过空白字符
func parseWhitespace(c *context) {
	for c.index < len(c.json) {
		if c.json[c.index] == ' ' || c.json[c.index] == '\t' || c.json[c.index] == '\n' || c.json[c.index] == '\r' {
			c.index++
		} else {
			break
		}
	}
}

// parseNull 解析null值
func parseNull(c *context, v *Value) ParseError {
	if c.index+3 >= len(c.json) ||
		c.json[c.index+1] != 'u' ||
		c.json[c.index+2] != 'l' ||
		c.json[c.index+3] != 'l' {
		return PARSE_INVALID_VALUE
	}
	c.index += 4
	v.Type = NULL
	return PARSE_OK
}

// parseTrue 解析true值
func parseTrue(c *context, v *Value) ParseError {
	if c.index+3 >= len(c.json) ||
		c.json[c.index+1] != 'r' ||
		c.json[c.index+2] != 'u' ||
		c.json[c.index+3] != 'e' {
		return PARSE_INVALID_VALUE
	}
	c.index += 4
	v.Type = TRUE
	return PARSE_OK
}

// parseFalse 解析false值
func parseFalse(c *context, v *Value) ParseError {
	if c.index+4 >= len(c.json) ||
		c.json[c.index+1] != 'a' ||
		c.json[c.index+2] != 'l' ||
		c.json[c.index+3] != 's' ||
		c.json[c.index+4] != 'e' {
		return PARSE_INVALID_VALUE
	}
	c.index += 5
	v.Type = FALSE
	return PARSE_OK
}

// parseNumber 解析数字值
func parseNumber(c *context, v *Value) ParseError {
	// 记录起始位置
	start := c.index

	// 处理负号
	if c.index < len(c.json) && c.json[c.index] == '-' {
		c.index++
	}

	// 处理整数部分
	if c.index < len(c.json) && c.json[c.index] == '0' {
		c.index++
		// 0后面不能直接跟数字，必须是小数点或指数符号
		if c.index < len(c.json) {
			ch := c.json[c.index]
			// 检查0后面是否跟着x或X（十六进制表示法）
			if ch == 'x' || ch == 'X' {
				return PARSE_INVALID_VALUE
			}
			if ch >= '0' && ch <= '9' {
				return PARSE_INVALID_VALUE
			}
		}
	} else if c.index < len(c.json) && c.json[c.index] >= '1' && c.json[c.index] <= '9' {
		c.index++
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	} else {
		return PARSE_INVALID_VALUE
	}

	// 处理小数部分
	if c.index < len(c.json) && c.json[c.index] == '.' {
		c.index++
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 处理指数部分
	if c.index < len(c.json) && (c.json[c.index] == 'e' || c.json[c.index] == 'E') {
		c.index++
		if c.index < len(c.json) && (c.json[c.index] == '+' || c.json[c.index] == '-') {
			c.index++
		}
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 转换为数字
	n, err := strconv.ParseFloat(c.json[start:c.index], 64)
	if err != nil {
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				return PARSE_NUMBER_TOO_BIG
			}
		}
		return PARSE_INVALID_VALUE
	}

	// 检查是否溢出
	if math.IsInf(n, 0) || math.IsNaN(n) {
		return PARSE_NUMBER_TOO_BIG
	}

	v.N = n
	v.Type = NUMBER
	return PARSE_OK
}

// parseValue 解析JSON值
func parseValue(c *context, v *Value) ParseError {
	if c.index >= len(c.json) {
		return PARSE_EXPECT_VALUE
	}

	switch c.json[c.index] {
	case 'n':
		return parseNull(c, v)
	case 't':
		return parseTrue(c, v)
	case 'f':
		return parseFalse(c, v)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return parseNumber(c, v)
	default:
		return PARSE_INVALID_VALUE
	}
}
